//$(which go) run $0 $@; exit $?

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/input-output-hk/jorvit/internal/kit"
	"github.com/input-output-hk/jorvit/internal/loader"
)

var (
	// Version and build info that can be set on build
	Version    = "dev"
	CommitHash = "none"
	BuildDate  = "unknown"
)

type ChainTime struct {
	Epoch  int64 `json:"epoch"`
	SlotID int64 `json:"slot_id"`
}

type VoteOption struct {
	Start uint8 `json:"start"`
	End   uint8 `json:"end"`
}

type TallyResult map[string]struct {
	Result struct {
		Options VoteOption `json:"options"`
		Results []uint     `json:"results"`
	} `json:"result"`
}

type VoteProposal struct {
	Index      uint8       `json:"index"`
	ProposalID string      `json:"proposal_id"`
	Options    VoteOption  `json:"options"`
	Tally      TallyResult `json:"tally"`
	VotesCast  uint        `json:"votes_cast"`
}

type VotePlans struct {
	ID                  string         `json:"id"`
	Payload             string         `json:"payload"`
	VoteStart           ChainTime      `json:"vote_start"`
	VoteEnd             ChainTime      `json:"vote_end"`
	CommitteeEnd        ChainTime      `json:"committee_end"`
	CommitteeMemberKeys []string       `json:"committee_member_keys"`
	Proposals           []VoteProposal `json:"proposals"`
}

// TallyOptions total 16 choices available (0-15)
type TallyOptions struct {
	// TODO: this is ...
	Tally00 uint `json:"tally_0"  csv:"tally_0_BLANK"`
	Tally01 uint `json:"tally_1"  csv:"tally_1_YES"`
	Tally02 uint `json:"tally_2"  csv:"tally_2_NO"`
	Tally03 uint `json:"-"        csv:"-"`
	Tally04 uint `json:"-"        csv:"-"`
	Tally05 uint `json:"-"        csv:"-"`
	Tally06 uint `json:"-"        csv:"-"`
	Tally07 uint `json:"-"        csv:"-"`
	Tally08 uint `json:"-"        csv:"-"`
	Tally09 uint `json:"-"        csv:"-"`
	Tally10 uint `json:"-"        csv:"-"`
	Tally11 uint `json:"-"        csv:"-"`
	Tally12 uint `json:"-"        csv:"-"`
	Tally13 uint `json:"-"        csv:"-"`
	Tally14 uint `json:"-"        csv:"-"`
	Tally15 uint `json:"-"        csv:"-"`
}

type ProposalsResult struct {
	loader.ProposalData
	TallyOptions
}

func getData(client *http.Client, u *url.URL, dst interface{}) error {
	var (
		data []byte
		err  error
	)

	switch u.Scheme {
	case "http", "https":
		data, err = httpGet(client, u.String())
	case "file":
		data, err = ioutil.ReadFile(u.Host + u.Path)
	default:
		err = fmt.Errorf("unknown schema: [%s] from [%s]", u.Scheme, u.String())
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &dst)
}

func httpGet(client *http.Client, u string) ([]byte, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Close = true

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(ioutil.Discard, res.Body) // WE READ THE BODY
	if err != nil {
		return nil, err
	}
	res.Body.Close()

	return data, nil
}

