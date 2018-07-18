package validations

import (
	"encoding/json"
	"errors"

	"github.com/mragiadakos/tendermoney/server/ctrls/utils"

	"github.com/mragiadakos/tendermoney/server/confs"
	"github.com/mragiadakos/tendermoney/server/ctrls/dbpkg"
	"github.com/mragiadakos/tendermoney/server/ctrls/models"
)

var (
	ERR_NEW_OWNERS_NOT_EQUAL_TO_FEES   = errors.New("The list of new owners is not equal to fees.")
	ERR_TRANSACTION_DOES_NOT_HAVE_FEE  = errors.New("The transaction does not have fee.")
	ERR_COIN_IS_NOT_IN_FEE_TRANSACTION = func(uuid string) error {
		return errors.New("The coin " + uuid + " is not in the fee of the transaction.")
	}
	ERR_TRANSACTION_HAS_BEEN_RETRIEVED = errors.New("The fees from the transaction has already been received.")
)

func ValidateRetrieve(state *dbpkg.State, rd models.RetrieveData, sig []byte) (uint32, error) {
	tr, err := state.GetTransaction(rd.TransactionHash)
	if err != nil {
		return models.CodeTypeUnauthorized, ERR_TRANSACTION_HASH_DOES_NOT_EXIST
	}
	if len(tr.Fee) == 0 {
		return models.CodeTypeUnauthorized, ERR_TRANSACTION_DOES_NOT_HAVE_FEE
	}
	if len(rd.NewOwners) != len(tr.Fee) {
		return models.CodeTypeUnauthorized, ERR_NEW_OWNERS_NOT_EQUAL_TO_FEES
	}
	if len(rd.Inflator) == 0 {
		return models.CodeTypeUnauthorized, ERR_INFLATOR_EMPTY
	}

	inList := false
	for _, v := range confs.Conf.Inflators {
		if v == rd.Inflator {
			inList = true
			break
		}
	}
	if !inList {
		return models.CodeTypeUnauthorized, ERR_INFLATOR_NOT_IN_LIST
	}

	allPubs := []string{rd.Inflator}
	for coin, owner := range rd.NewOwners {
		_, err := state.GetOwner(owner)
		if err == nil {
			return models.CodeTypeUnauthorized, ERR_OWNER_FROM_COINS_EXISTS_ALREADY(owner, coin)
		}

		isFoundCoin := false
		for _, trCoin := range tr.Fee {
			if coin == trCoin {
				isFoundCoin = true
				break
			}
		}
		if !isFoundCoin {
			return models.CodeTypeUnauthorized, ERR_COIN_IS_NOT_IN_FEE_TRANSACTION(coin)
		}
		allPubs = append(allPubs, owner)
	}
	msg, _ := json.Marshal(rd)
	isValid, err := utils.MultiVerify(allPubs, sig, msg)
	if err != nil {
		return models.CodeTypeUnauthorized, err
	}

	if !isValid {
		return models.CodeTypeUnauthorized, ERR_SIGNATURE_NOT_VALID
	}

	if tr.IsFeeReceived {
		return models.CodeTypeUnauthorized, ERR_TRANSACTION_HAS_BEEN_RETRIEVED
	}
	return models.CodeTypeOK, nil
}
