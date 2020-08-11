package datastore

import "github.com/input-output-hk/jorvit/internal/loader"

type ProposalsStore interface {
	Initialize(filename string) error
	All() *[]*loader.ProposalData
	SearchID(internalID string) *loader.ProposalData
	Total() int
	Payloads() map[string]int
}

type FundsStore interface {
	Initialize(filename string) error
	First() *loader.FundData
	Total() int
}
