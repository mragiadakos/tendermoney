package validations

import (
	"encoding/json"
	"errors"

	"github.com/mragiadakos/tendermoney/server/ctrls/utils"

	"github.com/mragiadakos/tendermoney/server/ctrls/dbpkg"
	"github.com/mragiadakos/tendermoney/server/ctrls/models"
)

var (
	ERR_COINS_EMPTY               = errors.New("The coins are empty.")
	ERR_NEW_COIN_EMPTY            = errors.New("The new coin is empty.")
	ERR_NEW_OWNER_EMPTY           = errors.New("The new owner is empty.")
	ERR_SUM_OF_COINS_NON_CONSTANT = errors.New("The sum of coins is not constant.")
	ERR_NEW_COIN_EXISTS_ALREADY   = errors.New("The new coin exists already.")
	ERR_NEW_OWNER_EXISTS_ALREADY  = errors.New("The new owner exists already.")
)

func ValidateSum(s *dbpkg.State, sd models.SumData, sig []byte) (uint32, error) {
	if len(sd.Coins) == 0 {
		return models.CodeTypeUnauthorized, ERR_COINS_EMPTY
	}
	if len(sig) == 0 {
		return models.CodeTypeUnauthorized, ERR_SIGNATURE_EMPTY
	}
	if len(sd.NewCoin) == 0 {
		return models.CodeTypeUnauthorized, ERR_NEW_COIN_EMPTY
	}
	if len(sd.NewOwner) == 0 {
		return models.CodeTypeUnauthorized, ERR_NEW_OWNER_EMPTY
	}

	// check if the value is in the list of constants
	var sum float64 = 0
	ownersPubs := []string{sd.NewOwner}
	for _, v := range sd.Coins {
		sc, err := s.GetCoin(v)
		if err != nil {
			return models.CodeTypeUnauthorized, err
		}
		sum += sc.Value
		ownersPubs = append(ownersPubs, sc.Owner)
	}

	isFound := false
	for _, v := range models.CONSTANT_VALUES {
		if v == sum {
			isFound = true
			break
		}
	}
	if !isFound {
		return models.CodeTypeUnauthorized, ERR_SUM_OF_COINS_NON_CONSTANT
	}

	// check if the new coin exists already
	_, err := s.GetCoin(sd.NewCoin)
	if err == nil {
		return models.CodeTypeUnauthorized, ERR_NEW_COIN_EXISTS_ALREADY
	}

	// check if the new owner exists already
	_, err = s.GetOwner(sd.NewOwner)
	if err == nil {
		return models.CodeTypeUnauthorized, ERR_NEW_OWNER_EXISTS_ALREADY
	}
	msg, _ := json.Marshal(sd)
	isValid, err := utils.MultiVerify(ownersPubs, sig, msg)
	if err != nil {
		return models.CodeTypeUnauthorized, err
	}
	if !isValid {
		return models.CodeTypeUnauthorized, ERR_SIGNATURE_NOT_VALIDATE
	}
	return models.CodeTypeOK, nil
}
