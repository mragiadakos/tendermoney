package ctrls

import "github.com/tendermint/abci/types"

func (tva *TMApplication) Query(qreq types.RequestQuery) types.ResponseQuery {

	resp := types.ResponseQuery{Code: CodeTypeOK}
	return resp
}
