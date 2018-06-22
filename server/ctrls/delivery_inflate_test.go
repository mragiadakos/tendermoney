package ctrls

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/dedis/kyber/group/edwards25519"
	"github.com/satori/go.uuid"

	"github.com/mragiadakos/tendermoney/server/confs"
	"github.com/mragiadakos/tendermoney/server/ctrls/dbpkg"
	"github.com/mragiadakos/tendermoney/server/ctrls/models"
	"github.com/mragiadakos/tendermoney/server/ctrls/utils"
	"github.com/mragiadakos/tendermoney/server/ctrls/validations"

	"github.com/stretchr/testify/assert"
)

func TestDeliveryInflationFailOnCoinEmpty(t *testing.T) {
	app := NewTMApplication()
	di := models.Delivery{}
	di.Type = models.INFLATE
	data := models.InflationData{}
	di.Data = data
	b, _ := json.Marshal(di)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_COIN_EMPTY, errors.New(resp.Log))
}

func TestDeliveryInflationFailOnSignatureEmpty(t *testing.T) {
	app := NewTMApplication()
	di := models.Delivery{}
	di.Type = models.INFLATE
	data := models.InflationData{}
	data.Coin = uuid.NewV4().String()
	di.Data = data
	b, _ := json.Marshal(di)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_SIGNATURE_EMPTY, errors.New(resp.Log))

}

func TestDeliveryInflationFailOnOwnerEmpty(t *testing.T) {
	app := NewTMApplication()
	kp, _ := utils.CreateKeyPair()
	di := models.Delivery{}
	di.Type = models.INFLATE
	di.Signature = "lalla"
	data := models.InflationData{}
	data.Coin = uuid.NewV4().String()
	data.Owner = ""
	di.Data = data
	msg, _ := json.Marshal(di.Data)
	di.Signature, _ = utils.Sign(kp.Private, msg)
	b, _ := json.Marshal(di)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_OWNER_EMPTY, errors.New(resp.Log))

}

func TestDeliveryInflationFailOnInflatorEmpty(t *testing.T) {
	app := NewTMApplication()
	kp, pubHex := utils.CreateKeyPair()
	di := models.Delivery{}
	di.Type = models.INFLATE
	data := models.InflationData{}
	data.Coin = uuid.NewV4().String()
	data.Owner = pubHex
	di.Data = data
	msg, _ := json.Marshal(di.Data)
	di.Signature, _ = utils.Sign(kp.Private, msg)
	b, _ := json.Marshal(di)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_INFLATOR_EMPTY, errors.New(resp.Log))
}

func TestDeliveryInflationFailTheInflatorIsNotInTheList(t *testing.T) {
	app := NewTMApplication()
	kp, pubHex := utils.CreateKeyPair()
	di := models.Delivery{}
	di.Type = models.INFLATE
	data := models.InflationData{}
	data.Coin = uuid.NewV4().String()
	data.Owner = pubHex
	data.Inflator = pubHex
	di.Data = data
	msg, _ := json.Marshal(di.Data)
	di.Signature, _ = utils.Sign(kp.Private, msg)
	b, _ := json.Marshal(di)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_INFLATOR_NOT_IN_LIST, errors.New(resp.Log))
}

func TestDeliveryInflationFailOnNotConstantValue(t *testing.T) {
	app := NewTMApplication()
	kp, pubHex := utils.CreateKeyPair()
	di := models.Delivery{}
	di.Type = models.INFLATE
	data := models.InflationData{}
	data.Coin = uuid.NewV4().String()
	data.Owner = pubHex
	data.Inflator = pubHex
	confs.Conf.Inflators = []string{pubHex}
	data.Value = 2.5
	di.Data = data
	msg, _ := json.Marshal(di.Data)
	di.Signature, _ = utils.Sign(kp.Private, msg)
	b, _ := json.Marshal(di)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_VALUE_NOT_IN_LIST, errors.New(resp.Log))
}

