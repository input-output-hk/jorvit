package loader

import (
	"io"

	"github.com/gocarina/gocsv"
)

type Proposal struct {
	ID          string `json:"proposal_id"         csv:"proposal_id"`
	Title       string `json:"proposal_title"      csv:"proposal_title"`
	Summary     string `json:"proposal_summary"    csv:"proposal_summary"`
	Problem     string `json:"proposal_problem"    csv:"proposal_problem"`
	Solution    string `json:"proposal_solution"   csv:"proposal_solution"`
	ProposalURL string `json:"proposal_url"        csv:"proposal_url"`
	DataURL     string `json:"proposal_files_url"  csv:"proposal_files_url"`
	PublicKey   string `json:"proposal_public_key" csv:"proposal_public_key"`
	Funds       uint64 `json:"proposal_funds"      csv:"proposal_funds"`
}

type ProposalCategory struct {
	CategoryID   string `json:"category_id"          csv:"category_id"`
	CategoryName string `json:"category_name"        csv:"category_name"`
	CategoryDesc string `json:"category_description" csv:"category_description"`
}

type Proposer struct {
	ProposerEmail string `json:"proposer_email" csv:"proposer_email"`
	ProposerName  string `json:"proposer_name"  csv:"proposer_name"`
	ProposerURL   string `json:"proposer_url"   csv:"proposer_url"`
}

type ChainProposal struct {
	ExternalID  string           `json:"chain_proposal_id"`
	Index       uint8            `json:"chain_proposal_index"`
	VoteOptions ChainVoteOptions `json:"chain_vote_options"`
}

type ChainVoteOptions map[string]uint8

type ChainVotePlan struct {
	VotePlanID   string `json:"chain_voteplan_id"`
	VoteStart    string `json:"chain_vote_starttime"`
	VoteEnd      string `json:"chain_vote_endtime"`
	CommitteeEnd string `json:"chain_committee_endtime"`
	Payload      string `json:"chain_voteplan_payload" csv:"chain_voteplan_payload"`
}

type ProposalData struct {
	InternalID       string `json:"internal_id" csv:"internal_id"`
	ProposalCategory `json:"category"`
	Proposal
	Proposer      `json:"proposer"`
	ChainProposal // `json:"chain_proposal"`
	ChainVotePlan // `json:"chain_voteplan"`
}

func LoadData(r io.Reader) (*[]*ProposalData, error) {
	proposals := make([]*ProposalData, 0)
	err := gocsv.Unmarshal(r, &proposals)
	return &proposals, err
}

type FundData struct {
	Name            string          `json:"fund_name"            csv:"fund_name"`
	Goal            string          `json:"fund_goal"            csv:"fund_goal"`
	VotingPowerInfo string          `json:"voting_power_info"    csv:"voting_power_info"`
	RewardsInfo     string          `json:"rewards_info"         csv:"rewards_info"`
	StartTime       string          `json:"fund_start_time"      csv:"fund_start_time"`
	EndTime         string          `json:"fund_end_time"        csv:"fund_end_time"`
	NextStartTime   string          `json:"next_fund_start_time" csv:"next_fund_start_time"`
	Voteplans       []ChainVotePlan `json:"chain_vote_plans"`
}

func LoadFundData(r io.Reader) (*[]*FundData, error) {
	funds := make([]*FundData, 0)
	err := gocsv.Unmarshal(r, &funds)
	return &funds, err
}
