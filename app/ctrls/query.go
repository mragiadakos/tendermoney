package ctrls

import (
	"encoding/json"
	"errors"
	"net/url"

	"github.com/mragiadakos/tendermoney/server/ctrls/models"
	"github.com/mragiadakos/tendermoney/server/ctrls/query"
	"github.com/tendermint/abci/types"
)

var (
	ERR_THE_QUERY_METHOD_HAS_NOT_BEEN_FOUND = errors.New("The query's method has not been found.")
)

func (tva *TMApplication) Query(qreq types.RequestQuery) types.ResponseQuery {
	u, err := url.Parse(qreq.Path)
	if err != nil {
		return types.ResponseQuery{Code: models.CodeTypeEncodingError, Log: err.Error()}
	}

	if u.Path == "get_coin" {
		qr, err := query.GetCoin(&tva.state, u)
		if err != nil {
			return types.ResponseQuery{Code: models.CodeTypeUnauthorized, Log: err.Error()}
		}

		b, _ := json.Marshal(qr)
		return types.ResponseQuery{Code: models.CodeTypeOK, Value: b}
	} else if u.Path == "get_coin_by_owner" {
		qr, err := query.GetCoinByOwner(&tva.state, u)
		if err != nil {
			return types.ResponseQuery{Code: models.CodeTypeUnauthorized, Log: err.Error()}
		}

		b, _ := json.Marshal(qr)
		return types.ResponseQuery{Code: models.CodeTypeOK, Value: b}
	} else if u.Path == "get_latest_tax" {
		qt, err := query.GetLatestTax(&tva.state)
		if err != nil {
			return types.ResponseQuery{Code: models.CodeTypeUnauthorized, Log: err.Error()}
		}

		b, _ := json.Marshal(qt)
		return types.ResponseQuery{Code: models.CodeTypeOK, Value: b}
	} else if u.Path == "get_transaction" {
		qt, err := query.GetTransaction(&tva.state, u)
		if err != nil {
			return types.ResponseQuery{Code: models.CodeTypeUnauthorized, Log: err.Error()}
		}
		b, _ := json.Marshal(qt)
		return types.ResponseQuery{Code: models.CodeTypeOK, Value: b}

	} else if u.Path == "get_transactions_with_unreceived_fee" {
		qts := query.GetTransactionsWithUnreceivedFee(&tva.state)
		b, _ := json.Marshal(qts)
		return types.ResponseQuery{Code: models.CodeTypeOK, Value: b}

	}
	return types.ResponseQuery{Code: models.CodeTypeUnauthorized,
		Log: ERR_THE_QUERY_METHOD_HAS_NOT_BEEN_FOUND.Error()}

}
