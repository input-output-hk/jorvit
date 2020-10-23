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
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/input-output-hk/jorvit/internal/datastore"
	"github.com/input-output-hk/jorvit/internal/kit"
	"github.com/input-output-hk/jorvit/internal/loader"
	"github.com/input-output-hk/jorvit/internal/webproxy"
	"github.com/rinor/jorcli/jcli"
	"github.com/rinor/jorcli/jnode"
	"github.com/rinor/vitcli/vcli"
	"github.com/rinor/vitcli/vstation"
	"golang.org/x/crypto/blake2b"
)

var (
	// Version and build info that can be set on build
	Version    = "dev"
	CommitHash = "none"
	BuildDate  = "unknown"

	// memory sprocessing stores
	proposals datastore.ProposalsStore
	funds     datastore.FundsStore
)

type bftLeader struct {
	sk      string
	pk      string
	acc     string
	skFile  string
	cfgFile string
}

type jcliProposal struct {
	ExternalID string `json:"external_id"`
	Options    uint8  `json:"options"`
	Action     string `json:"action"`
}

type jcliVotePlan struct {
	Payload                   string         `json:"payload_type"`
	VoteStart                 ChainTime      `json:"vote_start"`
	VoteEnd                   ChainTime      `json:"vote_end"`
	CommitteeEnd              ChainTime      `json:"committee_end"`
	Proposals                 []jcliProposal `json:"proposals"`
	CommitteeMemberPublicKeys []string       `json:"committee_member_public_keys"` // privacy encyption keys
	VotePlanID                string         `json:"-"`
	Certificate               string         `json:"-"`
}

type ChainTime struct {
	Epoch  int64 `json:"epoch"`
	SlotID int64 `json:"slot_id"`
}

func (ct ChainTime) String() string {
	return strconv.FormatInt(ct.Epoch, 10) + "." + strconv.FormatInt(ct.SlotID, 10)
}

func ToChainTime(block0Time int64, SlotDuration uint8, SlotsPerEpoch uint32, dataTime int64) ChainTime {
	slotsTotal := (dataTime - block0Time) / int64(SlotDuration)
	epoch := slotsTotal / int64(SlotsPerEpoch)
	slot := slotsTotal % int64(SlotsPerEpoch)

	return ChainTime{
		Epoch:  epoch,
		SlotID: slot,
	}
}

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

func votePlansNeeded(proposalsTot int, max int) int {
	votePlansNeeded, more := proposalsTot/max, proposalsTot%max
	if more > 0 {
		votePlansNeeded = votePlansNeeded + 1
	}
	return votePlansNeeded
}

type sliceFlag []string

func (sf *sliceFlag) String() string {
	return strings.Join(*sf, ",")
}

func (sf *sliceFlag) Set(val string) error {
	*sf = append(*sf, val)
	return nil
}

