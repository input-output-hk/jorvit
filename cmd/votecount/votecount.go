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
	"time"

	"github.com/input-output-hk/jorvit/internal/kit"
	"github.com/input-output-hk/jorvit/internal/loader"
)

type ChainTime struct {
	Epoch  int64 `json:"epoch"`
	SlotID int64 `json:"slot_id"`
}

type VoteOption struct {
	End   uint8 `json:"end"`
	Start uint8 `json:"start"`
}

type PublicTallyResult struct {
	Public struct {
		Result struct {
			Options VoteOption `json:"options"`
			Results []int64    `json:"results"`
		} `json:"result"`
	} `json:"Public"`
}

type PublicVoteChoice struct {
	Public struct {
		Choice int64 `json:"choice"`
	} `json:"Public"`
}

type VoteProposal struct {
	Index      uint8                       `json:"index"`
	ProposalID string                      `json:"proposal_id"`
	Options    VoteOption                  `json:"options"`
	Tally      *PublicTallyResult          `json:"tally"`
	Votes      map[string]PublicVoteChoice `json:"votes"`
}

type VotePlans struct {
	ID           string         `json:"id"`
	Payload      string         `json:"payload"`
	VoteStart    ChainTime      `json:"vote_start"`
	VoteEnd      ChainTime      `json:"vote_end"`
	CommitteeEnd ChainTime      `json:"committee_end"`
	Proposals    []VoteProposal `json:"proposals"`
}

func getVotePlans(client *http.Client, u *url.URL) ([]VotePlans, error) {
	var dst []VotePlans

	data, err := getData(client, u)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &dst)
	if err != nil {
		return nil, err
	}

	return dst, nil
}

func getProposals(client *http.Client, u *url.URL) ([]loader.ProposalData, error) {
	var dst []loader.ProposalData

	data, err := getData(client, u)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &dst)
	if err != nil {
		return nil, err
	}

	return dst, nil
}

func getFunds(client *http.Client, u *url.URL) ([]loader.FundData, error) {
	var dst loader.FundData

	data, err := getData(client, u)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &dst)
	if err != nil {
		return nil, err
	}

	return []loader.FundData{dst}, nil
}

func getData(client *http.Client, u *url.URL) ([]byte, error) {
	var (
		data []byte
		err  error
	)

	switch u.Scheme {
	case "http", "https":
		data, err = httpGet(client, u.String())
	case "file", "":
		data, err = ioutil.ReadFile(u.String())
	default:
		err = fmt.Errorf("unknown schema: [%s] from [%s]", u.Scheme, u.String())
	}

	return data, err
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
		// Flags
		serviceUrl   = flag.String("service-addr", "http://127.0.0.1:8000", "Address of remote service, or file://")
		votePlansUrl = flag.String("vote-plans", "/api/v0/vote/active/plans", "Endpoint (or file path) containing  tally results from the chain, added to \"service-addr\"")
		proposalsUrl = flag.String("proposals", "/api/v0/proposals", "Endpoint (or file path) containing proposals, added to \"service-addr\"")
		fundsUrl     = flag.String("funds", "/api/v0/fund", "Endpoint (or file path) containing fund info, added to \"service-addr\"")
		timeout      = flag.String("http-timeout", "10s", "Http request timeout")
	)

	flag.Parse()

	timeoutDur, err := time.ParseDuration(*timeout)
	kit.FatalOn(err, "http-timeout")
	client.Timeout = timeoutDur

	vpUrl, err := url.ParseRequestURI(*serviceUrl + *votePlansUrl)
	kit.FatalOn(err, "url.ParseRequestURI: "+*votePlansUrl)

	prUrl, err := url.ParseRequestURI(*serviceUrl + *proposalsUrl)
	kit.FatalOn(err, "url.ParseRequestURI: "+*proposalsUrl)

	fuUrl, err := url.ParseRequestURI(*serviceUrl + *fundsUrl)
	kit.FatalOn(err, "url.ParseRequestURI: "+*fundsUrl)

	vps, err := getVotePlans(&client, vpUrl)
	kit.FatalOn(err, "getVotePlans")
	dt, err := json.MarshalIndent(vps, "", "  ")
	kit.FatalOn(err, "getVotePlans - MarshalIndent")
	fmt.Printf("%s\n", dt)

	props, err := getProposals(&client, prUrl)
	kit.FatalOn(err, "getProposals")
	dtp, err := json.MarshalIndent(props, "", "  ")
	kit.FatalOn(err, "getProposals - MarshalIndent")
	_ = dtp //fmt.Printf("%s\n", dtp)

	fund, err := getFunds(&client, fuUrl)
	kit.FatalOn(err, "getFunds")
	dtf, err := json.MarshalIndent(fund, "", "  ")
	kit.FatalOn(err, "getFunds - MarshalIndent")
	_ = dtf // fmt.Printf("%s\n", dtf)
}
