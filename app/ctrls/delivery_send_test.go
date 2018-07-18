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

	"github.com/satori/go.uuid"

	"github.com/mragiadakos/tendermoney/server/confs"
	"github.com/mragiadakos/tendermoney/server/ctrls/dbpkg"
	"github.com/mragiadakos/tendermoney/server/ctrls/models"
	"github.com/mragiadakos/tendermoney/server/ctrls/utils"
	"github.com/mragiadakos/tendermoney/server/ctrls/validations"
	"github.com/stretchr/testify/assert"
)

func TestDeliverySendFailOnEmptyCoins(t *testing.T) {
	app := NewTMApplication()

	d := models.Delivery{}
	d.Type = models.SEND
	data := models.SendData{}
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_COINS_EMPTY, errors.New(resp.Log))
}

func TestDeliverySendFailOnEmptyFees(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}

	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	coin1, _ := newCoin(t, app, inflatorKp, inflatorPubHex, 0.50)

	d := models.Delivery{}
	d.Type = models.SEND
	data := models.SendData{}
	data.Coins = []string{coin1}
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_FEES_EMPTY, errors.New(resp.Log))
}

func TestDeliverySendFailOnCoinAddedTwice(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}

	coin1, _ := newCoin(t, app, inflatorKp, inflatorPubHex, 0.50)

	d := models.Delivery{}
	d.Type = models.SEND
	data := models.SendData{}
	data.Coins = []string{coin1, coin1}
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_COIN_FROM_COINS_ADDED_TWICE(coin1), errors.New(resp.Log))
}

func TestDeliverySendFailOnCoinFromFeeAddedTwice(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}
	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	coin1, _ := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)
	fee1, _ := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

	d := models.Delivery{}
	d.Type = models.SEND
	data := models.SendData{}
	data.Coins = []string{coin1}
	data.Fee = []string{fee1, fee1}
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_COIN_FROM_FEE_ADDED_TWICE(fee1), errors.New(resp.Log))
}

func TestDeliverySendFailOnCoinAddedOnBoth(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}
	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	coin1, _ := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)
	fee1, _ := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

	d := models.Delivery{}
	d.Type = models.SEND
	data := models.SendData{}
	data.Coins = []string{coin1}
	data.Fee = []string{fee1, coin1}
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_COIN_ADDED_ON_BOTH_COINS_AND_FEE(coin1), errors.New(resp.Log))
}

func TestDeliverySendFailOnCoinDoesNotExist(t *testing.T) {
	app := NewTMApplication()

	d := models.Delivery{}
	d.Type = models.SEND
	data := models.SendData{}
	coin := uuid.NewV4().String()
	data.Coins = []string{coin}
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_COIN_FROM_COINS_DOES_NOT_EXISTS(coin), errors.New(resp.Log))
}

func TestDeliverySendFailOnFeeCoinDoesNotExist(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}

	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	coin1, _ := newCoin(t, app, inflatorKp, inflatorPubHex, 0.50)

	d := models.Delivery{}
	d.Type = models.SEND
	data := models.SendData{}
	data.Coins = []string{coin1}
	fee := uuid.NewV4().String()
	data.Fee = []string{fee}
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_COIN_FROM_FEE_DOES_NOT_EXISTS(fee), errors.New(resp.Log))
}

func TestDeliverySendFailOnFeeNotBasedOnTax(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}

	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	coin, _ := newCoin(t, app, inflatorKp, inflatorPubHex, 1)
	fee1, _ := newCoin(t, app, inflatorKp, inflatorPubHex, 0.20)
	fee2, _ := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

	d := models.Delivery{}
	d.Type = models.SEND
	data := models.SendData{}
	data.Coins = []string{coin}
	data.Fee = []string{fee1, fee2}
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_FEE_NOT_BASED_ON_TAX(0.02), errors.New(resp.Log))
}

