package ctrls

import (
	"github.com/mragiadakos/tendermoney/server/ctrls/models"
	"github.com/tendermint/abci/types"
)

func (tva *TMApplication) Query(qreq types.RequestQuery) types.ResponseQuery {

	resp := types.ResponseQuery{Code: models.CodeTypeOK}
	return resp
}
