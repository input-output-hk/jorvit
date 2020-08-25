package loader

import (
	"io"
	"strings"

	"github.com/gocarina/gocsv"
)

type Proposal struct {
	ID          string  `json:"proposal_id"           csv:"proposal_id"`
	Title       string  `json:"proposal_title"        csv:"proposal_title"`
	Summary     string  `json:"proposal_summary"      csv:"proposal_summary"`
	Problem     string  `json:"proposal_problem"      csv:"proposal_problem"`
	Solution    string  `json:"proposal_solution"     csv:"proposal_solution"`
	ProposalURL string  `json:"proposal_url"          csv:"proposal_url"`
	DataURL     string  `json:"proposal_files_url"    csv:"proposal_files_url"`
	PublicKey   string  `json:"proposal_public_key"   csv:"proposal_public_key"`
	Funds       uint64  `json:"proposal_funds"        csv:"proposal_funds"`
	ImpactScore float32 `json:"proposal_impact_score" csv:"proposal_impact_score"`
}

type ProposalCategory struct {
	CategoryID   string `json:"category_id"          csv:"-"`
	CategoryName string `json:"category_name"        csv:"category_name"`
	CategoryDesc string `json:"category_description" csv:"-"`
}

type Proposer struct {
	ProposerEmail      string `json:"proposer_email"               csv:"proposer_email"`
	ProposerName       string `json:"proposer_name"                csv:"proposer_name"`
	ProposerURL        string `json:"proposer_url"                 csv:"proposer_url"`
	ProposerExperience string `json:"proposer_relevant_experience" csv:"proposer_relevant_experience"`
}

type ChainProposal struct {
	ExternalID  string           `json:"chain_proposal_id"    csv:"chain_proposal_id"`
	Index       uint8            `json:"chain_proposal_index" csv:"chain_proposal_index"`
	VoteOptions ChainVoteOptions `json:"chain_vote_options"   csv:"chain_vote_options"`
	VoteType    string           `json:"-"                    csv:"chain_vote_type"`
	VoteAction  string           `json:"-"                    csv:"chain_vote_action"`
}

type ChainVoteOptions map[string]uint8

func (cvo *ChainVoteOptions) UnmarshalCSV(csv string) error {
	options := strings.Split(csv, ",")
	*cvo = make(map[string]uint8, len(options))
	for i, opt := range options {
		(*cvo)[opt] = uint8(i)
	}
	return nil
}
func (cvo *ChainVoteOptions) MarshalCSV() (string, error) {
	opts := make([]string, len(*cvo))
	for opt, i := range *cvo {
		opts[i] = opt
	}
	return strings.Join(opts, ","), nil
}

/*
type DateTime struct{ time.Time }

// Convert the internal date as CSV string
func (date *DateTime) MarshalCSV() (int64, error) {
	return date.Time.Unix(), nil
}

// Convert the CSV string as internal date
func (date *DateTime) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse(time.RFC3339, csv)
	return err
}
*/

type ChainVotePlan struct {
	VpInternalID string `json:"-"                        csv:"id"`
	VotePlanID   string `json:"chain_voteplan_id"        csv:"chain_voteplan_id"`
	VoteStart    string `json:"chain_vote_start_time"    csv:"chain_vote_start_time"`
	VoteEnd      string `json:"chain_vote_end_time"      csv:"chain_vote_end_time"`
	CommitteeEnd string `json:"chain_committee_end_time" csv:"chain_committee_end_time"`
	Payload      string `json:"chain_voteplan_payload"   csv:"chain_voteplan_payload"`
	FundID       string `json:"fund_id"                  csv:"fund_id"`
}

type ProposalData struct {
	InternalID       string `json:"internal_id" csv:"internal_id"`
	ProposalCategory `json:"proposals_category"`
	Proposal         //
	Proposer         `json:"proposer"`
	ChainProposal
	*ChainVotePlan
}

func LoadData(r io.Reader) (*[]*ProposalData, error) {
	proposals := make([]*ProposalData, 0)
	err := gocsv.Unmarshal(r, &proposals)
	return &proposals, err
}

type FundData struct {
	FundID          string          `json:"id,omitempty"         csv:"id"`
	Name            string          `json:"fund_name"            csv:"fund_name"`
	Goal            string          `json:"fund_goal"            csv:"fund_goal"`
	VotingPowerInfo string          `json:"voting_power_info"    csv:"voting_power_info"`
	RewardsInfo     string          `json:"rewards_info"         csv:"rewards_info"`
	StartTime       string          `json:"fund_start_time"      csv:"fund_start_time"`
	EndTime         string          `json:"fund_end_time"        csv:"fund_end_time"`
	NextStartTime   string          `json:"next_fund_start_time" csv:"next_fund_start_time"`
	VotePlans       []ChainVotePlan `json:"chain_vote_plans"     csv:"-"`
}

func LoadFundData(r io.Reader) (*[]*FundData, error) {
	funds := make([]*FundData, 0)
	err := gocsv.Unmarshal(r, &funds)
	return &funds, err
}
