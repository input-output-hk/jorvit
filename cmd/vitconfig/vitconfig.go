//$(which go) run $0 $@; exit $?

package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/input-output-hk/jorvit/internal/datastore"
	"github.com/input-output-hk/jorvit/internal/kit"
	"github.com/input-output-hk/jorvit/internal/loader"
	"github.com/input-output-hk/jorvit/internal/wallet"
	"github.com/input-output-hk/jorvit/internal/webproxy"
	"github.com/rinor/jorcli/jcli"
	"github.com/rinor/jorcli/jnode"
	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/blake2b"
)

type jcliProposal struct {
	ExternalID string `json:"external_id"`
	Options    uint8  `json:"options"`
	Action     string `json:"action"` // set it "off_chain" for now
}

type jcliVotePlan struct {
	Payload      string         `json:"payload_type"` // set it "public" for now
	VoteStart    ChainTime      `json:"vote_start"`
	VoteEnd      ChainTime      `json:"vote_end"`
	CommitteeEnd ChainTime      `json:"committee_end"`
	Proposals    []jcliProposal `json:"proposals"`
	VotePlanID   string         `json:"-"`
	Certificate  string         `json:"-"`
}

type ChainTime struct {
	Epoch  int `json:"epoch"`
	SlotID int `json:"slot_id"`
}

type ChainVotePlan struct {
	VotePlanID   string    `csv:"chain_voteplan_id"`
	VoteStart    ChainTime `csv:"chain_vote_starttime"`
	VoteEnd      ChainTime `csv:"chain_vote_endtime"`
	CommitteeEnd ChainTime `csv:"chain_committee_endtime"`
	Payload      string    `csv:"chain_voteplan_payload"`
	Certificate  string
	proposalID   []string
}

func (ct *ChainTime) String() string {
	return strconv.Itoa(ct.Epoch) + "." + strconv.Itoa(ct.SlotID)
}

func (ct *ChainTime) ToSeconds(SlotDuration, SlotsPerEpoch int) int64 {
	epochDuration := SlotDuration * SlotsPerEpoch
	return int64(ct.Epoch*epochDuration + ct.SlotID*SlotDuration)
}

var (
	votePlanProposalsMax = 254
	voteStart            = ChainTime{0, 0}
	voteEnd              = ChainTime{28, 0}
	committeeEnd         = ChainTime{40, 0}
)

var (
	leaderSK = []byte("ed25519_sk1pl4vp0grkl2puspv4c3hwhz89r68yjyzalc78pyt0pujpmk8mxkq6kpc5j")
	wallets  = wallet.SampleWallets()
)

var (
	proposals datastore.ProposalsStore
	funds     datastore.FundsStore
)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func loadProposals(file string) error {
	defer timeTrack(time.Now(), "Proposals File load")
	proposals = &datastore.Proposals{}
	return proposals.Initialize(file)
}

func loadFundInfo(file string) error {
	defer timeTrack(time.Now(), "Fund File load")
	funds = &datastore.Funds{}
	return funds.Initialize(file)
}

