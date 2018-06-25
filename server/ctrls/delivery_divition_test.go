package ctrls

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"

	"github.com/dedis/kyber"
	"github.com/dedis/kyber/group/edwards25519"
	"github.com/dedis/kyber/proof/dleq"
	"github.com/dedis/kyber/util/random"
	"github.com/mragiadakos/tendermoney/server/ctrls/utils"

	"github.com/satori/go.uuid"

	"github.com/mragiadakos/tendermoney/server/ctrls/models"
	"github.com/mragiadakos/tendermoney/server/ctrls/validations"
	"github.com/stretchr/testify/assert"
)

func TestDeliveryDivitionFailCoinEmpty(t *testing.T) {
	app := NewTMApplication()
	d := models.Delivery{}
	d.Type = models.DIVIDE
	data := models.DivitionData{}
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_COIN_EMPTY, errors.New(resp.Log))
}

func TestDeliveryDivitionFailNewCoinsEmpty(t *testing.T) {
	app := NewTMApplication()
	d := models.Delivery{}
	d.Type = models.DIVIDE
	data := models.DivitionData{}
	data.Coin = uuid.NewV4().String()
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_NEW_COINS_EMPTY, errors.New(resp.Log))
}

func TestDeliveryDivitionFailNewCoinsOnEmptyOwner(t *testing.T) {
	app := NewTMApplication()
	d := models.Delivery{}
	d.Type = models.DIVIDE
	data := models.DivitionData{}
	data.Coin = uuid.NewV4().String()
	data.NewCoins = map[string]models.Coin{}
	newCoin := uuid.NewV4().String()
	data.NewCoins[newCoin] = models.Coin{Value: 2}
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_NEW_COIN_OWNER_EMPTY(newCoin), errors.New(resp.Log))
}

func TestDeliveryDivitionFailNewCoinsNonConstantValues(t *testing.T) {
	app := NewTMApplication()
	_, pubHex := utils.CreateKeyPair()
	d := models.Delivery{}
	d.Type = models.DIVIDE
	data := models.DivitionData{}
	data.Coin = uuid.NewV4().String()
	data.NewCoins = map[string]models.Coin{}
	newCoin := uuid.NewV4().String()
	data.NewCoins[newCoin] = models.Coin{
		Value: 2.5,
		Owner: pubHex,
	}
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_NEW_COIN_NON_CONSTANT_VALUE(newCoin), errors.New(resp.Log))
}

func TestDeliveryDivitionFailTheSameOwner(t *testing.T) {
	app := NewTMApplication()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	coin, _ := newCoin(t, app, inflatorKp, inflatorPubHex, 1)

	d := models.Delivery{}
	d.Type = models.DIVIDE
	data := models.DivitionData{}
	data.Coin = coin
	data.NewCoins = map[string]models.Coin{}

	_, newCoin1PubHex := utils.CreateKeyPair()
	data.NewCoins[uuid.NewV4().String()] = models.Coin{
		Value: 0.50,
		Owner: newCoin1PubHex,
	}

	data.NewCoins[uuid.NewV4().String()] = models.Coin{
		Value: 0.50,
		Owner: newCoin1PubHex,
	}

	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_NEW_COINS_EQUAL_OWNER, errors.New(resp.Log))
}

func TestDeliveryDivitionFailCoinDoesNotExist(t *testing.T) {
	app := NewTMApplication()

	d := models.Delivery{}
	d.Type = models.DIVIDE
	data := models.DivitionData{}
	data.Coin = uuid.NewV4().String()
	data.NewCoins = map[string]models.Coin{}

	_, newCoin1PubHex := utils.CreateKeyPair()
	data.NewCoins[uuid.NewV4().String()] = models.Coin{
		Value: 0.50,
		Owner: newCoin1PubHex,
	}

	_, newCoin2PubHex := utils.CreateKeyPair()
	data.NewCoins[uuid.NewV4().String()] = models.Coin{
		Value: 1,
		Owner: newCoin2PubHex,
	}

	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_COIN_DOES_NOT_EXISTS, errors.New(resp.Log))
}

func TestDeliveryDivitionFailTheSumOfNewCoins(t *testing.T) {
	app := NewTMApplication()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	coin, _ := newCoin(t, app, inflatorKp, inflatorPubHex, 1)

	d := models.Delivery{}
	d.Type = models.DIVIDE
	data := models.DivitionData{}
	data.Coin = coin
	data.NewCoins = map[string]models.Coin{}

	_, newCoin1PubHex := utils.CreateKeyPair()
	data.NewCoins[uuid.NewV4().String()] = models.Coin{
		Value: 0.50,
		Owner: newCoin1PubHex,
	}

	_, newCoin2PubHex := utils.CreateKeyPair()
	data.NewCoins[uuid.NewV4().String()] = models.Coin{
		Value: 1,
		Owner: newCoin2PubHex,
	}

	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_NEW_COINS_IS_NOT_EQUAL_TO_THE_COIN, errors.New(resp.Log))
}

func TestDeliveryDivitionFailOnNewCoinExists(t *testing.T) {
	app := NewTMApplication()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	coin, _ := newCoin(t, app, inflatorKp, inflatorPubHex, 1)

	d := models.Delivery{}
	d.Type = models.DIVIDE
	data := models.DivitionData{}
	data.Coin = coin
	data.NewCoins = map[string]models.Coin{}

	_, newCoin1PubHex := utils.CreateKeyPair()
	data.NewCoins[coin] = models.Coin{
		Value: 0.50,
		Owner: newCoin1PubHex,
	}

	_, newCoin2PubHex := utils.CreateKeyPair()
	data.NewCoins[uuid.NewV4().String()] = models.Coin{
		Value: 0.50,
		Owner: newCoin2PubHex,
	}

	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_COIN_FROM_NEW_COINS_EXISTS_ALREADY(coin), errors.New(resp.Log))
}

