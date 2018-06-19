package ctrls

import "github.com/tendermint/abci/types"

func (app *TMApplication) CheckTx(tx []byte) types.ResponseCheckTx {

	return types.ResponseCheckTx{Code: CodeTypeOK}
}