func TestDeliverySendFailOnSignature(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}

	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 1)
	fee1, fee1Kp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.20)
	fee2, fee2Kp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.02)
	fee3, fee3Kp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)
	fee4, _ := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

	d := models.Delivery{}
	d.Type = models.SEND
	data := models.SendData{}
	data.Coins = []string{coin}
	data.Fee = []string{fee1, fee2, fee3}
	d.Data = data
	msg, _ := json.Marshal(data)
	d.Signature, _ = utils.MultiSignature([]kyber.Scalar{coinKp.Private, fee1Kp.Private, fee2Kp.Private, fee3Kp.Private}, msg)
	data.Fee = append(data.Fee, fee4) // we add an extra fee without signing it
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_SIGNATURE_NOT_VALID, errors.New(resp.Log))
}

func TestDeliverySendFailOnProofNotEncodedProperly(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}

	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 1)
	fee1, fee1Kp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.20)
	fee2, fee2Kp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.02)
	fee3, fee3Kp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

	d := models.Delivery{}
	d.Type = models.SEND
	data := models.SendData{}
	data.Coins = []string{coin}
	data.Fee = []string{fee1, fee2, fee3}

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	proof, xG, xH, err := dleq.NewDLEQProof(suite, g, h, x)
	data.Proof = models.NewProof(proof)

	// making sure that the proof validates after de-serialization
	checkProof, err := data.Proof.GetProof()
	assert.Nil(t, err)
	err = checkProof.Verify(suite, g, h, xG, xH)
	assert.Nil(t, err)

	// destroying the hex on one so to throw the error
	data.Proof.CHex += "l"

	d.Data = data
	msg, _ := json.Marshal(data)
	d.Signature, _ = utils.MultiSignature([]kyber.Scalar{coinKp.Private, fee1Kp.Private, fee2Kp.Private, fee3Kp.Private}, msg)
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_PROOF_NOT_CORRECT, errors.New(resp.Log))
}

func TestDeliverySendSuccessful(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}

	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 1)
	fee1, fee1Kp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.20)
	fee2, fee2Kp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.02)
	fee3, fee3Kp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

	d := models.Delivery{}
	d.Type = models.SEND
	data := models.SendData{}
	data.Coins = []string{coin}
	data.Fee = []string{fee1, fee2, fee3}

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	proof, xG, xH, err := dleq.NewDLEQProof(suite, g, h, x)
	data.Proof = models.NewProof(proof)

	// making sure that the proof validates after de-serialization
	checkProof, err := data.Proof.GetProof()
	assert.Nil(t, err)
	err = checkProof.Verify(suite, g, h, xG, xH)
	assert.Nil(t, err)

	d.Data = data
	msg, _ := json.Marshal(data)
	d.Signature, _ = utils.MultiSignature([]kyber.Scalar{coinKp.Private, fee1Kp.Private, fee2Kp.Private, fee3Kp.Private}, msg)
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeOK, resp.Code)

	coinb, _ := json.Marshal(data.Coins)
	hash := sha256.Sum256(coinb)
	hashHex := hex.EncodeToString(hash[:])
	st, err := app.state.GetTransaction(hashHex)
	assert.Nil(t, err)
	assert.Equal(t, &dbpkg.StateTransaction{SendData: data}, st)
}

func TestDeliverySendFailOnUsingLockedCoin(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}

	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 1)
	fee1, fee1Kp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.20)
	fee2, fee2Kp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.02)
	fee3, fee3Kp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

	d := models.Delivery{}
	d.Type = models.SEND
	data := models.SendData{}
	data.Coins = []string{coin}
	data.Fee = []string{fee1, fee2, fee3}

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	proof, xG, xH, err := dleq.NewDLEQProof(suite, g, h, x)
	data.Proof = models.NewProof(proof)

	// making sure that the proof validates after de-serialization
	checkProof, err := data.Proof.GetProof()
	assert.Nil(t, err)
	err = checkProof.Verify(suite, g, h, xG, xH)
	assert.Nil(t, err)

	d.Data = data
	msg, _ := json.Marshal(data)
	d.Signature, _ = utils.MultiSignature([]kyber.Scalar{coinKp.Private, fee1Kp.Private, fee2Kp.Private, fee3Kp.Private}, msg)

	// we will transact the same coins successfully before transacting them again
	transact(t, app, data.Coins, data.Fee, proof, []kyber.Scalar{coinKp.Private, fee1Kp.Private, fee2Kp.Private, fee3Kp.Private})

	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_COIN_IS_LOCKED(coin), errors.New(resp.Log))
}
