package ctrls

import (
	"encoding/hex"
	"encoding/json"

	"github.com/mragiadakos/tendermoney/app/ctrls/dbpkg"

	"github.com/mragiadakos/tendermoney/app/ctrls/models"
	"github.com/mragiadakos/tendermoney/app/ctrls/validations"
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
	case models.DIVIDE:
		dd := dts.GetDivitionData()
		code, err := validations.ValidateDivition(&app.state, dd, sigB)
		if err != nil {
			return types.ResponseDeliverTx{Code: code, Log: err.Error()}
		}
		app.state.DeleteCoinAndOwner(dd.Coin)
		for k, v := range dd.NewCoins {
			sc := dbpkg.StateCoin{}
			sc.Coin = k
			sc.Owner = v.Owner
			sc.Value = v.Value
			err = app.state.AddCoin(sc)
			if err != nil {
				return types.ResponseDeliverTx{Code: models.CodeTypeUnauthorized, Log: err.Error()}
			}
		}
	case models.TAX:
		td := dts.GetTaxData()
		code, err := validations.ValidateTax(&app.state, td, sigB)
		if err != nil {
			return types.ResponseDeliverTx{Code: code, Log: err.Error()}
		}
		app.state.AddTax(td)
	case models.SEND:
		sd := dts.GetSendData()
		code, err := validations.ValidateSend(&app.state, sd, sigB)
		if err != nil {
			return types.ResponseDeliverTx{Code: code, Log: err.Error()}
		}
		app.state.AddTransaction(sd)
		allCoins := append(sd.Coins, sd.Fee...)
		for _, v := range allCoins {
			app.state.LockCoin(v)
		}
	case models.RECEIVE:
		rd := dts.GetReceiveData()
		code, err := validations.ValidateReceive(&app.state, rd, sigB)
		if err != nil {
			return types.ResponseDeliverTx{Code: code, Log: err.Error()}
		}

		for coin, newOwner := range rd.NewOwners {
			err := app.state.UnlockCoin(coin)
			if err != nil {
				return types.ResponseDeliverTx{Code: models.CodeTypeServerError, Log: err.Error()}
			}
			sc, err := app.state.GetCoin(coin)
			if err != nil {
				return types.ResponseDeliverTx{Code: models.CodeTypeServerError, Log: err.Error()}
			}
			err = app.state.DeleteOwner(sc.Owner)
			if err != nil {
				return types.ResponseDeliverTx{Code: models.CodeTypeServerError, Log: err.Error()}
			}
			err = app.state.SetNewOwner(coin, newOwner)
			if err != nil {
				return types.ResponseDeliverTx{Code: models.CodeTypeServerError, Log: err.Error()}
			}
		}
		err = app.state.CoinsReceivedFromTransaction(rd.TransactionHash)
		if err != nil {
			return types.ResponseDeliverTx{Code: models.CodeTypeServerError, Log: err.Error()}
		}

	case models.RETRIEVE_FEE:
		rd := dts.GetRetrieveData()
		code, err := validations.ValidateRetrieve(&app.state, rd, sigB)
		if err != nil {
			return types.ResponseDeliverTx{Code: code, Log: err.Error()}
		}
		err = app.state.FeeRetrievedFromTransaction(rd.TransactionHash)
		if err != nil {
			return types.ResponseDeliverTx{Code: models.CodeTypeServerError, Log: err.Error()}
		}
		for coin, newOwner := range rd.NewOwners {
			err := app.state.UnlockCoin(coin)
			if err != nil {
				return types.ResponseDeliverTx{Code: models.CodeTypeServerError, Log: err.Error()}
			}
			sc, err := app.state.GetCoin(coin)
			if err != nil {
				return types.ResponseDeliverTx{Code: models.CodeTypeServerError, Log: err.Error()}
			}
			err = app.state.DeleteOwner(sc.Owner)
			if err != nil {
				return types.ResponseDeliverTx{Code: models.CodeTypeServerError, Log: err.Error()}
			}
			err = app.state.SetNewOwner(coin, newOwner)
			if err != nil {
				return types.ResponseDeliverTx{Code: models.CodeTypeServerError, Log: err.Error()}
			}
		}

	default:
		return types.ResponseDeliverTx{Code: models.CodeTypeUnauthorized, Log: "This type of action does not exists."}

	}
	return types.ResponseDeliverTx{Code: models.CodeTypeOK}
}
