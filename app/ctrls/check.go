package ctrls

import (
	"encoding/hex"
	"encoding/json"

	"github.com/mragiadakos/tendermoney/app/ctrls/models"
	"github.com/mragiadakos/tendermoney/app/ctrls/validations"
	"github.com/tendermint/abci/types"
)

func (app *TMApplication) CheckTx(tx []byte) types.ResponseCheckTx {
	dts := models.Delivery{}
	err := json.Unmarshal(tx, &dts)
	if err != nil {
		return types.ResponseCheckTx{Code: models.CodeTypeEncodingError, Log: "The delivery is not correct json: " + err.Error()}
	}
	sigB, err := hex.DecodeString(dts.Signature)
	if err != nil {
		return types.ResponseCheckTx{Code: models.CodeTypeEncodingError, Log: "The signature is not correct hex: " + err.Error()}
	}
	switch dts.GetType() {
	case models.INFLATE:
		id := dts.GetInflationData()
		code, err := validations.ValidateInflation(id, sigB)
		if err != nil {
			return types.ResponseCheckTx{Code: code, Log: err.Error()}
		}

	case models.SUM:
		sd := dts.GetSumData()
		code, err := validations.ValidateSum(&app.state, sd, sigB)
		if err != nil {
			return types.ResponseCheckTx{Code: code, Log: err.Error()}
		}

	case models.DIVIDE:
		dd := dts.GetDivitionData()
		code, err := validations.ValidateDivition(&app.state, dd, sigB)
		if err != nil {
			return types.ResponseCheckTx{Code: code, Log: err.Error()}
		}

	case models.TAX:
		td := dts.GetTaxData()
		code, err := validations.ValidateTax(&app.state, td, sigB)
		if err != nil {
			return types.ResponseCheckTx{Code: code, Log: err.Error()}
		}
	case models.SEND:
		sd := dts.GetSendData()
		code, err := validations.ValidateSend(&app.state, sd, sigB)
		if err != nil {
			return types.ResponseCheckTx{Code: code, Log: err.Error()}
		}

	case models.RECEIVE:
		rd := dts.GetReceiveData()
		code, err := validations.ValidateReceive(&app.state, rd, sigB)
		if err != nil {
			return types.ResponseCheckTx{Code: code, Log: err.Error()}
		}

	case models.RETRIEVE_FEE:
		rd := dts.GetRetrieveData()
		code, err := validations.ValidateRetrieve(&app.state, rd, sigB)
		if err != nil {
			return types.ResponseCheckTx{Code: code, Log: err.Error()}
		}

	default:
		return types.ResponseCheckTx{Code: models.CodeTypeUnauthorized, Log: "This type of action does not exists."}

	}
	return types.ResponseCheckTx{Code: models.CodeTypeOK}
}
