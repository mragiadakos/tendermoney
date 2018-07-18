package ctrls

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/dedis/kyber"
	"github.com/dedis/kyber/group/edwards25519"
	"github.com/dedis/kyber/proof/dleq"
	"github.com/dedis/kyber/util/key"
	"github.com/mragiadakos/tendermoney/app/confs"
	"github.com/mragiadakos/tendermoney/app/ctrls/models"
	"github.com/mragiadakos/tendermoney/app/ctrls/utils"
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

func receivedFee(t *testing.T, app *TMApplication, inflatorKp *key.Pair, inflatorPubHex string, coins, fee []string) {
	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	d := models.Delivery{}
	d.Type = models.RETRIEVE_FEE
	data := models.RetrieveData{}
	data.TransactionHash = hashHex
	data.NewOwners = map[string]string{}
	privs := []kyber.Scalar{inflatorKp.Private}
	for _, v := range fee {
		newOwnerPk, newOwnerPubHex := utils.CreateKeyPair()
		data.NewOwners[v] = newOwnerPubHex
		privs = append(privs, newOwnerPk.Private)
	}
	data.Inflator = inflatorPubHex
	dataB, _ := json.Marshal(data)
	d.Signature, _ = utils.MultiSignature(privs, dataB)
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeOK, resp.Code)
}
