package ctrls

import (
	"github.com/mragiadakos/tendermoney/server/ctrls/models"
	"github.com/tendermint/abci/types"
)

func (app *TMApplication) CheckTx(tx []byte) types.ResponseCheckTx {

	return types.ResponseCheckTx{Code: models.CodeTypeOK}
}
