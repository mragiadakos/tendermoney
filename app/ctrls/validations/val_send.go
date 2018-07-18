package validations

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mragiadakos/tendermoney/app/ctrls/utils"

	"github.com/mragiadakos/tendermoney/app/ctrls/dbpkg"
	"github.com/mragiadakos/tendermoney/app/ctrls/models"
)

var (
	ERR_FEES_EMPTY                      = errors.New("The list of coins for the fee is empty.")
	ERR_COIN_FROM_COINS_DOES_NOT_EXISTS = func(uuid string) error {
		return errors.New("The coin " + uuid + " from coins, does not exists.")
	}
	ERR_COIN_FROM_FEE_DOES_NOT_EXISTS = func(uuid string) error {
		return errors.New("The coin " + uuid + " from fee, does not exists.")
	}
	ERR_FEE_NOT_BASED_ON_TAX = func(missing float64) error {
		return errors.New(fmt.Sprint("The fee is not based on the tax, it is missing", missing, "."))
	}
	ERR_COIN_FROM_COINS_ADDED_TWICE = func(uuid string) error {
		return errors.New("The coin " + uuid + " from coins added twice.")
	}
	ERR_COIN_FROM_FEE_ADDED_TWICE = func(uuid string) error {
		return errors.New("The coin " + uuid + " from fee added twice.")
	}
	ERR_COIN_ADDED_ON_BOTH_COINS_AND_FEE = func(uuid string) error {
		return errors.New("The coin " + uuid + " added on both coins and fee.")
	}
	ERR_PROOF_NOT_CORRECT = errors.New("The proof is not correct.")
	ERR_COIN_IS_LOCKED    = func(uuid string) error {
		return errors.New("The coin " + uuid + " is locked.")
	}
)

func ValidateSend(s *dbpkg.State, sd models.SendData, sig []byte) (uint32, error) {

	if len(sd.Coins) == 0 {
		return models.CodeTypeUnauthorized, ERR_COINS_EMPTY
	}

	checkCoins := map[string]int{}
	for _, v := range sd.Coins {
		_, ok := checkCoins[v]
		if ok {
			return models.CodeTypeUnauthorized, ERR_COIN_FROM_COINS_ADDED_TWICE(v)
		}
		checkCoins[v] = 0
	}

	tax := s.GetTax()
	if tax.Percentage > 0 {
		if len(sd.Fee) == 0 {
			return models.CodeTypeUnauthorized, ERR_FEES_EMPTY
		}
	}

	checkFees := map[string]int{}
	for _, v := range sd.Fee {
		_, ok := checkFees[v]
		if ok {
			return models.CodeTypeUnauthorized, ERR_COIN_FROM_FEE_ADDED_TWICE(v)
		}
		checkFees[v] = 0
	}

	allCoins := append(sd.Coins, sd.Fee...)
	checkAllCoins := map[string]int{}
	for _, v := range allCoins {
		_, ok := checkAllCoins[v]
		if ok {
			return models.CodeTypeUnauthorized, ERR_COIN_ADDED_ON_BOTH_COINS_AND_FEE(v)
		}
		checkAllCoins[v] = 0
	}

	sumCoins := 0.0
	for _, v := range sd.Coins {
		c, err := s.GetCoin(v)
		if err != nil {
			return models.CodeTypeUnauthorized, ERR_COIN_FROM_COINS_DOES_NOT_EXISTS(v)
		}
		sumCoins += c.Value
	}

	sumFee := 0.0
	for _, v := range sd.Fee {
		f, err := s.GetCoin(v)
		if err != nil {
			return models.CodeTypeUnauthorized, ERR_COIN_FROM_FEE_DOES_NOT_EXISTS(v)
		}
		sumFee += f.Value
	}
	taxFee := tax.GetFeeFromTransaction(sumCoins)
	if taxFee != 0 {
		if taxFee > sumFee {
			sub := utils.ToFixed(taxFee-sumFee, 2)
			return models.CodeTypeUnauthorized, ERR_FEE_NOT_BASED_ON_TAX(sub)
		}
	}

	allPubs := []string{}
	for _, v := range allCoins {
		c, _ := s.GetCoin(v)
		allPubs = append(allPubs, c.Owner)
	}

	msg, _ := json.Marshal(sd)
	isVer, err := utils.MultiVerify(allPubs, sig, msg)
	if err != nil {
		return models.CodeTypeEncodingError, err
	}
	if !isVer {
		return models.CodeTypeUnauthorized, ERR_SIGNATURE_NOT_VALID
	}

	_, err = sd.Proof.GetProof()
	if err != nil {
		return models.CodeTypeUnauthorized, ERR_PROOF_NOT_CORRECT
	}

	for _, v := range allCoins {
		isLocked, err := s.IsCoinLocked(v)
		if err != nil {
			return models.CodeTypeUnauthorized, err
		}
		if isLocked {
			return models.CodeTypeUnauthorized, ERR_COIN_IS_LOCKED(v)
		}
	}
	return models.CodeTypeOK, nil
}
