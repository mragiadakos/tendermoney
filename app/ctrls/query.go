package ctrls

import (
	"encoding/json"
	"errors"
	"net/url"

	"github.com/mragiadakos/tendermoney/app/ctrls/models"
	"github.com/mragiadakos/tendermoney/app/ctrls/query"
	"github.com/tendermint/abci/types"
)

var (
	ERR_THE_QUERY_METHOD_HAS_NOT_BEEN_FOUND = errors.New("The query's method has not been found.")
)

const (
	QUERY_GET_COIN                            = "get_coin"
	QUERY_GET_COIN_BY_OWNER                   = "get_coin_by_owner"
	QUERY_GET_LATEST_TAX                      = "get_latest_tax"
	QUERY_GET_TRANSACTION                     = "get_transaction"
	QUERY_GET_TRANSACTION_WITH_UNRECEIVED_FEE = "get_transactions_with_unreceived_fee"
)

func (tva *TMApplication) Query(qreq types.RequestQuery) types.ResponseQuery {
	u, err := url.Parse(qreq.Path)
	if err != nil {
		return types.ResponseQuery{Code: models.CodeTypeEncodingError, Log: err.Error()}
	}

	switch u.Path {
	case QUERY_GET_COIN:
		qr, err := query.GetCoin(&tva.state, u)
		if err != nil {
			return types.ResponseQuery{Code: models.CodeTypeUnauthorized, Log: err.Error()}
		}

		b, _ := json.Marshal(qr)
		return types.ResponseQuery{Code: models.CodeTypeOK, Value: b}
	case QUERY_GET_COIN_BY_OWNER:
		qr, err := query.GetCoinByOwner(&tva.state, u)
		if err != nil {
			return types.ResponseQuery{Code: models.CodeTypeUnauthorized, Log: err.Error()}
		}

		b, _ := json.Marshal(qr)
		return types.ResponseQuery{Code: models.CodeTypeOK, Value: b}
	case QUERY_GET_LATEST_TAX:
		qt, err := query.GetLatestTax(&tva.state)
		if err != nil {
			return types.ResponseQuery{Code: models.CodeTypeUnauthorized, Log: err.Error()}
		}

		b, _ := json.Marshal(qt)
		return types.ResponseQuery{Code: models.CodeTypeOK, Value: b}
	case QUERY_GET_TRANSACTION:
		qt, err := query.GetTransaction(&tva.state, u)
		if err != nil {
			return types.ResponseQuery{Code: models.CodeTypeUnauthorized, Log: err.Error()}
		}
		b, _ := json.Marshal(qt)
		return types.ResponseQuery{Code: models.CodeTypeOK, Value: b}

	case QUERY_GET_TRANSACTION_WITH_UNRECEIVED_FEE:
		qts := query.GetTransactionsWithUnreceivedFee(&tva.state)
		b, _ := json.Marshal(qts)
		return types.ResponseQuery{Code: models.CodeTypeOK, Value: b}
	}
	return types.ResponseQuery{Code: models.CodeTypeUnauthorized,
		Log: ERR_THE_QUERY_METHOD_HAS_NOT_BEEN_FOUND.Error()}

}
