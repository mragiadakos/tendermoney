package validations

import (
	"encoding/json"
	"errors"

	"github.com/mragiadakos/tendermoney/app/ctrls/utils"

	"github.com/dedis/kyber/group/edwards25519"
	"github.com/mragiadakos/tendermoney/app/ctrls/dbpkg"
	"github.com/mragiadakos/tendermoney/app/ctrls/models"
)

var (
	ERR_TRANSACTION_HASH_EMPTY          = errors.New("The hash of the transaction is empty.")
	ERR_TRANSACTION_HASH_DOES_NOT_EXIST = errors.New("The hash of the transaction does not exists.")
	ERR_COIN_IS_NOT_IN_TRANSACTION      = func(uuid string) error {
		return errors.New("The coin " + uuid + " not in transaction.")
	}
	ERR_OWNER_FROM_COINS_EXISTS_ALREADY = func(pub, coin string) error {
		return errors.New("The owner " + pub + " for coin " + coin + " exists already.")
	}
	ERR_PROOF_VERIFICATION_IS_NOT_CORRECT = errors.New("The proof's verification is not correct.")
	ERR_PROOF_VERIFICATION_IS_NOT_VALID   = errors.New("The proof's verification is not valid.")
	ERR_TRANSACTION_HAS_BEEN_RECEIVED     = errors.New("The transaction has been received.")
	ERR_NEW_OWNERS_NOT_EQUAL_TO_COINS     = errors.New("The number of new owners is not equal to the coins.")
)

func ValidateReceive(state *dbpkg.State, rd models.ReceiveData, sig []byte) (uint32, error) {
	if len(rd.TransactionHash) == 0 {
		return models.CodeTypeUnauthorized, ERR_TRANSACTION_HASH_EMPTY
	}
	tr, err := state.GetTransaction(rd.TransactionHash)
	if err != nil {
		return models.CodeTypeUnauthorized, ERR_TRANSACTION_HASH_DOES_NOT_EXIST
	}

	if len(rd.NewOwners) != len(tr.Coins) {
		return models.CodeTypeUnauthorized, ERR_NEW_OWNERS_NOT_EQUAL_TO_COINS
	}
	for coin, owner := range rd.NewOwners {
		isFoundCoin := false
		for _, trCoin := range tr.Coins {
			if coin == trCoin {
				isFoundCoin = true
				break
			}
		}
		if !isFoundCoin {
			return models.CodeTypeUnauthorized, ERR_COIN_IS_NOT_IN_TRANSACTION(coin)
		}

		_, err := state.GetOwner(owner)
		if err == nil {
			return models.CodeTypeUnauthorized, ERR_OWNER_FROM_COINS_EXISTS_ALREADY(owner, coin)
		}
	}

	proof, err := tr.Proof.GetProof()
	if err != nil {
		return models.CodeTypeServerError, errors.New("A validator accepted an incorrect proof for the transaction " + rd.TransactionHash)
	}

	pvp, err := rd.ProofVerification.GetProof()
	if err != nil {
		return models.CodeTypeUnauthorized, ERR_PROOF_VERIFICATION_IS_NOT_CORRECT
	}

	suite := edwards25519.NewBlakeSHA256Ed25519()
	err = proof.Verify(suite, pvp.G, pvp.H, pvp.XG, pvp.XH)
	if err != nil {
		return models.CodeTypeUnauthorized, ERR_PROOF_VERIFICATION_IS_NOT_VALID
	}
	owners := []string{}
	for _, v := range rd.NewOwners {
		owners = append(owners, v)
	}
	msg, _ := json.Marshal(rd)
	isValid, err := utils.MultiVerify(owners, sig, msg)
	if err != nil {
		return models.CodeTypeEncodingError, err
	}
	if !isValid {
		return models.CodeTypeUnauthorized, ERR_SIGNATURE_NOT_VALID
	}

	if tr.IsCoinsReceived {
		return models.CodeTypeUnauthorized, ERR_TRANSACTION_HAS_BEEN_RECEIVED
	}
	return models.CodeTypeOK, nil
}