func main() {
	var (
		// Http
		client = http.Client{
			Timeout: time.Second * 10,
		}
		// Data
		votePlans []VotePlans
		proposals []ProposalsResult
		funds     loader.FundData
		// Flags
		serviceUrl   = flag.String("service-addr", "https://servicing-station.vit.iohk.io", "Address of remote service, or file://")
		nodeUrl      = flag.String("node-addr", "https://servicing-station.vit.iohk.io", "Address of remote service, or file://")
		votePlansUrl = flag.String("vote-plans", "/api/v0/vote/active/plans", "Endpoint (or file path) containing  tally results from the chain, added to \"node-addr\"")
		proposalsUrl = flag.String("proposals", "/api/v0/proposals", "Endpoint (or file path) containing proposals, added to \"service-addr\"")
		fundsUrl     = flag.String("funds", "/api/v0/fund", "Endpoint (or file path) containing fund info, added to \"service-addr\"")
		timeout      = flag.String("http-timeout", "10s", "Http request timeout")
		// Flags - TallyResult file
		tallyResultFile = flag.String("result-file", "TallyResult.csv", "File name of the output result")
		// Flags - version info
		version = flag.Bool("version", false, "Print current app version and build info")
	)

	flag.Parse()

	// version info
	if *version {
		fmt.Printf("Version - %s\n", Version)
		fmt.Printf("Commit  - %s\n", CommitHash)
		fmt.Printf("Date    - %s\n", BuildDate)
		os.Exit(0)
	}

	// Http timeout
	timeoutDur, err := time.ParseDuration(*timeout)
	kit.FatalOn(err, "http-timeout:", *timeout)
	client.Timeout = timeoutDur

	// Parse URI
	vpUrl, err := url.ParseRequestURI(*nodeUrl + *votePlansUrl)
	kit.FatalOn(err, "url.ParseRequestURI:", *votePlansUrl)
	prUrl, err := url.ParseRequestURI(*serviceUrl + *proposalsUrl)
	kit.FatalOn(err, "url.ParseRequestURI:", *proposalsUrl)
	fuUrl, err := url.ParseRequestURI(*serviceUrl + *fundsUrl)
	kit.FatalOn(err, "url.ParseRequestURI:", *fundsUrl)

	// Fetch Data
	kit.FatalOn(getData(&client, vpUrl, &votePlans), "getData VotePlans")
	kit.FatalOn(getData(&client, prUrl, &proposals), "getData Proposals")
	kit.FatalOn(getData(&client, fuUrl, &funds), "getData Funds")

	for i := range proposals {
		for x := range votePlans {
			// skip other voteplans id
			if proposals[i].VotePlanID != votePlans[x].ID {
				continue
			}

			for y := range votePlans[x].Proposals {
				// skip other proposals index
				if proposals[i].ChainProposal.Index != votePlans[x].Proposals[y].Index {
					continue
				}
				// skip other proposals id - in theory this should not never be the case since we matched index
				if proposals[i].ChainProposal.ExternalID != votePlans[x].Proposals[y].ProposalID {
					continue
				}

				// we should have only one payload
				if len(votePlans[x].Proposals[y].Tally) != 1 {
					continue
				}

				for _, res := range votePlans[x].Proposals[y].Tally {
					for r, tr := range res.Result.Results {
						switch r {
						// TODO: this is ...
						case 0:
							proposals[i].TallyOptions.Tally00 = tr
						case 1:
							proposals[i].TallyOptions.Tally01 = tr
						case 2:
							proposals[i].TallyOptions.Tally02 = tr
						case 3:
							proposals[i].TallyOptions.Tally03 = tr
						case 4:
							proposals[i].TallyOptions.Tally04 = tr
						case 5:
							proposals[i].TallyOptions.Tally05 = tr
						case 6:
							proposals[i].TallyOptions.Tally06 = tr
						case 7:
							proposals[i].TallyOptions.Tally07 = tr
						case 8:
							proposals[i].TallyOptions.Tally08 = tr
						case 9:
							proposals[i].TallyOptions.Tally09 = tr
						case 10:
							proposals[i].TallyOptions.Tally10 = tr
						case 11:
							proposals[i].TallyOptions.Tally11 = tr
						case 12:
							proposals[i].TallyOptions.Tally12 = tr
						case 13:
							proposals[i].TallyOptions.Tally13 = tr
						case 14:
							proposals[i].TallyOptions.Tally14 = tr
						case 15:
							proposals[i].TallyOptions.Tally15 = tr
						}
					}
				}
			}
		}
	}

	// TallyResult - dump
	tallyFile, err := os.Create(*tallyResultFile)
	kit.FatalOn(err, "tallyFile csv CREATE", *tallyResultFile)
	err = gocsv.MarshalFile(&proposals, tallyFile)
	kit.FatalOn(err, "tallyFile csv WRITE", *tallyResultFile)
	err = tallyFile.Close()
	kit.FatalOn(err, "tallyFile csv CLOSE", *tallyResultFile)

	fmt.Printf("Result ready at: %s\n", *tallyResultFile)
}
