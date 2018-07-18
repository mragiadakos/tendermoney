package ctrls

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"

	"github.com/dedis/kyber"
	"github.com/dedis/kyber/group/edwards25519"
	"github.com/dedis/kyber/proof/dleq"
	"github.com/dedis/kyber/util/random"
	"github.com/mragiadakos/tendermoney/app/confs"
	"github.com/mragiadakos/tendermoney/app/ctrls/models"
	"github.com/mragiadakos/tendermoney/app/ctrls/query"
	"github.com/mragiadakos/tendermoney/app/ctrls/utils"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	"github.com/tendermint/abci/types"
)

func TestQueryCoinFailOnNotFound(t *testing.T) {
	app := NewTMApplication()

	qreq := types.RequestQuery{}
	uuid := uuid.NewV4().String()
	qreq.Path = "get_coin?coin=" + uuid

	resp := app.Query(qreq)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, query.ERR_COIN_NOT_FOUND(uuid), errors.New(resp.Log))
}

func TestQueryCoinFailOnCoinNotSubmitted(t *testing.T) {
	app := NewTMApplication()

	qreq := types.RequestQuery{}
	qreq.Path = "get_coin"
	resp := app.Query(qreq)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, query.ERR_COIN_HAS_NOT_BEEN_SUBMITTED, errors.New(resp.Log))
}

func TestQueryCoinSuccess(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}

	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.50)

	qreq := types.RequestQuery{}
	qreq.Path = "get_coin?coin=" + coin
	resp := app.Query(qreq)
	assert.Equal(t, models.CodeTypeOK, resp.Code)
	qmc := query.QueryModelCoin{}
	json.Unmarshal(resp.Value, &qmc)
	assert.Equal(t, 0.50, qmc.Value)
	pub, _ := coinKp.Public.MarshalBinary()
	ownerPubHex := hex.EncodeToString(pub)
	assert.Equal(t, ownerPubHex, qmc.Owner)
	assert.False(t, qmc.IsLocked)
	assert.Equal(t, coin, qmc.Coin)

}

func TestQueryCoinByOwnerFailPublicIsEmpty(t *testing.T) {
	app := NewTMApplication()

	qreq := types.RequestQuery{}
	qreq.Path = "get_coin_by_owner"
	resp := app.Query(qreq)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, query.ERR_OWNER_HAS_NOT_BEEN_SUBMITTED, errors.New(resp.Log))
}

func TestQueryCoinByOwnerFailPublicIsNotFound(t *testing.T) {
	app := NewTMApplication()

	qreq := types.RequestQuery{}
	owner := "blalalla"
	qreq.Path = "get_coin_by_owner?owner=" + owner
	resp := app.Query(qreq)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, query.ERR_OWNER_HAS_NOT_BEEN_FOUND(owner), errors.New(resp.Log))
}

func TestQueryCoinByOwnerSuccessful(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}

	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.50)

	pub, _ := coinKp.Public.MarshalBinary()
	ownerPubHex := hex.EncodeToString(pub)

	qreq := types.RequestQuery{}
	qreq.Path = "get_coin_by_owner?owner=" + ownerPubHex
	resp := app.Query(qreq)
	assert.Equal(t, models.CodeTypeOK, resp.Code)

	qmc := query.QueryModelCoin{}
	json.Unmarshal(resp.Value, &qmc)
	assert.Equal(t, 0.50, qmc.Value)
	assert.Equal(t, ownerPubHex, qmc.Owner)
	assert.Equal(t, coin, qmc.Coin)
	assert.False(t, qmc.IsLocked)
}

func TestQueryTaxFailOnNoTaxes(t *testing.T) {
	app := NewTMApplication()
	qreq := types.RequestQuery{}
	qreq.Path = "get_latest_tax"
	resp := app.Query(qreq)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, query.ERR_THERE_NO_TAXES, errors.New(resp.Log))
}

func TestQueryTaxSuccess(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}

	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	createTax(t, app, inflatorKp, inflatorPubHex, 10)
	createTax(t, app, inflatorKp, inflatorPubHex, 2)

	qreq := types.RequestQuery{}
	qreq.Path = "get_latest_tax"
	resp := app.Query(qreq)
	assert.Equal(t, models.CodeTypeOK, resp.Code)

	qmt := query.QueryModelTax{}
	json.Unmarshal(resp.Value, &qmt)
	assert.Equal(t, inflatorPubHex, qmt.Inflator)
	assert.Equal(t, 2, qmt.Percentage)
}

func TestQueryTransactionFailOnHashIsEmpty(t *testing.T) {
	app := NewTMApplication()
	qreq := types.RequestQuery{}
	qreq.Path = "get_transaction"
	resp := app.Query(qreq)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, query.ERR_TRANSACTION_HAS_NOT_BEEN_SUBMITTED, errors.New(resp.Log))
}

func TestQueryTransactionFailOnHashNotFound(t *testing.T) {
	app := NewTMApplication()
	qreq := types.RequestQuery{}
	hash := "blblblblblb"
	qreq.Path = "get_transaction?hash=" + hash
	resp := app.Query(qreq)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, query.ERR_TRANSACTION_HAS_NOT_BEEN_FOUND(hash), errors.New(resp.Log))
}

func TestQueryTransactionSuccess(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()

	confs.Conf.Inflators = []string{inflatorPubHex}

	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 1)

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	coins := []string{coin}
	proof, _, _, _ := dleq.NewDLEQProof(suite, g, h, x)
	transact(t, app, coins, []string{}, proof, []kyber.Scalar{coinKp.Private})

	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	qreq := types.RequestQuery{}
	qreq.Path = "get_transaction?hash=" + hashHex
	resp := app.Query(qreq)
	assert.Equal(t, models.CodeTypeOK, resp.Code)

	qmt := query.QueryModelTransaction{}
	json.Unmarshal(resp.Value, &qmt)

	assert.Equal(t, hashHex, qmt.Hash)
	assert.Equal(t, coins, qmt.Coins)
	assert.Equal(t, []string{}, qmt.Fee)
	assert.False(t, qmt.IsCoinsReceived)
	assert.False(t, qmt.IsFeeReceived)
}

func TestGetUnreceivedFeeTransactionsSuccess(t *testing.T) {
	app := NewTMApplication()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}
	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	suite := edwards25519.NewBlakeSHA256Ed25519()

	for i := 0; i < 3; i++ {
		coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)
		fee, feeKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

		rng := random.New()

		// Create some random secrets and base points
		x := suite.Scalar().Pick(rng)
		g := suite.Point().Pick(rng)
		h := suite.Point().Pick(rng)

		coins := []string{coin}
		fees := []string{fee}
		proof, _, _, _ := dleq.NewDLEQProof(suite, g, h, x)
		transact(t, app, coins, fees, proof, []kyber.Scalar{coinKp.Private, feeKp.Private})
		if i == 0 { //receive only for the first, so to expect two transactions that have not been received
			receivedFee(t, app, inflatorKp, inflatorPubHex, coins, fees)
		}
	}

	qreq := types.RequestQuery{}
	qreq.Path = "get_transactions_with_unreceived_fee"
	resp := app.Query(qreq)
	assert.Equal(t, models.CodeTypeOK, resp.Code)

	qts := []query.QueryModelTransaction{}
	json.Unmarshal(resp.Value, &qts)
	assert.Equal(t, 2, len(qts))
}
