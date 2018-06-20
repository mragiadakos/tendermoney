package ctrls

import (
	"encoding/hex"
	"encoding/json"

	"github.com/mragiadakos/tendermoney/server/ctrls/models"
	"github.com/mragiadakos/tendermoney/server/ctrls/validations"
	"github.com/tendermint/abci/types"
)

func (app *TMApplication) DeliverTx(tx []byte) types.ResponseDeliverTx {
	dts := models.Delivery{}
	err := json.Unmarshal(tx, &dts)
	if err != nil {
		return types.ResponseDeliverTx{Code: models.CodeTypeEncodingError, Log: "The delivery is not correct json: " + err.Error()}
	}
	sigB, err := hex.DecodeString(dts.Signature)
	if err != nil {
		return types.ResponseDeliverTx{Code: models.CodeTypeEncodingError, Log: "The signature is not correct hex: " + err.Error()}
	}
	switch dts.GetType() {
	case models.INFLATE:
		id := dts.GetInflationData()
		code, err := validations.ValidateInflation(id, sigB)
		if err != nil {
			return types.ResponseDeliverTx{Code: code, Log: err.Error()}
		}
		err = app.state.AddCoin(&id)
		if err != nil {
			return types.ResponseDeliverTx{Code: models.CodeTypeUnauthorized, Log: err.Error()}
		}
	}
	return types.ResponseDeliverTx{Code: models.CodeTypeOK}
}
