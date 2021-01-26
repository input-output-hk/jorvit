// Package vstation provides vit-servicing-station-server binary helpers.
package vstation

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	vstationName = "vit-servicing-station-server"
)

// Vstation contains the vit-servicing-station-server commandline/file config parameters.
type Vstation struct {
	InSettingsFile  string `json:"-"` // --in-settings-file <in-settings-file>
	OutSettingsFile string `json:"-"` // --out-settings-file <out-settings-file>

	EnableApiTokens bool   `json:"enable_api_tokens"` // --enable-api-tokens
	Address         string `json:"address"`           // --address <address>
	Block0Path      string `json:"block0_path"`       // --block0-path <block0-path>
	DbUrl           string `json:"db_url"`            // --db-url <db-url>

	Log struct {
		LogLevel      string `json:"log_level"`       // --log-level <log-level>
		LogOutputPath string `json:"log_output_path"` // --log-output-path <log-output-path>
	} `json:"log"`

	Cors struct {
		AllowedOrigins []string `json:"allowed_origins"` // --allowed-origins <allowed-origins>
		MaxAgeSecs     uint     `json:"max_age_secs"`    // --max-age-secs <max-age-secs>
	} `json:"cors"`

	Tls struct {
		CertFile    string `json:"cert_file"`     // --cert-file <cert-file>
		PrivKeyFile string `json:"priv_key_file"` // --priv-key-file <priv-key-file>
	} `json:"tls"`

	// Extra
	WorkingDir string    `json:"-"`
	Stdout     io.Writer `json:"-"`
	Stderr     io.Writer `json:"-"`
	// internal usage
	cmd  *exec.Cmd
	done chan struct{}
}

// NewVstation returns a Vstation with some defaults.
func NewVstation() *Vstation {
	return &Vstation{
		WorkingDir: os.TempDir(),
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		done:       make(chan struct{}),
	}
}

// BuildCmdArg and return a slice
func (vstation *Vstation) BuildCmdArg() []string {
	var arg []string

	if vstation.InSettingsFile != "" {
		arg = append(arg, "--in-settings-file", vstation.InSettingsFile)
	}

	if vstation.OutSettingsFile != "" {
		arg = append(arg, "--out-settings-file", vstation.OutSettingsFile)
	}

	if vstation.EnableApiTokens {
		arg = append(arg, "--enable_api_tokens")
	}

	if vstation.Address != "" {
		arg = append(arg, "--address", vstation.Address)
	}

	if vstation.Block0Path != "" {
		arg = append(arg, "--block0-path", vstation.Block0Path)
	}

	if vstation.DbUrl != "" {
		arg = append(arg, "--db-url", vstation.DbUrl)
	}

	if vstation.Log.LogLevel != "" {
		arg = append(arg, "--log-level", vstation.Log.LogLevel)
	}
	if vstation.Log.LogOutputPath != "" {
		arg = append(arg, "--log-output-path", vstation.Log.LogOutputPath)
	}

	if vstation.Cors.MaxAgeSecs != 0 {
		arg = append(arg, "--max-age-secs", strconv.Itoa(int(vstation.Cors.MaxAgeSecs)))
	}

	if len(vstation.Cors.AllowedOrigins) > 0 {
		arg = append(arg, "--allowed-origins", strings.Join(vstation.Cors.AllowedOrigins, ";"))
	}

	return arg
}

// Run starts the node.
func (vstation *Vstation) Run() error {
	vstation.cmd = exec.Command(vstationName, vstation.BuildCmdArg()...)

	vstation.cmd.Dir = vstation.WorkingDir
	vstation.cmd.Stdout = vstation.Stdout
	vstation.cmd.Stderr = vstation.Stderr

	err := vstation.cmd.Start()
	if err != nil {
		return err
	}

	// FIXME: find an effective way to catch errors of stderr
	// since cmd.Start does not care about them.
	// Ex: config errors will cause Start() to report no errors
	//     when in fact the node reports errors on stderr and stops.

	vstation.handleSigs()
	go vstation.cmdWait()

	return nil
}

// cmdWait for the node to terminate,
// and close the stop channel.
func (vstation *Vstation) cmdWait() {
	err := vstation.cmd.Wait()
	if err != nil {
		log.Printf("cmd.Wait() - %v", err) // FIXME: handle shutdown
	}
	select {
	case <-vstation.done:
	default:
		close(vstation.done)
	}
}

// handleSigs SIGINT + SIGTERM
func (vstation *Vstation) handleSigs() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		_ = vstation.Stop()
	}()
}

// Wait for the node to stop.
func (vstation *Vstation) Wait() {
	<-vstation.done
}

// Stop the node if running.
func (vstation *Vstation) Stop() error {
	if vstation.cmd.Process == nil {
		return fmt.Errorf("%s : exec: not started", "vstation.Stop")
	}
	return vstation.cmd.Process.Kill()
}

// StopAfter seconds.
func (vstation *Vstation) StopAfter(d time.Duration) error {
	if vstation.cmd.Process == nil {
		return fmt.Errorf("%s : exec: not started", "vstation.StopAfter")
	}

	go func() {
		select {
		case <-vstation.done:
		case <-time.After(d):
			_ = vstation.Stop()
		}
	}()
	return nil
}

// Pid provided for the running node process.
func (vstation *Vstation) Pid() int {
	if vstation.cmd.Process == nil {
		return 0
	}
	return vstation.cmd.Process.Pid
}

// BinName set the executable name/full path if not the default one.
func BinName(name string) {
	vstationName = name
}
