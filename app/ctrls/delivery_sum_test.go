package ctrls

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"

	"github.com/mragiadakos/tendermoney/app/ctrls/dbpkg"

	"github.com/dedis/kyber"
	"github.com/dedis/kyber/group/edwards25519"
	"github.com/dedis/kyber/proof/dleq"
	"github.com/dedis/kyber/util/random"

	"github.com/mragiadakos/tendermoney/app/ctrls/models"
	"github.com/mragiadakos/tendermoney/app/ctrls/utils"
	"github.com/mragiadakos/tendermoney/app/ctrls/validations"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestDeliverySumFailOnEmptyCoins(t *testing.T) {
	app := NewTMApplication()
	d := models.Delivery{}
	d.Type = models.SUM
	data := models.SumData{}
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_COINS_EMPTY, errors.New(resp.Log))
}

func TestDeliverySumFailOnCoinAddedTwice(t *testing.T) {
	app := NewTMApplication()
	d := models.Delivery{}
	d.Type = models.SUM
	data := models.SumData{}
	coin := uuid.NewV4().String()
	data.Coins = []string{coin, coin}
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_COIN_FROM_COINS_ADDED_TWICE(coin), errors.New(resp.Log))
}

func TestDeliverySumFailOnSignatureEmpty(t *testing.T) {
	app := NewTMApplication()
	d := models.Delivery{}
	d.Type = models.SUM
	data := models.SumData{}
	data.Coins = []string{uuid.NewV4().String()}
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_SIGNATURE_EMPTY, errors.New(resp.Log))
}

func TestDeliverySumFailOnEmptyNewCoin(t *testing.T) {
	app := NewTMApplication()
	kp, _ := utils.CreateKeyPair()
	d := models.Delivery{}
	d.Type = models.SUM
	data := models.SumData{}
	data.Coins = []string{uuid.NewV4().String()}
	d.Data = data
	msg, _ := json.Marshal(d.Data)
	d.Signature, _ = utils.Sign(kp.Private, msg)
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_NEW_COIN_EMPTY, errors.New(resp.Log))
}

func TestDeliverySumFailOnEmptyNewOwner(t *testing.T) {
	app := NewTMApplication()
	kp, _ := utils.CreateKeyPair()
	d := models.Delivery{}
	d.Type = models.SUM
	data := models.SumData{}
	data.Coins = []string{uuid.NewV4().String()}
	data.NewCoin = uuid.NewV4().String()
	d.Data = data
	msg, _ := json.Marshal(d.Data)
	d.Signature, _ = utils.Sign(kp.Private, msg)
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_NEW_OWNER_EMPTY, errors.New(resp.Log))
}

func TestDeliveryOnCoinDoesNotExists(t *testing.T) {
	app := NewTMApplication()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	newOwnerKp, newOwnerPubHex := utils.CreateKeyPair()

	coin1, owner1 := newCoin(t, app, inflatorKp, inflatorPubHex, 0.50)

	d := models.Delivery{}
	d.Type = models.SUM
	data := models.SumData{}
	data.Coins = []string{coin1 + "k"}
	data.NewCoin = uuid.NewV4().String()
	data.NewOwner = newOwnerPubHex
	d.Data = data
	msg, _ := json.Marshal(d.Data)
	d.Signature, _ = utils.MultiSignature(
		[]kyber.Scalar{newOwnerKp.Private, owner1.Private},
		msg,
	)
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, dbpkg.ERR_COIN_DOES_NOT_EXISTS(data.Coins[0]), errors.New(resp.Log))
}

func TestDeliverySumFailSumOfCoinsNotConstantValue(t *testing.T) {
	app := NewTMApplication()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	newOwnerKp, newOwnerPubHex := utils.CreateKeyPair()

	coin1, owner1 := newCoin(t, app, inflatorKp, inflatorPubHex, 0.50)
	coin2, owner2 := newCoin(t, app, inflatorKp, inflatorPubHex, 2)

	d := models.Delivery{}
	d.Type = models.SUM
	data := models.SumData{}
	data.Coins = []string{coin1, coin2}
	data.NewCoin = uuid.NewV4().String()
	data.NewOwner = newOwnerPubHex
	d.Data = data
	msg, _ := json.Marshal(d.Data)
	d.Signature, _ = utils.MultiSignature(
		[]kyber.Scalar{newOwnerKp.Private, owner1.Private, owner2.Private},
		msg,
	)
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_SUM_OF_COINS_NON_CONSTANT, errors.New(resp.Log))
}

func TestDeliverySumFailOnNewCoinExistsAlready(t *testing.T) {
	app := NewTMApplication()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	newOwnerKp, newOwnerPubHex := utils.CreateKeyPair()

	coin1, owner1 := newCoin(t, app, inflatorKp, inflatorPubHex, 0.50)

	d := models.Delivery{}
	d.Type = models.SUM
	data := models.SumData{}
	data.Coins = []string{coin1}
	data.NewCoin = coin1
	data.NewOwner = newOwnerPubHex
	d.Data = data
	msg, _ := json.Marshal(d.Data)
	d.Signature, _ = utils.MultiSignature(
		[]kyber.Scalar{newOwnerKp.Private, owner1.Private},
		msg,
	)
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_NEW_COIN_EXISTS_ALREADY, errors.New(resp.Log))
}

