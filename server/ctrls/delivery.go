package ctrls

import "github.com/tendermint/abci/types"

func (app *TMApplication) DeliverTx(tx []byte) types.ResponseDeliverTx {

	return types.ResponseDeliverTx{Code: CodeTypeOK}
}
