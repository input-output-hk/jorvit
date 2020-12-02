package loader

import (
	"io"
	"strconv"
	"strings"

	"github.com/gocarina/gocsv"
)

type Proposal struct {
	ID          string   `json:"proposal_id"           csv:"proposal_id"`
	Title       string   `json:"proposal_title"        csv:"proposal_title"`
	Summary     string   `json:"proposal_summary"      csv:"proposal_summary"`
	Problem     string   `json:"proposal_problem"      csv:"proposal_problem"`
	Solution    string   `json:"proposal_solution"     csv:"proposal_solution"`
	ProposalURL string   `json:"proposal_url"          csv:"proposal_url"`
	DataURL     string   `json:"proposal_files_url"    csv:"proposal_files_url"`
	PublicKey   string   `json:"proposal_public_key"   csv:"proposal_public_key"`
	Funds       Lovelace `json:"proposal_funds"        csv:"proposal_funds"`
	ImpactScore Score    `json:"proposal_impact_score" csv:"proposal_impact_score"`
}

type Lovelace uint64

func (lvl *Lovelace) UnmarshalCSV(csv string) error {
	ada, err := strconv.ParseUint(csv, 10, 64)
	if err != nil {
		return err
	}
	*lvl = Lovelace(ada * 1_000_000)
	return nil
}

type Score int

func (sc *Score) UnmarshalCSV(csv string) error {
	f, err := strconv.ParseFloat(csv, 32)
	if err != nil {
		return err
	}
	*sc = Score(f * 100)
	return nil
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

type ChainVotePlan struct {
	VpInternalID      string `json:"-"                        csv:"id"`
	VotePlanID        string `json:"chain_voteplan_id"         csv:"chain_voteplan_id"`
	VoteStart         string `json:"chain_vote_start_time"     csv:"chain_vote_start_time"`
	VoteEnd           string `json:"chain_vote_end_time"       csv:"chain_vote_end_time"`
	CommitteeEnd      string `json:"chain_committee_end_time"  csv:"chain_committee_end_time"`
	Payload           string `json:"chain_voteplan_payload"    csv:"chain_voteplan_payload"`
	VoteEncryptionKey string `json:"chain_vote_encryption_key" csv:"chain_vote_encryption_key"`
	FundID            uint64 `json:"fund_id"                   csv:"fund_id"`
}

type ProposalData struct {
	InternalID       uint64 `json:"internal_id" csv:"internal_id"`
	ProposalCategory `json:"proposal_category"`
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
	FundID               uint64          `json:"id,omitempty"           csv:"id"`
	Name                 string          `json:"fund_name"              csv:"fund_name"`
	VotingPowerThreshold Lovelace        `json:"voting_power_threshold" csv:"voting_power_threshold"`
	Goal                 string          `json:"fund_goal"              csv:"fund_goal"`
	VotingPowerInfo      string          `json:"voting_power_info"      csv:"voting_power_info"`
	RewardsInfo          string          `json:"rewards_info"           csv:"rewards_info"`
	StartTime            string          `json:"fund_start_time"        csv:"fund_start_time"`
	EndTime              string          `json:"fund_end_time"          csv:"fund_end_time"`
	NextStartTime        string          `json:"next_fund_start_time"   csv:"next_fund_start_time"`
	VotePlans            []ChainVotePlan `json:"chain_vote_plans"       csv:"-"`
}

func LoadFundData(r io.Reader) (*[]*FundData, error) {
	funds := make([]*FundData, 0)
	err := gocsv.Unmarshal(r, &funds)
	return &funds, err
}