func TestDeliveryDivitionFailOnNewOwnerExists(t *testing.T) {
	app := NewTMApplication()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	coin, kp := newCoin(t, app, inflatorKp, inflatorPubHex, 1)
	pubB, _ := kp.Public.MarshalBinary()
	oldOwnerPubHex := hex.EncodeToString(pubB)
	d := models.Delivery{}
	d.Type = models.DIVIDE
	data := models.DivitionData{}
	data.Coin = coin
	data.NewCoins = map[string]models.Coin{}

	_, newCoin1PubHex := utils.CreateKeyPair()
	data.NewCoins[uuid.NewV4().String()] = models.Coin{
		Value: 0.50,
		Owner: newCoin1PubHex,
	}

	data.NewCoins[uuid.NewV4().String()] = models.Coin{
		Value: 0.50,
		Owner: oldOwnerPubHex,
	}

	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_OWNER_FROM_NEW_COINS_EXISTS_ALREADY(oldOwnerPubHex), errors.New(resp.Log))
}

func TestDeliveryDivitionFailOnSignature(t *testing.T) {
	app := NewTMApplication()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	coin, oldKp := newCoin(t, app, inflatorKp, inflatorPubHex, 1)

	d := models.Delivery{}
	d.Type = models.DIVIDE
	data := models.DivitionData{}
	data.Coin = coin
	data.NewCoins = map[string]models.Coin{}

	nc1, newCoin1PubHex := utils.CreateKeyPair()
	data.NewCoins[uuid.NewV4().String()] = models.Coin{
		Value: 0.50,
		Owner: newCoin1PubHex,
	}

	nc2, newCoin2PubHex := utils.CreateKeyPair()
	data.NewCoins[uuid.NewV4().String()] = models.Coin{
		Value: 0.50,
		Owner: newCoin2PubHex,
	}

	d.Data = data
	msg, _ := json.Marshal(d.Data)

	// add a letter to the keys so the signature does not validate
	newCoins := map[string]models.Coin{}
	for k, v := range data.NewCoins {
		newCoins[k+"l"] = v
	}
	data.NewCoins = newCoins

	d.Data = data
	d.Signature, _ = utils.MultiSignature(
		[]kyber.Scalar{oldKp.Private, nc1.Private, nc2.Private},
		msg,
	)

	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_SIGNATURE_NOT_VALIDATE, errors.New(resp.Log))
}

func TestDeliveryDivitionSuccess(t *testing.T) {
	app := NewTMApplication()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	coin, oldKp := newCoin(t, app, inflatorKp, inflatorPubHex, 1)

	d := models.Delivery{}
	d.Type = models.DIVIDE
	data := models.DivitionData{}
	data.Coin = coin
	data.NewCoins = map[string]models.Coin{}

	nc1, newCoin1PubHex := utils.CreateKeyPair()
	data.NewCoins[uuid.NewV4().String()] = models.Coin{
		Value: 0.50,
		Owner: newCoin1PubHex,
	}

	nc2, newCoin2PubHex := utils.CreateKeyPair()
	data.NewCoins[uuid.NewV4().String()] = models.Coin{
		Value: 0.50,
		Owner: newCoin2PubHex,
	}

	d.Data = data
	msg, _ := json.Marshal(d.Data)

	d.Data = data
	d.Signature, _ = utils.MultiSignature(
		[]kyber.Scalar{oldKp.Private, nc1.Private, nc2.Private},
		msg,
	)

	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeOK, resp.Code)

	// check that coin and its owner does not exists
	_, err := app.state.GetCoin(coin)
	assert.NotNil(t, err)
	pubB, _ := oldKp.Public.MarshalBinary()
	oldOwnerPubHex := hex.EncodeToString(pubB)
	_, err = app.state.GetOwner(oldOwnerPubHex)
	assert.NotNil(t, err)

	// check that the new coins with the owners exist
	for k, v := range data.NewCoins {
		c, err := app.state.GetCoin(k)
		assert.Nil(t, err)
		if c != nil {
			assert.Equal(t, v.Value, c.Value)
		}
		uuid, err := app.state.GetOwner(v.Owner)
		assert.Nil(t, err)
		assert.Equal(t, k, uuid)
	}
}

func TestDeliveryDivitionFailOnLockCoin(t *testing.T) {
	app := NewTMApplication()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	coin, oldKp := newCoin(t, app, inflatorKp, inflatorPubHex, 1)

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	proof, _, _, _ := dleq.NewDLEQProof(suite, g, h, x)

	// transaction will lock the coins
	transact(t, app, []string{coin}, []string{}, proof, []kyber.Scalar{oldKp.Private})

	d := models.Delivery{}
	d.Type = models.DIVIDE
	data := models.DivitionData{}
	data.Coin = coin
	data.NewCoins = map[string]models.Coin{}

	nc1, newCoin1PubHex := utils.CreateKeyPair()
	data.NewCoins[uuid.NewV4().String()] = models.Coin{
		Value: 0.50,
		Owner: newCoin1PubHex,
	}

	nc2, newCoin2PubHex := utils.CreateKeyPair()
	data.NewCoins[uuid.NewV4().String()] = models.Coin{
		Value: 0.50,
		Owner: newCoin2PubHex,
	}

	d.Data = data
	msg, _ := json.Marshal(d.Data)

	d.Data = data
	d.Signature, _ = utils.MultiSignature(
		[]kyber.Scalar{oldKp.Private, nc1.Private, nc2.Private},
		msg,
	)

	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_COIN_IS_LOCKED(coin), errors.New(resp.Log))
}