func main() {
	var (
		err error

		// BFT Leaders
		bftLeadersSecretKeys sliceFlag
		bftLeadersPublicKeys sliceFlag

		// Committee auth + privacy members
		committeeAuthPublicKeys    sliceFlag
		committeePrivacyPublicKeys sliceFlag

		// max proposals included within one voteplan - hard limit
		votePlanProposalsMax uint

		// Lovelace amount for Bft Leader and Committee Auth members
		bftLeaderFund uint64
		committeeFund uint64
	)

	// node settings
	proxyAddrPort := flag.String("proxy", "0.0.0.0:8000", "Address where REST api PROXY should listen in IP:PORT format")
	restAddrPort := flag.String("rest", "0.0.0.0:8001", "Address where Jörmungandr REST api should listen in IP:PORT format")
	nodeAddrPort := flag.String("node", "127.0.0.1:9001", "Address where Jörmungandr node should listen in IP:PORT format")
	explorerEnabled := flag.Bool("explorer", false, "Enable/Disable explorer")
	restCorsAllowed := flag.String("cors", "https://api.vit.iohk.io,https://127.0.0.1,http://127.0.0.1,http://127.0.0.1:8000,http://127.0.0.1:8001,https://localhost,http://localhost,http://localhost:8000,http://localhost:8001,http://0.0.0.0:8000,http://0.0.0.0:8001", "Comma separated list of CORS allowed origins")
	skipBootstrap := flag.Bool("skip-bootstrap", true, "Skip node bootstrap, in case of first/single genesis leader (default true)")
	nodeLogLevel := flag.String("node-log-level", "warn", "Jörmungandr node log level, [off, critical, error, warn, info, debug, trace]")
	// extra node
	allowNodeRestart := flag.Bool("allow-node-restart", true, "Allows to stop the node started from the service and restart it manually")
	shutdownNode := flag.Bool("shutdown-node", true, "When exiting try node shutdown in case the node was restarted manually")
	startNode := flag.Bool("start-node", false, "Start jörmungandr node. When false only config will be generated")

	// vit service station settings
	vitAddrPort := flag.String("vit-station", "0.0.0.0:3030", "Address where vit-servicing-station-server should listen in IP:PORT format")
	vitLogLevel := flag.String("vit-log-level", "warn", "vit-servicing-station-server log level, [off, critical, error, warn, info, debug, trace]")
	// extra vit
	startVit := flag.Bool("start-vit", false, "Start vit-servicing-station-server. When false only config will be generated")

	// external proposal data
	proposalsPath := flag.String("proposals", "."+string(os.PathSeparator)+"assets"+string(os.PathSeparator)+"proposals.csv", "CSV full path (filename) to load PROPOSALS from")
	fundsPath := flag.String("fund", "."+string(os.PathSeparator)+"assets"+string(os.PathSeparator)+"fund.csv", "CSV full path (filename) to load FUND info from")
	genesisExtraDataPath := flag.String("genesis-extra-data", "."+string(os.PathSeparator)+"assets"+string(os.PathSeparator)+"extra_genesis_data.yaml", "YAML full path (filename) to load extra genesis funds from")

	// vote and committee related timing
	voteStartFlag := flag.String("vote-start", "", "Vote start time in '2006-01-02T15:04:05Z07:00' RFC3339 format. If not set 'genesis-time' will be used")
	voteEndFlag := flag.String("vote-end", "", "Vote end time in '2006-01-02T15:04:05Z07:00' RFC3339 format. If not set 'vote-duration' will be used")
	committeeEndFlag := flag.String("committee-end", "", "Committee end time in '2006-01-02T15:04:05Z07:00' RFC3339 format. If not set 'committee-duration' will be used")

	voteDurationFlag := flag.String("vote-duration", "144h", "Voting period duration. Ignored if 'vote-end' is set")
	committeeDurationFlag := flag.String("committee-duration", "24h", "Committee period duration. Ignored if 'committee-end' is set")

	flag.UintVar(&votePlanProposalsMax, "voteplan-proposals-max", 255, "Max number of proposals per voteplan [1-256]")

	block0Voteplans := flag.Bool("block0-voteplan", false, "Enable/Disable inclusion of proposals/voteplans signed certificate on block0")

	// genesis (block0) settings
	genesisTimeFlag := flag.String("genesis-time", "", "Genesis time in '2006-01-02T15:04:05Z07:00' RFC3339 format (default \"Now()\")")
	slotDurFlag := flag.String("slot-duration", "20s", "Slot period duration. 1s-255s")
	epochDurFlag := flag.String("epoch-duration", "24h", "Epoch period duration")

	// BFT Leaders - also promoted to Global Committee members
	bftLeaderTot := flag.Uint("bft-leader-min", 1, "Minimun number of BFT Leaders. NEW SK/PK key pair(s) will be autogenerated if > \"bft-leader-secret-key\" + \"bft-leader-public-key\". min: 1")
	flag.Var(&bftLeadersSecretKeys, "bft-leader-secret-key", "File containing SK (secret key) to be used as BFT leader")
	flag.Var(&bftLeadersPublicKeys, "bft-leader-public-key", "PK (public key) to be used as BFT leader. No config file will be generated for this (since don't have the SK). ex: ed25519_pk15f7p4nzektlrj6muvvmn0hatzekg7yf0qjx54pg72qq2zgjjzdzqwhm8rz")

	// Global Committee auth members public keys
	flag.Var(&committeeAuthPublicKeys, "committee-auth-public-key", "Global committee member public key. ex: ed25519_pk15f7p4nzektlrj6muvvmn0hatzekg7yf0qjx54pg72qq2zgjjzdzqwhm8rz")
	// Voteplan Committee privacy members public keys
	flag.Var(&committeePrivacyPublicKeys, "committee-privacy-public-key", "Privacy committee member public key used to build encyption key, hex encoded")

	// (bug) - 0 fees is ignored from the jorcli lib (needs fixing)
	// fees
	feesCertificate := flag.Uint64("fees-certificate", 0, "Default certificate fee (lovelace)")
	feesCoefficient := flag.Uint64("fees-coefficient", 0, "Coefficient fee")
	feesConstant := flag.Uint64("fees-constant", 0, "Constant fee (lovelace)")
	feesCertificatePoolRegistration := flag.Uint64("fees-certificate-pool-registration", 0, "Pool registration certificate fee (lovelace)")
	feesCertificateStakeDelegation := flag.Uint64("fees-certificate-stake-delegation", 0, "Stake delegation certificate fee (lovelace)")
	feesCertificateVotePlan := flag.Uint64("fees-certificate-vote-plan", 0, "VotePlan certificate fee (lovelace)")
	feesCertificateVoteCast := flag.Uint64("fees-certificate-vote-cast", 0, "VoteCast certificate fee (lovelace)")
	feesGoTo := flag.String("fees-go-to", "rewards", "Where to send the collected fees, rewards or treasury")

	// in memory service only
	dateTimeFormat := flag.String("time-format", time.RFC3339, "Date/Time format that will be used for display (go lang format), ex: \"2006-01-02 15:04:05 -0700 MST\"")

	// version info
	version := flag.Bool("version", false, "Print current app version and build info")

	// fund each btf leader and/or committee auth account address
	flag.Uint64Var(&bftLeaderFund, "bft-leader-fund", 1_000_000, "Lovelace amount to fund bft leader account")
	flag.Uint64Var(&committeeFund, "committee-auth-fund", 1_000_000, "Lovelace amount to fund committee auth account")

	flag.Parse()

	if *version {
		fmt.Printf("Version - %s\n", Version)
		fmt.Printf("Commit  - %s\n", CommitHash)
		fmt.Printf("Date    - %s\n", BuildDate)
		os.Exit(0)
	}

	if *nodeLogLevel == "" {
		*nodeLogLevel = "warn"
	}
	if *vitLogLevel == "" {
		*vitLogLevel = "warn"
	}

	// check if file exist - duplicate data check is performed later on
	for i := range bftLeadersSecretKeys {
		_, err = os.Stat(bftLeadersSecretKeys[i])
		kit.FatalOn(err)
	}

	// set new value for bftLeaderTot if provided inputs are more
	if len(bftLeadersSecretKeys) > 0 || len(bftLeadersPublicKeys) > 0 {
		inputLeaders := uint(len(bftLeadersSecretKeys) + len(bftLeadersPublicKeys))
		if inputLeaders > *bftLeaderTot {
			*bftLeaderTot = inputLeaders
		}
	}

	if *dateTimeFormat == "" {
		*dateTimeFormat = time.RFC3339
	}

	if *genesisTimeFlag == "" {
		*genesisTimeFlag = time.Now().UTC().Format(time.RFC3339)
	}
	genesisTime, err := time.Parse(time.RFC3339, *genesisTimeFlag)
	kit.FatalOn(err, "genesisTime")

	slotDur, err := time.ParseDuration(*slotDurFlag)
	kit.FatalOn(err, "slotDuration")
	switch {
	case slotDur == 0:
		log.Fatalf("[%s] - cannot be 0", "slotDuration")
	case slotDur%time.Second > 0:
		log.Fatalf("[%s] - smallest unit is [1s]", "slotDuration")
	case slotDur > 255*time.Second:
		log.Fatalf("[%s] - max allowed value is [255s]", "slotDuration")
	}

	epochDur, err := time.ParseDuration(*epochDurFlag)
	kit.FatalOn(err, "epochDuration")
	switch {
	case epochDur == 0:
		log.Fatalf("[%s] - cannot be 0", "epochDuration")
	case epochDur%time.Second > 0:
		log.Fatalf("[%s] - smallest unit is [1s]", "epochDuration")
	case epochDur%slotDur > 0:
		log.Fatalf("[%s: %s] - should be multiple of [%s: %s].", "epochDuration", epochDur.String(), "SlotDuration", slotDur.String())
	}

	voteDur, err := time.ParseDuration(*voteDurationFlag)
	kit.FatalOn(err, "voteDuration")
	switch {
	case voteDur == 0:
		log.Fatalf("[%s] - cannot be 0", "voteDuration")
	case voteDur%time.Second > 0:
		log.Fatalf("[%s] - smallest unit is [1s]", "voteDuration")
	case voteDur%slotDur > 0:
		log.Fatalf("[%s: %s] - should be multiple of [%s: %s].", "voteDuration", voteDur.String(), "SlotDuration", slotDur.String())
	}

	committeeDur, err := time.ParseDuration(*committeeDurationFlag)
	kit.FatalOn(err, "committeeDuration")
	switch {
	case committeeDur == 0:
		log.Fatalf("[%s] - cannot be 0", "committeeDuration")
	case committeeDur%time.Second > 0:
		log.Fatalf("[%s] - smallest unit is [1s]", "committeeDuration")
	case committeeDur%slotDur > 0:
		log.Fatalf("[%s: %s] - should be multiple of [%s: %s].", "committeeDuration", committeeDur.String(), "SlotDuration", slotDur.String())
	}

	if *voteStartFlag == "" {
		*voteStartFlag = *genesisTimeFlag
	}
	voteStartTime, err := time.Parse(time.RFC3339, *voteStartFlag)
	kit.FatalOn(err, "voteStartTime")
	switch {
	case voteStartTime.Sub(genesisTime) < 0:
		log.Fatalf("%s: [%s] can't be smaller than %s: [%s]", "voteStart", *voteStartFlag, "genesisTime", *genesisTimeFlag)
	case voteStartTime.Sub(genesisTime)%slotDur != 0:
		log.Fatalf("%s: [%s] needs to have %s: [%s] steps from %s: [%s]", "voteStart", *voteStartFlag, "SlotDuration", slotDur.String(), "genesisTime", *genesisTimeFlag)
	}

	if *voteEndFlag == "" {
		*voteEndFlag = voteStartTime.Add(voteDur).Format(time.RFC3339)
	}
	voteEndTime, err := time.Parse(time.RFC3339, *voteEndFlag)
	kit.FatalOn(err, "voteEndTime")
	switch {
	case voteEndTime.Sub(voteStartTime) < 0:
		log.Fatalf("%s: [%s] can't be smaller than %s: [%s]", "voteEnd", *voteEndFlag, "voteStart", *voteStartFlag)
	case voteEndTime.Sub(genesisTime)%slotDur != 0:
		log.Fatalf("%s: [%s] needs to have %s: [%s] steps from %s: [%s]", "voteEnd", *voteEndFlag, "SlotDuration", slotDur.String(), "genesisTime", *genesisTimeFlag)
	}

	if *committeeEndFlag == "" {
		*committeeEndFlag = voteEndTime.Add(committeeDur).Format(time.RFC3339)
	}
	committeeEndTime, err := time.Parse(time.RFC3339, *committeeEndFlag)
	kit.FatalOn(err, "committeeEndTime")
	switch {
	case committeeEndTime.Sub(voteEndTime) < 0:
		log.Fatalf("%s: [%s] can't be smaller than %s: [%s]", "committeeEnd", *committeeEndFlag, "voteEnd", *voteEndFlag)
	case committeeEndTime.Sub(genesisTime)%slotDur != 0:
		log.Fatalf("%s: [%s] needs to have %s: [%s] steps from %s: [%s]", "committeeEnd", *committeeEndFlag, "SlotDuration", slotDur.String(), "genesisTime", *genesisTimeFlag)
	}

	voteStart := ToChainTime(
		genesisTime.Unix(),
		uint8(slotDur.Seconds()),
		uint32(epochDur/slotDur),
		voteStartTime.Unix(),
	)

	voteEnd := ToChainTime(
		genesisTime.Unix(),
		uint8(slotDur.Seconds()),
		uint32(epochDur/slotDur),
		voteEndTime.Unix(),
	)

	committeeEnd := ToChainTime(
		genesisTime.Unix(),
		uint8(slotDur.Seconds()),
		uint32(epochDur/slotDur),
		committeeEndTime.Unix(),
	)

	switch {
	case *proposalsPath == "":
		log.Fatalf("[%s] - not provided", "proposals file")
	case *fundsPath == "":
		log.Fatalf("[%s] - not provided", "fund file")

	case *bftLeaderTot == 0:
		log.Fatalf("[%s: %d] - wrong value", "bftLeaderTot", *bftLeaderTot)

	case *proxyAddrPort == "":
		log.Fatalf("[%s] - not set", "proxy")
	case *restAddrPort == "":
		log.Fatalf("[%s] - not set", "rest")
	case *nodeAddrPort == "":
		log.Fatalf("[%s] - not set", "node")

	case *vitAddrPort == "":
		log.Fatalf("[%s] - not set", "vit-station")

	case votePlanProposalsMax < 1:
		log.Fatalf("[%s: %d] - wrong value, expected > 0", "votePlanProposalsMax", votePlanProposalsMax)
	}

	nodeListen := strings.Split(*nodeAddrPort, ":")
	nodeAddr := nodeListen[0]
	nodePort, err := strconv.Atoi(nodeListen[1])
	kit.FatalOn(err, "nodePort")

	err = loadProposals(*proposalsPath)
	kit.FatalOn(err, "loadProposals")

	err = loadFundInfo(*fundsPath)
	kit.FatalOn(err, "loadFundInfo")

	var (
		// Proxy
		proxyAddress = *proxyAddrPort

		// Rest
		restAddress = *restAddrPort

		// P2P
		p2pIPver, p2pProto           = "ip4", "tcp"
		p2pListenAddr, p2pListenPort = nodeAddr, nodePort
		p2pListenAddress             = "/" + p2pIPver + "/" + p2pListenAddr + "/" + p2pProto + "/" + strconv.Itoa(p2pListenPort)

		// General
		consensus      = "bft" // bft or genesis_praos
		discrimination = ""    // "" (empty defaults to "production")

		// Directories within main working dir "jnode_VIT_xxxxx"
		votePlanDir   = "vote_plans"
		vitStationDir = "vit_station"
	)

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	kit.FatalOn(err)

	// Check for jcli binary. Local folder first (jor_bins), then PATH
	jcliBin, err := kit.FindExecutable("jcli", "jor_bins")
	kit.FatalOn(err, jcliBin)
	jcli.BinName(jcliBin)

	// get jcli version
	jcliVersion, err := jcli.VersionFull()
	kit.FatalOn(err, kit.B2S(jcliVersion))

	/* Working directories */

	// create a new working directory
	workingDir, err := ioutil.TempDir(dir, "jnode_VIT_")
	kit.FatalOn(err, "workingDir")
	log.Printf("Working Directory: %s", workingDir)

	// directory to dump the voteplan(s) config(s) and certificate(s)
	votePlanDir = filepath.Join(workingDir, votePlanDir)
	err = os.Mkdir(votePlanDir, 0755)
	kit.FatalOn(err, "votePlanDir")

	// directory to dump the vit servicing station configs
	vitStationDir = filepath.Join(workingDir, vitStationDir)
	err = os.Mkdir(vitStationDir, 0755)
	kit.FatalOn(err, "vitStationDir")

	/* BFT LEADER(s) */

	leaders := make([]bftLeader, 0, *bftLeaderTot)
	leadersPubKey := make(map[string]bool, *bftLeaderTot)

	var (
		bftFileIdx int
		bftPkIdx   int
	)
	for i := 0; uint(i) < *bftLeaderTot; i++ {
		var (
			leaderSK      []byte
			leaderPK      []byte
			bftSecretFile string
		)

		switch {

		case len(bftLeadersSecretKeys)-bftFileIdx > 0:
			leaderSK, err = ioutil.ReadFile(bftLeadersSecretKeys[bftFileIdx])
			kit.FatalOn(err, kit.B2S(leaderSK))
			leaderPK, err = jcli.KeyToPublic(leaderSK, "", "")
			kit.FatalOn(err, kit.B2S(leaderPK))
			bftFileIdx++

		case len(bftLeadersPublicKeys)-bftPkIdx > 0:
			leaderPK = []byte(bftLeadersPublicKeys[bftPkIdx])
			bftPkIdx++

		default:
			leaderSK, err = jcli.KeyGenerate("", "Ed25519", "")
			kit.FatalOn(err, kit.B2S(leaderSK))
			leaderPK, err = jcli.KeyToPublic(leaderSK, "", "")
			kit.FatalOn(err, kit.B2S(leaderPK))
		}

		if leadersPubKey[kit.B2S(leaderPK)] {
			i-- // needed to reach bftLeaderTot, won't go below 0
			log.Printf("***** Duplicate BFT Leader skip: %s *****", kit.B2S(leaderPK))
			log.Println()
			continue
		}
		leadersPubKey[kit.B2S(leaderPK)] = true

		leaderACC, err := jcli.AddressAccount(kit.B2S(leaderPK), "", "")
		kit.FatalOn(err, kit.B2S(leaderACC))

		if len(leaderSK) > 0 {
			// Needed later on to sign
			bftSecretFile = filepath.Join(workingDir, strconv.Itoa(i)+"_bft_secret.key")
			err = ioutil.WriteFile(bftSecretFile, leaderSK, 0744)
			kit.FatalOn(err)
		}

		leaders = append(leaders, bftLeader{
			sk:     kit.B2S(leaderSK),
			pk:     kit.B2S(leaderPK),
			acc:    kit.B2S(leaderACC),
			skFile: bftSecretFile,
		})
	}

	/////////////////////
	//  block0 config  //
	/////////////////////

	block0cfg := jnode.NewBlock0Config()

	block0Discrimination := "production"
	if discrimination == "testing" {
		block0Discrimination = "test"
	}

	// set/change config params
	block0cfg.BlockchainConfiguration.Block0Date = genesisTime.Unix()
	block0cfg.BlockchainConfiguration.Block0Consensus = consensus
	block0cfg.BlockchainConfiguration.Discrimination = block0Discrimination

	block0cfg.BlockchainConfiguration.SlotDuration = uint8(slotDur.Seconds())
	block0cfg.BlockchainConfiguration.SlotsPerEpoch = uint32(epochDur / slotDur)

	block0cfg.BlockchainConfiguration.LinearFees.Certificate = *feesCertificate
	block0cfg.BlockchainConfiguration.LinearFees.Coefficient = *feesCoefficient
	block0cfg.BlockchainConfiguration.LinearFees.Constant = *feesConstant

	block0cfg.BlockchainConfiguration.LinearFees.PerCertificateFees.CertificatePoolRegistration = *feesCertificatePoolRegistration
	block0cfg.BlockchainConfiguration.LinearFees.PerCertificateFees.CertificateStakeDelegation = *feesCertificateStakeDelegation

	block0cfg.BlockchainConfiguration.LinearFees.PerVoteCertificateFees.CertificateVoteCast = *feesCertificateVoteCast
	block0cfg.BlockchainConfiguration.LinearFees.PerVoteCertificateFees.CertificateVotePlan = *feesCertificateVotePlan

	block0cfg.BlockchainConfiguration.FeesGoTo = *feesGoTo

	// Bft Leader
	for i := range leaders {
		err = block0cfg.AddConsensusLeader(leaders[i].pk)
		kit.FatalOn(err)

		// add bft leader(s) accounts to block0 (with bftLeaderFund value)
		if bftLeaderFund > 0 {
			err = block0cfg.AddInitialFund(leaders[i].acc, bftLeaderFund)
			kit.FatalOn(err)
		}

	}

	// Global Committee Members list
	if len(committeeAuthPublicKeys) > 0 {
		committeePubAuth := make(map[string]bool, len(committeeAuthPublicKeys))
		for i := range committeeAuthPublicKeys {
			// Check if committee pk is on bft leaders
			if leadersPubKey[committeeAuthPublicKeys[i]] {
				log.Printf("***** Duplicate Committee member on BFT Leader, skip: %s *****", committeeAuthPublicKeys[i])
				log.Println()
				continue
			}

			if committeePubAuth[committeeAuthPublicKeys[i]] {
				log.Printf("***** Duplicate Committee member, skip: %s *****", committeeAuthPublicKeys[i])
				log.Println()
				continue
			}
			committeePubAuth[committeeAuthPublicKeys[i]] = true

			pk, err := jcli.KeyToBytes([]byte(committeeAuthPublicKeys[i]), "", "")
			kit.FatalOn(err, kit.B2S(pk))
			block0cfg.AddCommittee(kit.B2S(pk))

			// add committee accounts to block0 (with committeeFund value)
			if committeeFund > 0 {
				comACC, err := jcli.AddressAccount(committeeAuthPublicKeys[i], "", "")
				kit.FatalOn(err, kit.B2S(comACC))
				err = block0cfg.AddInitialFund(kit.B2S(comACC), committeeFund)
				kit.FatalOn(err)
			}
		}
	}

	// Proposals list per payload type
	payloadProposals := make(map[string][]*loader.ProposalData)
	for _, p := range *proposals.All() {
		payloadProposals[p.VoteType] = append(payloadProposals[p.VoteType], p)
	}

	// check if we have privacy committee members when we don't have private voteplans
	if len(payloadProposals["private"]) == 0 && len(committeePrivacyPublicKeys) > 0 {
		kit.FatalOn(fmt.Errorf(" %s provided, but no %s proposals found", "committee-privacy-public-key", "private"))
	}

	// check we have also privacy committee members when we have private voteplans
	if len(payloadProposals["private"]) > 0 && len(committeePrivacyPublicKeys) == 0 {
		log.Printf("%s proposals found, but no %s provided...building one for you in %s", "private", "committee-privacy-public-key", votePlanDir)

		csr, err := jcli.VotesCRSGenerate("", filepath.Join(votePlanDir, "committee.csr"))
		kit.FatalOn(err, "jcli.VotesCRSGenerate", kit.B2S(csr))

		commSKFile := filepath.Join(votePlanDir, "committee_communication_key.sk")
		commPKFile := filepath.Join(votePlanDir, "committee_communication_key.pk")

		commSK, err := jcli.VotesCommitteeCommunicationKeyGenerate("", commSKFile)
		kit.FatalOn(err, "jcli.VotesCommitteeCommunicationKeyGenerate", kit.B2S(commSK))
		commPK, err := jcli.VotesCommitteeCommunicationKeyToPublic(nil, commSKFile, commPKFile)
		kit.FatalOn(err, "jcli.VotesCommitteeCommunicationKeyGenerate", kit.B2S(commPK))

		memberSKFile := filepath.Join(votePlanDir, "committee_member_key.sk")
		memberPKFile := filepath.Join(votePlanDir, "committee_member_key.pk")

		memberSK, err := jcli.VotesCommitteeMemberKeyGenerate(kit.B2S(csr), 1, []string{kit.B2S(commPK)}, 0, "", "" /* memberSKFile */)
		kit.FatalOn(err, "jcli.VotesCommitteeMemberKeyGenerate", kit.B2S(memberSK))
		memberPK, err := jcli.VotesCommitteeMemberKeyToPublic(memberSK, "", "")
		kit.FatalOn(err, "jcli.VotesCommitteeMemberKeyToPublic", kit.B2S(memberPK))

		cf, err := os.Create(memberSKFile)
		kit.FatalOn(err, "memberSKFile CREATE")
		_, err = cf.Write(memberSK)
		kit.FatalOn(err, "memberSKFile WRITE")
		err = cf.Close()
		kit.FatalOn(err, "memberSKFile CLOSE")

		cf, err = os.Create(memberPKFile)
		kit.FatalOn(err, "memberPKFile CREATE")
		_, err = cf.Write(memberPK)
		kit.FatalOn(err, "memberPKFile WRITE")
		err = cf.Close()
		kit.FatalOn(err, "memberPKFile CLOSE")

		committeePrivacyPublicKeys = append(committeePrivacyPublicKeys, kit.B2S(memberPK))
		log.Println()
	}

	// save vote encryption key
	var (
		voteEncKeyFile string
		voteEncKey     []byte
	)

	if len(payloadProposals["private"]) > 0 && len(committeePrivacyPublicKeys) > 0 {
		voteEncKeyFile = filepath.Join(votePlanDir, "vote_encryption_key.pk")

		voteEncKey, err = jcli.VotesEncryptingVoteKey(committeePrivacyPublicKeys, "" /* voteEncKeyFile */)
		kit.FatalOn(err, "jcli.VotesEncryptingVoteKey", kit.B2S(voteEncKey))

		cf, err := os.Create(voteEncKeyFile)
		kit.FatalOn(err, "voteEncKeyFile CREATE")
		_, err = cf.Write(voteEncKey)
		kit.FatalOn(err, "voteEncKeyFile WRITE")
		err = cf.Close()
		kit.FatalOn(err, "voteEncKeyFile CLOSE")
	}

	// Calculate nr of needed voteplans since there is a limit of proposals a plan can have (255)
	// Taking in consideration also payload
	vpNeeded := 0
	for _, vpp := range payloadProposals {
		vpNeeded += votePlansNeeded(len(vpp), int(votePlanProposalsMax))
	}

	jcliVotePlans := make([]jcliVotePlan, vpNeeded)
	funds.First().VotePlans = make([]loader.ChainVotePlan, vpNeeded)

	jcliVotePlansCreated := 0
	for pt := range payloadProposals {
		vpi := 0
		// Generate proposals hash and associate it to a voteplan
		for i, proposal := range payloadProposals[pt] {

			// tmp - hash the proposal (TODO: decide what to hash in production, file bytes ???)
			externalID := blake2b.Sum256([]byte(proposal.Proposal.ID + proposal.InternalID + pt))
			proposal.ChainProposal.ExternalID = hex.EncodeToString(externalID[:])

			// retrieve the voteplan internal index based on the proposal index we are at
			// taking in consideration also previous payloads voteplans created
			vpi = (i / int(votePlanProposalsMax)) + jcliVotePlansCreated

			// Set payload once
			if jcliVotePlans[vpi].Payload == "" {
				jcliVotePlans[vpi].Payload = pt
			}

			// add proposal hash to the respective voteplan internal container
			jcliVotePlans[vpi].Proposals = append(
				jcliVotePlans[vpi].Proposals,
				jcliProposal{
					ExternalID: proposal.ChainProposal.ExternalID,
					Options:    uint8(len(proposal.ChainProposal.VoteOptions)),
					Action:     proposal.VoteAction,
				},
			)
		}
		jcliVotePlansCreated = jcliVotePlansCreated + vpi + 1 // vpi is an index so we need +1
	}

	certSignersFiles := make([]string, 0) //, 0, len(leaders))
	for i := range leaders {
		// we need a secret key
		if leaders[i].skFile == "" {
			continue
		}
		certSignersFiles = append(certSignersFiles, leaders[i].skFile)
		break // right now only one key is needed to sign a certificate so bail as soon as we have one
	}

	if *block0Voteplans && len(certSignersFiles) == 0 {
		kit.FatalOn(fmt.Errorf("no [%s] available to sign the block0 certificate(s)", "bft leader SK (secret key)"), "block0-voteplan")
	}

	// Generate voteplan certificates and id
	for i := range jcliVotePlans {

		jcliVotePlans[i].VoteStart = voteStart
		jcliVotePlans[i].VoteEnd = voteEnd
		jcliVotePlans[i].CommitteeEnd = committeeEnd

		// Add committee privacy public keys if VotePlan payload is private
		switch jcliVotePlans[i].Payload {
		case "private":
			jcliVotePlans[i].CommitteeMemberPublicKeys = committeePrivacyPublicKeys
		case "public":
			jcliVotePlans[i].CommitteeMemberPublicKeys = []string{}
		}

		stdinConfig, err := json.MarshalIndent(jcliVotePlans[i], "", " ")
		kit.FatalOn(err, "json.Marshal VotePlan Config")

		ucert, err := jcli.CertificateNewVotePlan(stdinConfig, "", "")
		kit.FatalOn(err, "CertificateNewVotePlan", kit.B2S(ucert))

		id, err := jcli.CertificateGetVotePlanID(ucert, "", "")
		kit.FatalOn(err, "CertificateGetVotePlanID:", kit.B2S(id))

		jcliVotePlans[i].VotePlanID = kit.B2S(id)

		// Assuming that bft leaders will be part of committee signing keys
		scert := []byte{}
		if *block0Voteplans {
			scert, err = jcli.CertificateSign(ucert, certSignersFiles, "", "")
			kit.FatalOn(err, "CertificateSign:", kit.B2S(scert))

			jcliVotePlans[i].Certificate = kit.B2S(scert)
		}

		// VotePlan - configuration
		vpj, err := os.Create(filepath.Join(votePlanDir, jcliVotePlans[i].Payload+"_voteplan_"+kit.B2S(id)+".json"))
		kit.FatalOn(err, "VotePlan json CREATE", kit.B2S(id))
		_, err = vpj.Write(stdinConfig)
		kit.FatalOn(err, "VotePlan json WRITE", kit.B2S(id))
		err = vpj.Close()
		kit.FatalOn(err, "VotePlan json CLOSE", kit.B2S(id))

		// VotePlan - unsigned certificate
		vpuc, err := os.Create(filepath.Join(votePlanDir, jcliVotePlans[i].Payload+"_voteplan_"+kit.B2S(id)+".cert-unsigned"))
		kit.FatalOn(err, "VotePlan cert-unsigned CREATE", kit.B2S(id))
		_, err = vpuc.Write(ucert)
		kit.FatalOn(err, "VotePlan cert-unsigned WRITE", kit.B2S(id))
		err = vpuc.Close()
		kit.FatalOn(err, "VotePlan cert-unsigned CLOSE", kit.B2S(id))

		// VotePlan - signed certificate
		if len(scert) > 0 {
			vpsc, err := os.Create(filepath.Join(votePlanDir, jcliVotePlans[i].Payload+"_voteplan_"+kit.B2S(id)+".cert-signed"))
			kit.FatalOn(err, "VotePlan cert-signed CREATE", kit.B2S(id))
			_, err = vpsc.Write(scert)
			kit.FatalOn(err, "VotePlan cert-signed WRITE", kit.B2S(id))
			err = vpsc.Close()
			kit.FatalOn(err, "VotePlan cert-signed CLOSE", kit.B2S(id))
		}

		// Update Fund info with VotePlans Data - TODO: when defined update to support multiple funds
		funds.First().VotePlans[i].VotePlanID = jcliVotePlans[i].VotePlanID
		funds.First().VotePlans[i].VoteStart = voteStartTime.Format(*dateTimeFormat)
		funds.First().VotePlans[i].VoteEnd = voteEndTime.Format(*dateTimeFormat)
		funds.First().VotePlans[i].CommitteeEnd = committeeEndTime.Format(*dateTimeFormat)
		funds.First().VotePlans[i].Payload = jcliVotePlans[i].Payload

		funds.First().VotePlans[i].FundID = funds.First().FundID
		funds.First().VotePlans[i].VpInternalID = strconv.Itoa(i + 1)

		// set chain_vote_encryption_key for the api
		if jcliVotePlans[i].Payload == "private" {
			funds.First().VotePlans[i].VoteEncryptionKey = kit.B2S(voteEncKey)
		}

		// Update proposals index and voteplan
		for pi, prop := range jcliVotePlans[i].Proposals {
			// TODO: fix this search
			proposal := datastore.FilterSingle(
				proposals.All(),
				func(v *loader.ProposalData) bool {
					return v.ChainProposal.ExternalID == prop.ExternalID
				},
			)

			proposal.ChainProposal.Index = uint8(pi)
			proposal.ChainVotePlan = &(funds.First().VotePlans[i])
		}

		if *block0Voteplans {
			// Vote Plans add certificate to block0
			err = block0cfg.AddInitialCertificate(jcliVotePlans[i].Certificate)
			kit.FatalOn(err, "AddInitialCertificate")
		}
	}

	log.Printf("VIT - Voteplan(s) data are dumped at (%s)", votePlanDir)
	log.Println()

	//////////////////////////////////////////////
	/* TODO: TMP - remove once/if properly defined */
	if funds.First().StartTime == "" {
		funds.First().StartTime = voteStartTime.Format(*dateTimeFormat)
	}
	if funds.First().EndTime == "" {
		funds.First().EndTime = voteEndTime.Format(*dateTimeFormat)
	}
	if funds.First().VotingPowerInfo == "" {
		funds.First().VotingPowerInfo = funds.First().StartTime
	}
	if funds.First().RewardsInfo == "" {
		funds.First().RewardsInfo = committeeEndTime.Add(7 * epochDur).Format(*dateTimeFormat)
	}
	if funds.First().NextStartTime == "" {
		funds.First().NextStartTime = committeeEndTime.Add(15 * epochDur).Format(*dateTimeFormat)
	}
	/* TODO: TMP - remove once/if properly defined */
	//////////////////////////////////////////////

	// FUNDS - dump
	fundsFile, err := os.Create(filepath.Join(vitStationDir, "sql_funds.csv"))
	kit.FatalOn(err, "Funds csv CREATE")
	f := []*loader.FundData{funds.First()}
	err = gocsv.MarshalFile(&f, fundsFile) // Use this to save the CSV back to the file
	kit.FatalOn(err, "Funds csv WRITE")
	err = fundsFile.Close()
	kit.FatalOn(err, "Funds csv CLOSE")

	// VOTEPLANS - dump
	votePlansFile, err := os.Create(filepath.Join(vitStationDir, "sql_voteplans.csv"))
	kit.FatalOn(err, "Voteplans csv CREATE")
	vp := funds.First().VotePlans
	err = gocsv.MarshalFile(&vp, votePlansFile)
	kit.FatalOn(err, "Voteplans csv WRITE")
	err = votePlansFile.Close()
	kit.FatalOn(err, "Voteplans csv CLOSE")

	// PROPOSALS - dump
	proposalsFile, err := os.Create(filepath.Join(vitStationDir, "sql_proposals.csv"))
	kit.FatalOn(err, "Proposals csv CREATE")
	p := proposals.All()
	err = gocsv.MarshalFile(p, proposalsFile)
	kit.FatalOn(err, "Proposals csv WRITE")
	err = proposalsFile.Close()
	kit.FatalOn(err, "Proposals csv CLOSE")

	log.Printf("VIT - Station data are dumped at (%s)", vitStationDir)
	log.Println()

	block0Yaml, err := block0cfg.ToYaml()
	kit.FatalOn(err)

	if *genesisExtraDataPath != "" {
		bulkExtraData, err := ioutil.ReadFile(*genesisExtraDataPath)
		kit.FatalOn(err)
		if len(bulkExtraData) > 0 {
			block0Yaml = append(block0Yaml, bulkExtraData...)
		}
	}

	// need this file for starting the node (--genesis-block)
	block0BinFile := filepath.Join(workingDir, "VIT-block0.bin")

	// keep also the text block0 config
	block0TxtFile := filepath.Join(workingDir, "VIT-block0.yaml")

	// block0BinFile will be created by jcli
	block0Bin, err := jcli.GenesisEncode(block0Yaml, "", block0BinFile)
	kit.FatalOn(err, kit.B2S(block0Bin), kit.B2S(block0Yaml))

	block0Hash, err := jcli.GenesisHash(block0Bin, "")
	kit.FatalOn(err, kit.B2S(block0Hash))

	// block0TxtFile will be created by jcli
	block0Txt, err := jcli.GenesisDecode(block0Bin, "", block0TxtFile)
	kit.FatalOn(err, kit.B2S(block0Txt))

	//////////////////////
	//  secrets config  //
	//////////////////////

	for i := range leaders {
		// we need secret key, but only public ones may have been provided
		if leaders[i].sk == "" {
			continue
		}

		secretCfg := jnode.NewSecretConfig()

		secretCfg.Bft.SigningKey = leaders[i].sk

		secretCfgYaml, err := secretCfg.ToYaml()
		kit.FatalOn(err)

		// need this file for starting the node (--secret)
		secretCfgFile := leaders[i].skFile + ".yaml"
		err = ioutil.WriteFile(secretCfgFile, secretCfgYaml, 0744)
		kit.FatalOn(err)

		leaders[i].cfgFile = secretCfgFile
	}

	///////////////////
	//  node config  //
	///////////////////

	nodeCfg := jnode.NewNodeConfig()

	nodeCfg.Storage = filepath.Join(workingDir, "storage")

	nodeCfg.SkipBootstrap = *skipBootstrap
	nodeCfg.BootstrapFromTrustedPeers = true

	nodeCfg.Rest.Listen = restAddress
	nodeCfg.Rest.Cors.AllowedOrigins = strings.Split(*restCorsAllowed, ",")
	nodeCfg.Rest.Cors.MaxAgeSecs = 0

	nodeCfg.P2P.PublicAddress = p2pListenAddress
	nodeCfg.P2P.ListenAddress = p2pListenAddress
	nodeCfg.P2P.AllowPrivateAddresses = true
	nodeCfg.P2P.MaxBootstrapAttempts = 5

	nodeCfg.Log.Level = *nodeLogLevel

	nodeCfg.Explorer.Enabled = *explorerEnabled

	for i := range leaders {
		// we need secret key to build config file, but only public ones may have been provided
		if leaders[i].cfgFile == "" {
			continue
		}
		nodeCfg.AddSecretFile(leaders[i].cfgFile)
	}

	nodeCfgYaml, err := nodeCfg.ToYaml()
	kit.FatalOn(err)

	// need this file for starting the node (--config)
	nodeCfgFile := filepath.Join(workingDir, "node-config.yaml")
	err = ioutil.WriteFile(nodeCfgFile, nodeCfgYaml, 0644)
	kit.FatalOn(err)

	//////////////////////
	// running the node //
	//////////////////////

	// Check for jörmungandr binary. Local folder first, then PATH
	jnodeBin, err := kit.FindExecutable("jormungandr", "jor_bins")
	kit.FatalOn(err, jnodeBin)
	jnode.BinName(jnodeBin)

	// get jörmungandr version
	jormungandrVersion, err := jnode.VersionFull()
	kit.FatalOn(err, kit.B2S(jormungandrVersion))

	node := jnode.NewJnode()

	node.WorkingDir = workingDir
	node.GenesisBlock = block0BinFile
	node.ConfigFile = nodeCfgFile

	for i := range leaders {
		// we need secret key to build config file, but only public ones may have been provided so no leader config possible
		if leaders[i].cfgFile == "" {
			continue
		}
		node.AddSecretFile(leaders[i].cfgFile)
	}

	// Run the node (Start + Wait)
	if *startNode {
		node.Stdout, err = os.Create(filepath.Join(workingDir, "stdout.log"))
		kit.FatalOn(err)
		node.Stderr, err = os.Create(filepath.Join(workingDir, "stderr.log"))
		kit.FatalOn(err)

		err = os.Setenv("RUST_BACKTRACE", "full")
		kit.FatalOn(err, "Failed to set env (RUST_BACKTRACE=full)")

		err = node.Run()
		if err != nil {
			log.Fatalf("node.Run FAILED: %v", err)
		}
	}

	//////////////////////
	// VIT station data //
	//////////////////////

	// vit-servicing-station-cli
	var (
		vcliBin     string
		vcliVersion []byte

		vitDb      = filepath.Join(vitStationDir, "database.sqlite3")
		vitCfgFile = filepath.Join(vitStationDir, "vit_cfg.json")
	)

	// Check for vit-servicing-station-cli binary. Local folder first (vit_bins), then PATH
	vcliBin, err = kit.FindExecutable("vit-servicing-station-cli", "vit_bins")
	if err != nil {
		log.Printf("***** %s - DB data related to %s will NOT be generated", err.Error(), "vit-servicing-station")
		vcliBin = ""
	} else {
		vcli.BinName(vcliBin)
	}

	if vcliBin != "" {
		// get vit-servicing-station-cli version
		vcliVersion, err = vcli.Version()
		kit.FatalOn(err, kit.B2S(vcliVersion))

		// init database
		out, err := vcli.DbInit(vitDb)
		kit.FatalOn(err, "vcli.DbInit", kit.B2S(out))

		// populate the database with already dumped data
		out, err = vcli.CsvDataLoad(vitDb, fundsFile.Name(), proposalsFile.Name(), votePlansFile.Name())
		kit.FatalOn(err, "vcli.CsvDataLoad", kit.B2S(out))
	}

	// vit-servicing-station-server
	var (
		vstationBin     string
		vstationVersion []byte
	)

	vs := vstation.NewVstation()
	vs.WorkingDir = vitStationDir
	vs.Address = *vitAddrPort
	vs.Block0Path = block0BinFile
	vs.DbUrl = vitDb
	vs.Log.LogLevel = *vitLogLevel
	vs.Log.LogOutputPath = filepath.Join(vitStationDir, "vit_station.log")
	vs.Cors.AllowedOrigins = strings.Split(*restCorsAllowed, ",")

	vsJson, err := json.MarshalIndent(&vs, "", " ")
	kit.FatalOn(err, "vstation json.MarshalIndent")
	err = ioutil.WriteFile(vitCfgFile, vsJson, 0755)
	kit.FatalOn(err, "vstation ioutil.WriteFile", vitCfgFile)

	// Check for vit-servicing-station-server binary. Local folder first (vit_bins), then PATH
	vstationBin, err = kit.FindExecutable("vit-servicing-station-server", "vit_bins")
	if err != nil {
		log.Printf("***** %s", err.Error())
		vstationBin = ""
	} else {
		vstation.BinName(vstationBin)
	}

	if vstationBin != "" {
		// get vit-servicing-station-server version
		vstationVersion, err = vstation.Version()
		kit.FatalOn(err, kit.B2S(vstationVersion))

		if *startVit {
			err = vs.Run()
			if err != nil {
				log.Fatalf("vs.Run FAILED: %v", err)
			}
		}
	}

	////////////////////
	// internal proxy //
	////////////////////

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
	log.Printf("VIT - BFT Genesis: %s - %d", "COMMITTEE", len(block0cfg.BlockchainConfiguration.Committees)+len(block0cfg.BlockchainConfiguration.ConsensusLeaderIds))
	log.Printf("VIT - BFT Genesis: %s - %d", "VOTEPLANS", len(jcliVotePlans))
	log.Printf("VIT - BFT Genesis: %s - %d", "PROPOSALS", proposals.Total())
	log.Println()

	log.Printf("JÖRMUNGANDR listening at: %s - %v", p2pListenAddress, *startNode)
	log.Printf("JÖRMUNGANDR Rest API available at: http://%s/api - %v", restAddress, *startNode)
	log.Println()
	log.Printf("VIT-STATION API available at: http://%s/api - %v", *vitAddrPort, *startVit)
	log.Println()
	log.Printf("APP - PROXY Rest API available at: http://%s/api", proxyAddress)
	log.Println()
	log.Println("VIT - BFT Genesis Node - Running...")
	log.Println()

	if vstationBin != "" {
		log.Printf("\t%s %s", vstationBin, strings.Join(vs.BuildCmdArg(), " "))
		log.Println()
	}

	log.Printf("\t%s %s", jnodeBin, strings.Join(node.BuildCmdArg(), " "))
	log.Println()

	if *startVit && vstationBin != "" {
		vs.Wait() // Wait for the vit station to stop.
	}

	if *startNode {
		node.Wait() // Wait for the node to stop.
	}

	if *allowNodeRestart || !*startNode {
		switch {
		case !*startNode:
			log.Println("The node has to be started manually or issue SIGINT/SIGTERM again.")
		case *allowNodeRestart:
			log.Println("The node has stopped. Please start the node manually and keep the same running config or issue SIGINT/SIGTERM again.")
		}

		// Listen for the service syscalls
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs

		if *shutdownNode {
			// Attempt node shutdown in case the node was restarted manually again
			_, _ = jcli.RestShutdown("http://"+restAddress+"/api", "")
		}
	}

	log.Println("...VIT - BFT Genesis Node - Done") // All done. Node has stopped.
}
