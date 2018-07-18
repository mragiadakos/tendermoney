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
	"github.com/mragiadakos/tendermoney/app/ctrls/utils"
	"github.com/mragiadakos/tendermoney/app/ctrls/validations"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/abci/types"
)

func TestDeliveryRetrieveFailTransactionDoesNoExist(t *testing.T) {
	app := NewTMApplication()
	d := models.Delivery{}
	d.Type = models.RETRIEVE_FEE
	data := models.RetrieveData{}
	data.TransactionHash = "lalal"
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_TRANSACTION_HASH_DOES_NOT_EXIST, errors.New(resp.Log))
}

func TestDeliveryRetrieveFailTransactionDoesNotHaveFee(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}

	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	coins := []string{coin}
	fees := []string{}
	proof, _, _, _ := dleq.NewDLEQProof(suite, g, h, x)
	transact(t, app, coins, fees, proof, []kyber.Scalar{coinKp.Private})

	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	d := models.Delivery{}
	d.Type = models.RETRIEVE_FEE
	data := models.RetrieveData{}
	data.TransactionHash = hashHex
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_TRANSACTION_DOES_NOT_HAVE_FEE, errors.New(resp.Log))
}

func TestDeliveryRetrieveFailNewOwnersEmpty(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}

	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)
	fee, feeKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	coins := []string{coin}
	fees := []string{fee}
	proof, _, _, _ := dleq.NewDLEQProof(suite, g, h, x)
	transact(t, app, coins, fees, proof, []kyber.Scalar{coinKp.Private, feeKp.Private})

	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	d := models.Delivery{}
	d.Type = models.RETRIEVE_FEE
	data := models.RetrieveData{}
	data.TransactionHash = hashHex
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_NEW_OWNERS_NOT_EQUAL_TO_FEES, errors.New(resp.Log))
}

func TestDeliveryRetrieveFailOnInflatorIsEmpty(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	_, newOwnerPubHex := utils.CreateKeyPair()

	confs.Conf.Inflators = []string{inflatorPubHex}

	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)
	fee, feeKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	coins := []string{coin}
	fees := []string{fee}
	proof, _, _, _ := dleq.NewDLEQProof(suite, g, h, x)
	transact(t, app, coins, fees, proof, []kyber.Scalar{coinKp.Private, feeKp.Private})

	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	d := models.Delivery{}
	d.Type = models.RETRIEVE_FEE
	data := models.RetrieveData{}
	data.TransactionHash = hashHex
	data.NewOwners = map[string]string{}
	data.NewOwners[fee] = newOwnerPubHex
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_INFLATOR_EMPTY, errors.New(resp.Log))
}

func TestDeliveryRetrieveFailOnInflatorNotInList(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	_, newOwnerPubHex := utils.CreateKeyPair()

	confs.Conf.Inflators = []string{inflatorPubHex}

	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)
	fee, feeKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	coins := []string{coin}
	fees := []string{fee}
	proof, _, _, _ := dleq.NewDLEQProof(suite, g, h, x)
	transact(t, app, coins, fees, proof, []kyber.Scalar{coinKp.Private, feeKp.Private})

	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	d := models.Delivery{}
	d.Type = models.RETRIEVE_FEE
	data := models.RetrieveData{}
	data.TransactionHash = hashHex
	data.NewOwners = map[string]string{}
	data.NewOwners[fee] = newOwnerPubHex
	data.Inflator = newOwnerPubHex //using new owner as inflator to fail
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_INFLATOR_NOT_IN_LIST, errors.New(resp.Log))
}

func TestDeliveryRetrieveFailOnNewOwnerExists(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}

	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)
	fee, feeKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	coins := []string{coin}
	fees := []string{fee}
	proof, _, _, _ := dleq.NewDLEQProof(suite, g, h, x)
	transact(t, app, coins, fees, proof, []kyber.Scalar{coinKp.Private, feeKp.Private})

	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	d := models.Delivery{}
	d.Type = models.RETRIEVE_FEE
	data := models.RetrieveData{}
	data.TransactionHash = hashHex
	data.NewOwners = map[string]string{}
	oldOwnerPub, _ := coinKp.Public.MarshalBinary()
	oldOwnerPubHex := hex.EncodeToString(oldOwnerPub)
	data.NewOwners[fee] = oldOwnerPubHex
	data.Inflator = inflatorPubHex
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_OWNER_FROM_COINS_EXISTS_ALREADY(oldOwnerPubHex, fee), errors.New(resp.Log))
}

func TestDeliveryRetrieveFailOnCoinIsNotInFee(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	_, newOwnerPubHex := utils.CreateKeyPair()

	confs.Conf.Inflators = []string{inflatorPubHex}

	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)
	fee, feeKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	coins := []string{coin}
	fees := []string{fee}
	proof, _, _, _ := dleq.NewDLEQProof(suite, g, h, x)
	transact(t, app, coins, fees, proof, []kyber.Scalar{coinKp.Private, feeKp.Private})

	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	d := models.Delivery{}
	d.Type = models.RETRIEVE_FEE
	data := models.RetrieveData{}
	data.TransactionHash = hashHex
	data.NewOwners = map[string]string{}
	data.NewOwners[coin] = newOwnerPubHex // using coin, rather than the fee, to fail
	data.Inflator = inflatorPubHex
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_COIN_IS_NOT_IN_FEE_TRANSACTION(coin), errors.New(resp.Log))
}

