package validations

import (
	"encoding/json"
	"errors"

	"github.com/dedis/kyber/group/edwards25519"
	"github.com/dedis/kyber/sign/schnorr"
	"github.com/mragiadakos/tendermoney/server/confs"
	"github.com/mragiadakos/tendermoney/server/ctrls/models"
	"github.com/mragiadakos/tendermoney/server/ctrls/utils"
)

var (
	ERR_COIN_EMPTY           = errors.New("The coin is empty.")
	ERR_SIGNATURE_EMPTY      = errors.New("The signature is empty.")
	ERR_OWNER_EMPTY          = errors.New("The owner's public key is empty.")
	ERR_INFLATOR_EMPTY       = errors.New("The inflator's public key is empty.")
	ERR_INFLATOR_NOT_IN_LIST = errors.New("The inflator not in the list of inflators.")
	ERR_VALUE_NOT_IN_LIST    = errors.New("The value not in the list of constant values.")
	ERR_SIGNATURE_NOT_VALID  = errors.New("The public keys do not validate the signature.")
)

func ValidateInflation(id models.InflationData, sig []byte) (uint32, error) {
	if len(id.Coin) == 0 {
		return models.CodeTypeUnauthorized, ERR_COIN_EMPTY
	}
	if len(sig) == 0 {
		return models.CodeTypeUnauthorized, ERR_SIGNATURE_EMPTY
	}
	if len(id.Owner) == 0 {
		return models.CodeTypeUnauthorized, ERR_OWNER_EMPTY
	}

	if len(id.Inflator) == 0 {
		return models.CodeTypeUnauthorized, ERR_INFLATOR_EMPTY
	}

	inList := false
	for _, v := range confs.Conf.Inflators {
		if v == id.Inflator {
			inList = true
			break
		}
	}
	if !inList {
		return models.CodeTypeUnauthorized, ERR_INFLATOR_NOT_IN_LIST
	}

	inList = false
	for _, v := range models.CONSTANT_VALUES {
		if v == id.Value {
			inList = true
			break
		}
	}
	if !inList {
		return models.CodeTypeUnauthorized, ERR_VALUE_NOT_IN_LIST
	}

	suite := edwards25519.NewBlakeSHA256Ed25519()
	pubOwner, err := utils.UnmarshalPublicKey(id.Owner)
	if err != nil {
		return models.CodeTypeUnauthorized, err
	}
	pubInflator, err := utils.UnmarshalPublicKey(id.Inflator)
	if err != nil {
		return models.CodeTypeUnauthorized, err
	}
	onePublic := suite.Point().Add(pubInflator, pubOwner)
	msg, _ := json.Marshal(id)
	err = schnorr.Verify(suite, onePublic, msg, sig)
	if err != nil {
		return models.CodeTypeUnauthorized, ERR_SIGNATURE_NOT_VALID
	}
	return models.CodeTypeOK, nil
}
