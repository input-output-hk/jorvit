//$(which go) run $0 $@; exit $?

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/input-output-hk/jorvit/internal/kit"
	"github.com/input-output-hk/jorvit/internal/wallet"
	"github.com/rinor/jorcli/jcli"
	"github.com/rinor/jorcli/jnode"
	"github.com/skip2/go-qrcode"
)

var (
	leaderSK = []byte("ed25519_sk1pl4vp0grkl2puspv4c3hwhz89r68yjyzalc78pyt0pujpmk8mxkq6kpc5j")
	wallets  = wallet.SampleWallets()
)

func main() {
	var (
		// Rest
		restAddr, restPort = "0.0.0.0", 8001
		restAddress        = restAddr + ":" + strconv.Itoa(restPort)
		// P2P
		p2pIPver, p2pProto           = "ip4", "tcp"
		p2pListenAddr, p2pListenPort = "127.0.0.11", 9001
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
	leaderPK, err := jcli.KeyToPublic(leaderSK, "", "")
	kit.FatalOn(err, kit.B2S(leaderPK))

	/////////////////////
	//  block0 config  //
	/////////////////////

	block0cfg := jnode.NewBlock0Config()

	block0Discrimination := "production"
	if discrimination == "testing" {
		block0Discrimination = "test"
	}

	// set/change config params
	block0cfg.BlockchainConfiguration.Block0Date = block0Date()
	block0cfg.BlockchainConfiguration.Block0Consensus = consensus
	block0cfg.BlockchainConfiguration.Discrimination = block0Discrimination

	block0cfg.BlockchainConfiguration.SlotDuration = 10
	block0cfg.BlockchainConfiguration.SlotsPerEpoch = 6

	block0cfg.BlockchainConfiguration.LinearFees.Certificate = 5
	block0cfg.BlockchainConfiguration.LinearFees.Coefficient = 3
	block0cfg.BlockchainConfiguration.LinearFees.Constant = 2

	// Bft Leader
	err = block0cfg.AddConsensusLeader(kit.B2S(leaderPK))
	kit.FatalOn(err)

	// add legacy funds
	for i := range wallets {
		wallets[i].Totals = 0
		for _, lf := range wallets[i].Funds {
			err = block0cfg.AddInitialLegacyFund(lf.Address, lf.Value)
			kit.FatalOn(err)
			wallets[i].Totals += lf.Value
		}
	}

	block0Yaml, err := block0cfg.ToYaml()
	kit.FatalOn(err)

	// need this file for starting the node (--genesis-block)
	block0BinFile := workingDir + string(os.PathSeparator) + "VIT-block0.bin"

	// keep also the text block0 config
	block0TxtFile := workingDir + string(os.PathSeparator) + "VIT-block0.yaml"

	// block0BinFile will be created by jcli
	block0Bin, err := jcli.GenesisEncode(block0Yaml, "", block0BinFile)
	kit.FatalOn(err, kit.B2S(block0Bin))

	block0Hash, err := jcli.GenesisHash(block0Bin, "")
	kit.FatalOn(err, kit.B2S(block0Hash))

	// block0TxtFile will be created by jcli
	block0Txt, err := jcli.GenesisDecode(block0Bin, "", block0TxtFile)
	kit.FatalOn(err, kit.B2S(block0Txt))

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

	// JSON
	// walletsJson, err := json.Marshal(&wallets)
	// fatalOn(err, "json.Marshall - wallets")
	// var jsonWallets []wallet.Wallet
	// err = json.Unmarshal(walletsJson, &jsonWallets)
	// fatalOn(err, "json.Unmarshal - wallets")
	// fmt.Printf("%s", walletsJson)
	// qrPrint(jsonWallets)

	// return

	// Run the node (Start + Wait)
	err = os.Setenv("RUST_BACKTRACE", "full")
	kit.FatalOn(err, "Failed to set env (RUST_BACKTRACE=full)")

	err = node.Run()
	if err != nil {
		log.Fatalf("node.Run FAILED: %v", err)
	}

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
	log.Printf("VIT - BFT Genesis: %s", "NO COMMITTEE - YET")
	log.Printf("VIT - BFT Genesis: %s", "NO VOTEPLANS - YET")
	log.Println()
	log.Printf("VIT - BFT Genesis: %s", "Wallets available for recovery")

	qrPrint(wallets)

	log.Println()
	log.Printf("Rest API available at: http://%s/api", restAddress)
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

// get a fixed date if possible.
// needed for testing only to have a known genesis hash.
func block0Date() int64 {
	block0Date, err := time.Parse(time.RFC3339, "2020-05-01T00:00:00.000Z")
	if err != nil {
		return time.Now().Unix()
	}
	return block0Date.Unix()
}