func TestDeliveryRetrieveFailOnSignature(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	_, newOwnerPubHex := utils.CreateKeyPair()

	confs.Conf.Inflators = []string{inflatorPubHex}

	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)
	fee, feeKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	coins := []string{coin}
	fees := []string{fee}
	proof, _, _, _ := dleq.NewDLEQProof(suite, g, h, x)
	transact(t, app, coins, fees, proof, []kyber.Scalar{coinKp.Private, feeKp.Private})

	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	d := models.Delivery{}
	d.Type = models.RETRIEVE_FEE
	data := models.RetrieveData{}
	data.TransactionHash = hashHex
	data.NewOwners = map[string]string{}
	data.NewOwners[fee] = newOwnerPubHex
	data.Inflator = inflatorPubHex
	dataB, _ := json.Marshal(data)
	d.Signature, _ = utils.MultiSignature([]kyber.Scalar{inflatorKp.Private}, dataB) // adding only the inflator in the list
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_SIGNATURE_NOT_VALID, errors.New(resp.Log))
}

func TestDeliveryRetrieveFailOnTransactionHasAlreadyBeenRetrieved(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()

	confs.Conf.Inflators = []string{inflatorPubHex}

	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)
	fee, feeKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	coins := []string{coin}
	fees := []string{fee}
	proof, _, _, _ := dleq.NewDLEQProof(suite, g, h, x)
	transact(t, app, coins, fees, proof, []kyber.Scalar{coinKp.Private, feeKp.Private})

	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	resps := []types.ResponseDeliverTx{}
	for i := 0; i < 2; i++ {
		newOwnerPk, newOwnerPubHex := utils.CreateKeyPair()
		d := models.Delivery{}
		d.Type = models.RETRIEVE_FEE
		data := models.RetrieveData{}
		data.TransactionHash = hashHex
		data.NewOwners = map[string]string{}
		data.NewOwners[fee] = newOwnerPubHex
		data.Inflator = inflatorPubHex
		dataB, _ := json.Marshal(data)
		d.Signature, _ = utils.MultiSignature([]kyber.Scalar{inflatorKp.Private, newOwnerPk.Private}, dataB)
		d.Data = data
		b, _ := json.Marshal(d)
		resp := app.DeliverTx(b)
		resps = append(resps, resp)
	}

	assert.Equal(t, models.CodeTypeOK, resps[0].Code)
	assert.Equal(t, models.CodeTypeUnauthorized, resps[1].Code)
	assert.Equal(t, validations.ERR_TRANSACTION_HAS_BEEN_RETRIEVED, errors.New(resps[1].Log))
}

func TestDeliveryRetrieveSuccess(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()

	confs.Conf.Inflators = []string{inflatorPubHex}

	createTax(t, app, inflatorKp, inflatorPubHex, 23)
	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)
	fee, feeKp := newCoin(t, app, inflatorKp, inflatorPubHex, 0.01)

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	coins := []string{coin}
	fees := []string{fee}
	proof, _, _, _ := dleq.NewDLEQProof(suite, g, h, x)
	transact(t, app, coins, fees, proof, []kyber.Scalar{coinKp.Private, feeKp.Private})

	sc, err := app.state.GetCoin(fee)
	assert.Nil(t, err)
	assert.True(t, sc.IsLocked)

	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	newOwnerPk, newOwnerPubHex := utils.CreateKeyPair()
	d := models.Delivery{}
	d.Type = models.RETRIEVE_FEE
	data := models.RetrieveData{}
	data.TransactionHash = hashHex
	data.NewOwners = map[string]string{}
	data.NewOwners[fee] = newOwnerPubHex
	data.Inflator = inflatorPubHex
	dataB, _ := json.Marshal(data)
	d.Signature, _ = utils.MultiSignature([]kyber.Scalar{inflatorKp.Private, newOwnerPk.Private}, dataB)
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeOK, resp.Code)

	// the coins have the new owners
	sc, err = app.state.GetCoin(fee)
	assert.Nil(t, err)
	assert.Equal(t, newOwnerPubHex, sc.Owner)

	// the coin is not locked
	assert.False(t, sc.IsLocked)

	// older owner has been deleted
	oldOwnerPub, _ := feeKp.Public.MarshalBinary()
	oldOwnerPubHex := hex.EncodeToString(oldOwnerPub)
	_, err = app.state.GetOwner(oldOwnerPubHex)
	assert.NotNil(t, err)

	// the transaction has been retrieved
	st, err := app.state.GetTransaction(hashHex)
	assert.Nil(t, err)
	assert.True(t, st.IsFeeReceived)
}
