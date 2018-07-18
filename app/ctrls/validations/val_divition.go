package validations

import (
	"encoding/json"
	"errors"

	"github.com/mragiadakos/tendermoney/app/ctrls/dbpkg"
	"github.com/mragiadakos/tendermoney/app/ctrls/models"
	"github.com/mragiadakos/tendermoney/app/ctrls/utils"
)

var (
	ERR_NEW_COINS_EMPTY      = errors.New("The new coins are empty.")
	ERR_NEW_COIN_OWNER_EMPTY = func(uuid string) error {
		return errors.New("The owner of the coin " + uuid + " is empty.")
	}
	ERR_NEW_COIN_NON_CONSTANT_VALUE = func(uuid string) error {
		return errors.New("From the new coins, the coin " + uuid + " does not have a constant value.")
	}
	ERR_NEW_COINS_NOT_EQUAL_TO_SUM         = errors.New("The sum of the new coins is not equal to the old coin.")
	ERR_NEW_COINS_EQUAL_OWNER              = errors.New("The new coins owners are equal.")
	ERR_COIN_DOES_NOT_EXISTS               = errors.New("The coin does not exists.")
	ERR_NEW_COINS_IS_NOT_EQUAL_TO_THE_COIN = errors.New("The new coins is not equal to the coin.")
	ERR_COIN_FROM_NEW_COINS_EXISTS_ALREADY = func(uuid string) error {
		return errors.New("The coin " + uuid + " from new coins, exists already.")
	}
	ERR_OWNER_FROM_NEW_COINS_EXISTS_ALREADY = func(pub string) error {
		return errors.New("The owner " + pub + " from new coins, exists already.")
	}
)

func ValidateDivition(s *dbpkg.State, dd models.DivitionData, sig []byte) (uint32, error) {
	if len(dd.Coin) == 0 {
		return models.CodeTypeUnauthorized, ERR_COIN_EMPTY
	}

	if len(dd.NewCoins) == 0 {
		return models.CodeTypeUnauthorized, ERR_NEW_COINS_EMPTY
	}
	checkOwners := map[string]string{}
	sum := 0.0
	ownerPubs := []string{}
	for k, coin := range dd.NewCoins {
		if len(coin.Owner) == 0 {
			return models.CodeTypeUnauthorized, ERR_NEW_COIN_OWNER_EMPTY(k)
		}

		isFound := false
		for _, v := range models.CONSTANT_VALUES {
			if v == coin.Value {
				isFound = true
				break
			}
		}
		if !isFound {
			return models.CodeTypeUnauthorized, ERR_NEW_COIN_NON_CONSTANT_VALUE(k)
		}
		_, ok := checkOwners[coin.Owner]
		if ok {
			return models.CodeTypeUnauthorized, ERR_NEW_COINS_EQUAL_OWNER
		} else {
			checkOwners[coin.Owner] = coin.Owner
		}
		_, err := s.GetCoin(k)
		if err == nil {
			return models.CodeTypeUnauthorized, ERR_COIN_FROM_NEW_COINS_EXISTS_ALREADY(k)
		}
		_, err = s.GetOwner(coin.Owner)
		if err == nil {
			return models.CodeTypeUnauthorized, ERR_OWNER_FROM_NEW_COINS_EXISTS_ALREADY(coin.Owner)
		}
		sum += coin.Value
		ownerPubs = append(ownerPubs, coin.Owner)
	}

	sc, err := s.GetCoin(dd.Coin)
	if err != nil {
		return models.CodeTypeUnauthorized, ERR_COIN_DOES_NOT_EXISTS
	}

	if sum != sc.Value {
		return models.CodeTypeUnauthorized, ERR_NEW_COINS_IS_NOT_EQUAL_TO_THE_COIN
	}
	ownerPubs = append(ownerPubs, sc.Owner)
	msg, _ := json.Marshal(dd)
	isValid, err := utils.MultiVerify(ownerPubs, sig, msg)
	if err != nil {
		return models.CodeTypeUnauthorized, err
	}
	if !isValid {
		return models.CodeTypeUnauthorized, ERR_SIGNATURE_NOT_VALID
	}

	isLocked, err := s.IsCoinLocked(dd.Coin)
	if err != nil {
		return models.CodeTypeUnauthorized, err
	}
	if isLocked {
		return models.CodeTypeUnauthorized, ERR_COIN_IS_LOCKED(dd.Coin)
	}

	return models.CodeTypeOK, nil
}