func TestDeliveryInflationFailOnSignature(t *testing.T) {
	app := NewTMApplication()
	kp, pubHex := utils.CreateKeyPair()
	di := models.Delivery{}
	di.Type = models.INFLATE
	data := models.InflationData{}
	data.Coin = uuid.NewV4().String()
	data.Owner = pubHex
	data.Inflator = pubHex
	confs.Conf.Inflators = []string{pubHex}
	data.Value = 2
	di.Data = data
	msg, _ := json.Marshal(di.Data)

	suite := edwards25519.NewBlakeSHA256Ed25519()
	onePrivate := suite.Scalar().Add(kp.Private, kp.Private)
	data.Value = 5
	di.Data = data
	di.Signature, _ = utils.Sign(onePrivate, msg)
	b, _ := json.Marshal(di)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_SIGNATURE_NOT_VALIDATE, errors.New(resp.Log))
}

func TestDeliveryInflationSuccessOnSignature(t *testing.T) {
	app := NewTMApplication()
	ownerKp, ownerPubHex := utils.CreateKeyPair()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	di := models.Delivery{}
	di.Type = models.INFLATE
	data := models.InflationData{}
	data.Coin = uuid.NewV4().String()
	data.Owner = ownerPubHex
	data.Inflator = inflatorPubHex
	confs.Conf.Inflators = []string{inflatorPubHex}
	data.Value = 2
	di.Data = data
	msg, _ := json.Marshal(di.Data)

	suite := edwards25519.NewBlakeSHA256Ed25519()
	onePrivate := suite.Scalar().Add(inflatorKp.Private, ownerKp.Private)
	di.Signature, _ = utils.Sign(onePrivate, msg)
	b, _ := json.Marshal(di)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeOK, resp.Code)
}

func TestDeliveryInflationFailCoinOnAlreadyExists(t *testing.T) {
	app := NewTMApplication()
	ownerKp, ownerPubHex := utils.CreateKeyPair()
	ownerKp2, owner2PubHex := utils.CreateKeyPair()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	di := models.Delivery{}
	di.Type = models.INFLATE
	data := models.InflationData{}
	data.Coin = uuid.NewV4().String()
	data.Owner = ownerPubHex
	data.Inflator = inflatorPubHex
	confs.Conf.Inflators = []string{inflatorPubHex}
	data.Value = 2
	di.Data = data
	msg, _ := json.Marshal(di.Data)

	suite := edwards25519.NewBlakeSHA256Ed25519()
	onePrivate := suite.Scalar().Add(inflatorKp.Private, ownerKp.Private)
	di.Signature, _ = utils.Sign(onePrivate, msg)
	b, _ := json.Marshal(di)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeOK, resp.Code)

	data.Owner = owner2PubHex
	data.Inflator = inflatorPubHex
	confs.Conf.Inflators = []string{inflatorPubHex}
	data.Value = 2
	di.Data = data
	msg, _ = json.Marshal(di.Data)

	onePrivate = suite.Scalar().Add(inflatorKp.Private, ownerKp2.Private)
	di.Signature, _ = utils.Sign(onePrivate, msg)
	b, _ = json.Marshal(di)
	resp = app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, dbpkg.ERR_COIN_EXISTS_ALREADY(data.Coin), errors.New(resp.Log))
}

func TestDeliveryInflationFailOwnerOnAlreadyExists(t *testing.T) {
	app := NewTMApplication()
	ownerKp, ownerPubHex := utils.CreateKeyPair()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	di := models.Delivery{}
	di.Type = models.INFLATE
	data := models.InflationData{}
	data.Coin = uuid.NewV4().String()
	data.Owner = ownerPubHex
	data.Inflator = inflatorPubHex
	confs.Conf.Inflators = []string{inflatorPubHex}
	data.Value = 2
	di.Data = data
	msg, _ := json.Marshal(di.Data)

	suite := edwards25519.NewBlakeSHA256Ed25519()
	onePrivate := suite.Scalar().Add(inflatorKp.Private, ownerKp.Private)
	di.Signature, _ = utils.Sign(onePrivate, msg)
	b, _ := json.Marshal(di)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeOK, resp.Code)

	data.Coin = uuid.NewV4().String()
	data.Inflator = inflatorPubHex
	confs.Conf.Inflators = []string{inflatorPubHex}
	data.Value = 2
	di.Data = data
	msg, _ = json.Marshal(di.Data)

	onePrivate = suite.Scalar().Add(inflatorKp.Private, ownerKp.Private)
	di.Signature, _ = utils.Sign(onePrivate, msg)
	b, _ = json.Marshal(di)
	resp = app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, dbpkg.ERR_OWNER_EXISTS_ALREADY(data.Owner), errors.New(resp.Log))
}