func TestDeliverySumFailOnNewOwnerExistsAlready(t *testing.T) {
	app := NewTMApplication()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()

	coin1, owner1 := newCoin(t, app, inflatorKp, inflatorPubHex, 0.50)
	pubB, _ := owner1.Public.MarshalBinary()
	oldOwnerHex := hex.EncodeToString(pubB)
	d := models.Delivery{}
	d.Type = models.SUM
	data := models.SumData{}
	data.Coins = []string{coin1}
	data.NewCoin = uuid.NewV4().String()
	data.NewOwner = oldOwnerHex
	d.Data = data
	msg, _ := json.Marshal(d.Data)
	d.Signature, _ = utils.MultiSignature(
		[]kyber.Scalar{owner1.Private},
		msg,
	)
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_NEW_OWNER_EXISTS_ALREADY, errors.New(resp.Log))
}

func TestDeliverySumFailOnSignature(t *testing.T) {
	app := NewTMApplication()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	newOwnerKp, newOwnerPubHex := utils.CreateKeyPair()

	coin1, owner1 := newCoin(t, app, inflatorKp, inflatorPubHex, 0.50)
	coin2, owner2 := newCoin(t, app, inflatorKp, inflatorPubHex, 0.50)

	d := models.Delivery{}
	d.Type = models.SUM
	data := models.SumData{}
	data.Coins = []string{coin1, coin2}
	data.NewCoin = uuid.NewV4().String()
	data.NewOwner = newOwnerPubHex
	d.Data = data
	msg, _ := json.Marshal(d.Data)
	d.Signature, _ = utils.MultiSignature(
		[]kyber.Scalar{newOwnerKp.Private, owner1.Private, owner2.Private},
		msg,
	)
	data.NewCoin += "k"
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_SIGNATURE_NOT_VALID, errors.New(resp.Log))
}

func TestDeliverySumSuccess(t *testing.T) {
	app := NewTMApplication()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	newOwnerKp, newOwnerPubHex := utils.CreateKeyPair()

	coin1, owner1 := newCoin(t, app, inflatorKp, inflatorPubHex, 0.50)
	coin2, owner2 := newCoin(t, app, inflatorKp, inflatorPubHex, 0.50)

	d := models.Delivery{}
	d.Type = models.SUM
	data := models.SumData{}
	data.Coins = []string{coin1, coin2}
	data.NewCoin = uuid.NewV4().String()
	data.NewOwner = newOwnerPubHex
	d.Data = data
	msg, _ := json.Marshal(d.Data)
	d.Signature, _ = utils.MultiSignature(
		[]kyber.Scalar{newOwnerKp.Private, owner1.Private, owner2.Private},
		msg,
	)
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeOK, resp.Code)

	// check that coins do not exist
	_, err := app.state.GetCoin(coin1)
	assert.NotNil(t, err)
	_, err = app.state.GetCoin(coin2)
	assert.NotNil(t, err)

	// check that owners do not exist
	ownerPub1, _ := owner1.Public.MarshalBinary()
	owner1PubHex := hex.EncodeToString(ownerPub1)
	_, err = app.state.GetOwner(owner1PubHex)
	assert.NotNil(t, err)

	ownerPub2, _ := owner2.Public.MarshalBinary()
	owner2PubHex := hex.EncodeToString(ownerPub2)
	_, err = app.state.GetOwner(owner2PubHex)
	assert.NotNil(t, err)

	// check that new coin exists
	newCoin, err := app.state.GetCoin(data.NewCoin)
	assert.Nil(t, err)
	if newCoin != nil {
		assert.Equal(t, 1.0, newCoin.Value)
	}
	// check that new owner exists
	newCoinUuid, err := app.state.GetOwner(data.NewOwner)
	assert.Nil(t, err)
	assert.Equal(t, data.NewCoin, newCoinUuid)
}

func TestDeliverySumFailOnCoinLocked(t *testing.T) {
	app := NewTMApplication()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	newOwnerKp, newOwnerPubHex := utils.CreateKeyPair()

	coin1, owner1 := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)
	coin2, owner2 := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	proof, _, _, _ := dleq.NewDLEQProof(suite, g, h, x)

	// transaction will lock the coins
	transact(t, app, []string{coin1}, []string{}, proof, []kyber.Scalar{owner1.Private})

	d := models.Delivery{}
	d.Type = models.SUM
	data := models.SumData{}
	data.Coins = []string{coin1, coin2}
	data.NewCoin = uuid.NewV4().String()
	data.NewOwner = newOwnerPubHex
	d.Data = data
	msg, _ := json.Marshal(d.Data)
	d.Signature, _ = utils.MultiSignature(
		[]kyber.Scalar{newOwnerKp.Private, owner1.Private, owner2.Private},
		msg,
	)
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_COIN_IS_LOCKED(coin1), errors.New(resp.Log))

}
