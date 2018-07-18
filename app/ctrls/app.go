package ctrls

import (
	"github.com/mragiadakos/tendermoney/app/ctrls/dbpkg"
	"github.com/tendermint/abci/types"
	dbm "github.com/tendermint/tmlibs/db"
)

var _ types.Application = (*TMApplication)(nil)

type TMApplication struct {
	types.BaseApplication
	state dbpkg.State
}

func NewTMApplication() *TMApplication {
	state := dbpkg.LoadState(dbm.NewMemDB())
	return &TMApplication{state: state}
}
