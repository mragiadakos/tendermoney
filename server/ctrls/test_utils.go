package ctrls

import (
	"encoding/json"
	"testing"

	"github.com/dedis/kyber"
	"github.com/dedis/kyber/group/edwards25519"
	"github.com/dedis/kyber/proof/dleq"
	"github.com/dedis/kyber/util/key"
	"github.com/mragiadakos/tendermoney/server/confs"
	"github.com/mragiadakos/tendermoney/server/ctrls/models"
	"github.com/mragiadakos/tendermoney/server/ctrls/utils"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func createTax(t *testing.T, app *TMApplication, inflatorKp *key.Pair, inflatorPubHex string, percentage int) {
	d := models.Delivery{}
	d.Type = models.TAX
	data := models.TaxData{}
	data.Percentage = percentage
	data.Inflator = inflatorPubHex
	msg, _ := json.Marshal(data)
	d.Signature, _ = utils.Sign(inflatorKp.Private, msg)
	d.Data = data

	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeOK, resp.Code)
}

func newCoin(t *testing.T, app *TMApplication, inflatorKp *key.Pair, inflatorPubHex string, value float64) (string, *key.Pair) {
	ownerKp, ownerPubHex := utils.CreateKeyPair()
	d := models.Delivery{}
	d.Type = models.INFLATE
	data := models.InflationData{}
	data.Coin = uuid.NewV4().String()
	data.Owner = ownerPubHex
	data.Inflator = inflatorPubHex
	confs.Conf.Inflators = []string{inflatorPubHex}
	data.Value = value
	d.Data = data
	msg, _ := json.Marshal(d.Data)

	suite := edwards25519.NewBlakeSHA256Ed25519()
	onePrivate := suite.Scalar().Add(inflatorKp.Private, ownerKp.Private)
	d.Signature, _ = utils.Sign(onePrivate, msg)
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeOK, resp.Code)
	return data.Coin, ownerKp
}

func transact(t *testing.T, app *TMApplication, coins []string, fee []string, proof *dleq.Proof, privs []kyber.Scalar) {
	d := models.Delivery{}
	d.Type = models.SEND
	data := models.SendData{}
	data.Coins = coins
	data.Fee = fee
	data.Proof = models.NewProof(proof)
	d.Data = data
	msg, _ := json.Marshal(data)
	d.Signature, _ = utils.MultiSignature(privs, msg)
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeOK, resp.Code)
}