func main() {
	var (
		proxyAddrPort       = flag.String("proxy", "0.0.0.0:8000", "Address where REST api PROXY should listen in IP:PORT format")
		restAddrPort        = flag.String("rest", "0.0.0.0:8001", "Address where Jörmungandr REST api should listen in IP:PORT format")
		nodePort            = flag.Uint("node", 9001, "PORT where Jörmungandr node should listen")
		proposalsPath       = flag.String("proposals", "."+string(os.PathSeparator)+"assets"+string(os.PathSeparator)+"proposals.csv", "CSV full path (filename) to load PROPOSALS from")
		fundsPath           = flag.String("fund", "."+string(os.PathSeparator)+"assets"+string(os.PathSeparator)+"fund.csv", "CSV full path (filename) to load FUND info from")
		dumbGenesisDataPath = flag.String("dumbdata", "."+string(os.PathSeparator)+"assets"+string(os.PathSeparator)+"dumb_genesis_data.yaml", "YAML full path (filename) to load dumb genesis funds from")
		explorerEnabled     = flag.Bool("explorer", true, "Enable/Disable explorer")
		restCorsAllowed     = flag.String("cors", "http://127.0.0.1:8000,http://127.0.0.1:8001,http://localhost:8000,http://localhost:8801,http://0.0.0.0:8000,http://0.0.0.0:8001", "Comma separated list of CORS allowed origins")
	)

	flag.Parse()

	if *proxyAddrPort == "" || *restAddrPort == "" || *nodePort == 0 || *proposalsPath == "" || *fundsPath == "" || *dumbGenesisDataPath == "" {
		flag.Usage()
	}
	err := loadProposals(*proposalsPath)
	kit.FatalOn(err, "loadProposals")

	err = loadFundInfo(*fundsPath)
	kit.FatalOn(err, "loadFundInfo")

	var (

		// Proxy
		// proxyAddr, proxyPort = "0.0.0.0", 8000
		// proxyAddress         = proxyAddr + ":" + strconv.Itoa(proxyPort)

		proxyAddress = *proxyAddrPort

		// Rest
		// restAddr, restPort = "0.0.0.0", 8001
		// restAddress        = restAddr + ":" + strconv.Itoa(restPort)

		restAddress = *restAddrPort

		// P2P
		p2pIPver, p2pProto           = "ip4", "tcp"
		p2pListenAddr, p2pListenPort = "127.0.0.11", int(*nodePort) // 9001
		p2pListenAddress             = "/" + p2pIPver + "/" + p2pListenAddr + "/" + p2pProto + "/" + strconv.Itoa(p2pListenPort)

		// General
		consensus      = "bft" // bft or genesis_praos
		discrimination = ""    // "" (empty defaults to "production")

		// Node config log
		nodeCfgLogLevel = "info"
	)

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	kit.FatalOn(err)

	// Check for jcli binary. Local folder first (jor_bins), then PATH
	jcliBin, err := kit.FindExecutable("jcli", "jor_bins")
	kit.FatalOn(err, jcliBin)
	jcli.BinName(jcliBin)

	// Check for jörmungandr binary. Local folder first, then PATH
	jnodeBin, err := kit.FindExecutable("jormungandr", "jor_bins")
	kit.FatalOn(err, jnodeBin)
	jnode.BinName(jnodeBin)

	// get jcli version
	jcliVersion, err := jcli.VersionFull()
	kit.FatalOn(err, kit.B2S(jcliVersion))

	// get jörmungandr version
	jormungandrVersion, err := jnode.VersionFull()
	kit.FatalOn(err, kit.B2S(jormungandrVersion))

	// create a new temporary directory inside your systems temp dir
	workingDir, err := ioutil.TempDir(dir, "jnode_VIT_")
	kit.FatalOn(err, "workingDir")
	log.Printf("Working Directory: %s", workingDir)

	/* BFT LEADER */
	// TODO: build a loader once provided
	leaderPK, err := jcli.KeyToPublic(leaderSK, "", "")
	kit.FatalOn(err, kit.B2S(leaderPK))

	// Needed later on to sign
	bftSecretFile := workingDir + string(os.PathSeparator) + "bft_secret.key"
	err = ioutil.WriteFile(bftSecretFile, leaderSK, 0744)
	kit.FatalOn(err)

	/////////////////////
	//  block0 config  //
	/////////////////////

	block0cfg := jnode.NewBlock0Config()

	block0Discrimination := "production"
	if discrimination == "testing" {
		block0Discrimination = "test"
	}

	// set/change config params
	block0cfg.BlockchainConfiguration.Block0Date = time.Now().UTC().Unix() // block0Date()
	block0cfg.BlockchainConfiguration.Block0Consensus = consensus
	block0cfg.BlockchainConfiguration.Discrimination = block0Discrimination

	block0cfg.BlockchainConfiguration.SlotDuration = 1
	block0cfg.BlockchainConfiguration.SlotsPerEpoch = 21_600

	block0cfg.BlockchainConfiguration.LinearFees.Certificate = 200
	block0cfg.BlockchainConfiguration.LinearFees.Coefficient = 10
	block0cfg.BlockchainConfiguration.LinearFees.Constant = 100

	block0cfg.BlockchainConfiguration.LinearFees.PerVoteCertificateFees.CertificateVoteCast = 10_000_000
	block0cfg.BlockchainConfiguration.LinearFees.PerVoteCertificateFees.CertificateVotePlan = 100_000_000

	block0cfg.BlockchainConfiguration.FeesGoTo = "treasury"

	// Bft Leader
	err = block0cfg.AddConsensusLeader(kit.B2S(leaderPK))
	kit.FatalOn(err)

	// Committee list - TODO: build a loader once defined/provided
	// block0cfg.AddCommittee("568cb82664987cec6412230d02c8eb774e75a8514f2fc224539e0c041973795d")
	// block0cfg.AddCommittee("fdf83e0c1dbe95600c957e5ab92f807c4d98061ece092091e376cdfd2ae625a9")

	// add legacy funds
	for i := range wallets {
		wallets[i].Totals = 0
		for _, lf := range wallets[i].Funds {
			err = block0cfg.AddInitialLegacyFund(lf.Address, lf.Value)
			kit.FatalOn(err)
			wallets[i].Totals += lf.Value
		}
	}

	// Total nr of proposals
	proposalsTot := proposals.Total()

	// Calculate nr of needed voteplans since there is a limit of proposals a plan can have (255)
	// TODO: change to take in consideration also payload (we have only Public for now)
	votePlansNeeded := votePlansNeeded(proposalsTot, votePlanProposalsMax)
	var votePlans = make([]ChainVotePlan, votePlansNeeded)
	// TODO: ose only this one instead off votePlans
	var jcliVotePlans = make([]jcliVotePlan, votePlansNeeded)

	funds.First().Voteplans = make([]loader.ChainVotePlan, votePlansNeeded)

	// Generate proposals hash and associate it to a voteplan
	for i, proposal := range *proposals.All() {
		// retrieve the voteplan intenal idx based on the proposal idx we are at
		// TODO: change to take in consideration also payload (we have only Public for now)
		vpIdx := votePlanIndex(i, votePlanProposalsMax)

		// hash the proposal (TODO: decide what to hash in production)
		id := blake2b.Sum256([]byte(proposal.Proposal.ID + proposal.InternalID))
		proposal.ChainProposal.ExternalID = hex.EncodeToString(id[:])

		// add proposal hash to the respective voteplan internal container
		votePlans[vpIdx].proposalID = append(votePlans[vpIdx].proposalID, proposal.ChainProposal.ExternalID)
		/*
			// We could insert also here, but do it later when we have also VotePlanID
			proposal.ChainVotePlan.VoteStart = voteStart
			proposal.ChainVotePlan.VoteEnd = voteEnd
			proposal.ChainVotePlan.CommitteeEnd = committeeEnd
			proposal.ChainProposal.Index = uint8(len(votePlans[vpIdx].proposalID))
		*/

		// TODO: ose only this one instead off votePlans
		jcliVotePlans[vpIdx].Proposals = append(
			jcliVotePlans[vpIdx].Proposals,
			jcliProposal{
				ExternalID: proposal.ChainProposal.ExternalID,
				Options:    uint8(len(proposal.ChainProposal.VoteOptions)),
				Action:     "off_chain",
			},
		)

	}

	// Generate voteplan certificates and id
	for i := range votePlans {
		votePlans[i].VoteStart = voteStart
		votePlans[i].VoteEnd = voteEnd
		votePlans[i].CommitteeEnd = committeeEnd
		votePlans[i].Payload = "public"

		// TODO: ose only this one instead off votePlans
		jcliVotePlans[i].VoteStart = voteStart
		jcliVotePlans[i].VoteEnd = voteEnd
		jcliVotePlans[i].CommitteeEnd = committeeEnd
		jcliVotePlans[i].Payload = "public"

		stdinConfig, err := json.Marshal(jcliVotePlans[i])
		kit.FatalOn(err, "json.Marshal VotePlan Config")

		cert, err := jcli.CertificateNewVotePlan(stdinConfig, "", "")
		kit.FatalOn(err, "CertificateNewVotePlan")

		id, err := jcli.CertificateGetVotePlanID(cert, "", "")
		kit.FatalOn(err, "CertificateGetVotePlanID:", kit.B2S(id))

		cert, err = jcli.CertificateSign(cert, []string{bftSecretFile}, "", "")
		kit.FatalOn(err, "CertificateSign:", kit.B2S(cert))

		votePlans[i].Certificate = kit.B2S(cert)
		votePlans[i].VotePlanID = kit.B2S(id)

		// Vote Plans add certificate to block0
		err = block0cfg.AddInitialCertificate(votePlans[i].Certificate)
		kit.FatalOn(err, "AddInitialCertificate")

		voteStartUnix := votePlans[i].VoteStart.ToSeconds(
			int(block0cfg.BlockchainConfiguration.SlotDuration),
			int(block0cfg.BlockchainConfiguration.SlotsPerEpoch),
		) + block0cfg.BlockchainConfiguration.Block0Date

		voteEndUnix := votePlans[i].VoteEnd.ToSeconds(
			int(block0cfg.BlockchainConfiguration.SlotDuration),
			int(block0cfg.BlockchainConfiguration.SlotsPerEpoch),
		) + block0cfg.BlockchainConfiguration.Block0Date

		committeeEndUnix := votePlans[i].CommitteeEnd.ToSeconds(
			int(block0cfg.BlockchainConfiguration.SlotDuration),
			int(block0cfg.BlockchainConfiguration.SlotsPerEpoch),
		) + block0cfg.BlockchainConfiguration.Block0Date

		for pi, propHash := range votePlans[i].proposalID {
			// TODO: fix this search
			proposal := datastore.FilterSingle(proposals.All(), func(v *loader.ProposalData) bool {
				return v.ChainProposal.ExternalID == propHash
			})

			proposal.ChainVotePlan.VotePlanID = votePlans[i].VotePlanID
			proposal.ChainProposal.Index = uint8(pi)

			proposal.ChainVotePlan.VoteStart = time.Unix(voteStartUnix, 0).String()       // strconv.FormatInt(voteStartUnix, 10)
			proposal.ChainVotePlan.VoteEnd = time.Unix(voteEndUnix, 0).String()           // strconv.FormatInt(voteEndUnix, 10)
			proposal.ChainVotePlan.CommitteeEnd = time.Unix(committeeEndUnix, 0).String() // strconv.FormatInt(committeeEndUnix, 10)

		}

		funds.First().Voteplans[i].VotePlanID = votePlans[i].VotePlanID
		funds.First().Voteplans[i].VoteStart = time.Unix(voteStartUnix, 0).String()
		funds.First().Voteplans[i].VoteEnd = time.Unix(voteEndUnix, 0).String()
		funds.First().Voteplans[i].CommitteeEnd = time.Unix(committeeEndUnix, 0).String()
		funds.First().Voteplans[i].Payload = votePlans[i].Payload
	}

	block0Yaml, err := block0cfg.ToYaml()
	kit.FatalOn(err)

	bulkDumbData, _ := ioutil.ReadFile(*dumbGenesisDataPath)
	// kit.FatalOn(err)
	// ignore any error since that data is just to increase block0 size for testing

	if len(bulkDumbData) > 0 {
		block0Yaml = append(block0Yaml, bulkDumbData...)
	}

	// need this file for starting the node (--genesis-block)
	block0BinFile := workingDir + string(os.PathSeparator) + "VIT-block0.bin"

	// keep also the text block0 config
	block0TxtFile := workingDir + string(os.PathSeparator) + "VIT-block0.yaml"

	// block0BinFile will be created by jcli
	block0Bin, err := jcli.GenesisEncode(block0Yaml, "", block0BinFile)
	kit.FatalOn(err, kit.B2S(block0Bin))

	block0Hash, err := jcli.GenesisHash(block0Bin, "")
	kit.FatalOn(err, kit.B2S(block0Hash))

	// block0TxtFile will be created by jcli - it fails for now due to the voteplan cert hack
	block0Txt, err := jcli.GenesisDecode(block0Bin, "", block0TxtFile)
	_ = block0Txt
	// kit.FatalOn(err, kit.B2S(block0Txt))

	// TODO: remove once proper voteplan cert inside genesis is implemented
	if err != nil {
		err = ioutil.WriteFile(block0TxtFile, block0Yaml, 0755)
		kit.FatalOn(err)
	}

	//////////////////////
	//  secrets config  //
	//////////////////////

	secretCfg := jnode.NewSecretConfig()

	secretCfg.Bft.SigningKey = kit.B2S(leaderSK)

	secretCfgYaml, err := secretCfg.ToYaml()
	kit.FatalOn(err)

	// need this file for starting the node (--secret)
	secretCfgFile := workingDir + string(os.PathSeparator) + "bft-secret.yaml"
	err = ioutil.WriteFile(secretCfgFile, secretCfgYaml, 0644)
	kit.FatalOn(err)

	///////////////////
	//  node config  //
	///////////////////

	nodeCfg := jnode.NewNodeConfig()

	nodeCfg.Storage = "storage"
	nodeCfg.SkipBootstrap = true
	nodeCfg.Rest.Listen = restAddress
	nodeCfg.P2P.ListenAddress = p2pListenAddress
	nodeCfg.P2P.AllowPrivateAddresses = true
	nodeCfg.Log.Level = nodeCfgLogLevel
	nodeCfg.Rest.Cors.AllowedOrigins = strings.Split(*restCorsAllowed, ",")
	nodeCfg.Rest.Cors.MaxAgeSecs = 0

	nodeCfg.Explorer.Enabled = *explorerEnabled

	nodeCfgYaml, err := nodeCfg.ToYaml()
	kit.FatalOn(err)

	// need this file for starting the node (--config)
	nodeCfgFile := workingDir + string(os.PathSeparator) + "node-config.yaml"
	err = ioutil.WriteFile(nodeCfgFile, nodeCfgYaml, 0644)
	kit.FatalOn(err)

	//////////////////////
	// running the node //
	//////////////////////

	node := jnode.NewJnode()

	node.WorkingDir = workingDir
	node.GenesisBlock = block0BinFile
	node.ConfigFile = nodeCfgFile

	node.AddSecretFile(secretCfgFile)

	node.Stdout, err = os.Create(filepath.Join(workingDir, "stdout.log"))
	kit.FatalOn(err)
	node.Stderr, err = os.Create(filepath.Join(workingDir, "stderr.log"))
	kit.FatalOn(err)

	// Run the node (Start + Wait)
	err = os.Setenv("RUST_BACKTRACE", "full")
	kit.FatalOn(err, "Failed to set env (RUST_BACKTRACE=full)")

	err = node.Run()
	if err != nil {
		log.Fatalf("node.Run FAILED: %v", err)
	}

	go func() {
		err := webproxy.Run(proposals, funds, &block0Bin, proxyAddress, "http://"+restAddress)
		if err != nil {
			kit.FatalOn(err, "Proxy Run")
		}
	}()

	log.Println()
	log.Printf("OS: %s, ARCH: %s", runtime.GOOS, runtime.GOARCH)
	log.Println()
	log.Printf("jcli: %s", jcliBin)
	log.Printf("ver : %s", jcliVersion)
	log.Println()
	log.Printf("node: %s", jnodeBin)
	log.Printf("ver : %s", jormungandrVersion)
	log.Println()
	log.Printf("VIT - BFT Genesis Hash: %s\n", kit.B2S(block0Hash))
	log.Println()
	log.Printf("VIT - BFT Genesis: %s - %d", "COMMITTEE", len(block0cfg.BlockchainConfiguration.Committees))
	log.Printf("VIT - BFT Genesis: %s - %d", "VOTEPLANS", len(votePlans))
	log.Printf("VIT - BFT Genesis: %s - %d", "PROPOSALS", proposals.Total())
	log.Println()
	log.Printf("VIT - BFT Genesis: %s", "Wallets available for recovery")

	qrPrint(wallets)

	log.Println()
	log.Printf("JÖRMUNGANDR listening at: %s", p2pListenAddress)
	log.Printf("JÖRMUNGANDR Rest API available at: http://%s/api", restAddress)
	log.Println()
	log.Printf("APP - PROXY Rest API available at: http://%s/api", proxyAddress)
	log.Println()
	log.Println("VIT - BFT Genesis Node - Running...")
	node.Wait()                                     // Wait for the node to stop.
	log.Println("...VIT - BFT Genesis Node - Done") // All done. Node has stopped.
}

// Print Wallet data and QR
func qrPrint(w []wallet.Wallet) {
	for i := range wallets {
		q, err := qrcode.New(w[i].Mnemonics, qrcode.Medium)
		kit.FatalOn(err)

		fmt.Printf("\n%s\n%s\n", w[i], q.ToSmallString(false))
	}
}

func votePlanIndex(i int, max int) int {
	return i / max
}

func votePlansNeeded(proposalsTot int, max int) int {
	votePlansNeeded, more := proposalsTot/max, proposalsTot%max
	if more > 0 {
		votePlansNeeded = votePlansNeeded + 1
	}
	return votePlansNeeded
}
