package ctrls

import (
	"encoding/hex"
	"encoding/json"

	"github.com/mragiadakos/tendermoney/server/ctrls/dbpkg"

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
		sc := dbpkg.StateCoin{}
		sc.Coin = id.Coin
		sc.Owner = id.Owner
		sc.Value = id.Value
		err = app.state.AddCoin(sc)
		if err != nil {
			return types.ResponseDeliverTx{Code: models.CodeTypeUnauthorized, Log: err.Error()}
		}
	case models.SUM:
		sd := dts.GetSumData()
		code, err := validations.ValidateSum(&app.state, sd, sigB)
		if err != nil {
			return types.ResponseDeliverTx{Code: code, Log: err.Error()}
		}
		sum := 0.0
		for _, v := range sd.Coins {
			sc, err := app.state.GetCoin(v)
			if err != nil {
				return types.ResponseDeliverTx{Code: models.CodeTypeUnauthorized, Log: err.Error()}
			}
			sum += sc.Value
			app.state.DeleteCoinAndOwner(v)
		}

		sc := dbpkg.StateCoin{}
		sc.Coin = sd.NewCoin
		sc.Owner = sd.NewOwner
		sc.Value = sum
		err = app.state.AddCoin(sc)
		if err != nil {
			return types.ResponseDeliverTx{Code: models.CodeTypeUnauthorized, Log: err.Error()}
		}
	}
	return types.ResponseDeliverTx{Code: models.CodeTypeOK}
}
