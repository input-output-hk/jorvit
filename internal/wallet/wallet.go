package wallet

import (
	"fmt"

	"github.com/rinor/jorcli/jnode"
)

type Wallet struct {
	Note      string              `json:"note"`
	Mnemonics string              `json:"mnemonics"`
	Funds     []jnode.InitialFund `json:"funds"`
	Totals    uint64              `json:"-"`
}

func (w Wallet) String() string {
	return fmt.Sprintf("\nW: %s\nM: %s\nA: %d",
		w.Note,
		w.Mnemonics,
		w.Totals,
	)
}
