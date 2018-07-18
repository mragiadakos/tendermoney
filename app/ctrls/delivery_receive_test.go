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
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/abci/types"
)

func TestDeliveryReceiveFailOnEmptyHash(t *testing.T) {
	app := NewTMApplication()

	d := models.Delivery{}
	d.Type = models.RECEIVE
	data := models.ReceiveData{}
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_TRANSACTION_HASH_EMPTY, errors.New(resp.Log))
}

func TestDeliveryReceiveFailOnHashDoesNotExists(t *testing.T) {
	app := NewTMApplication()

	d := models.Delivery{}
	d.Type = models.RECEIVE
	data := models.ReceiveData{}
	data.TransactionHash = "lallalala"
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_TRANSACTION_HASH_DOES_NOT_EXIST, errors.New(resp.Log))
}

func TestDeliveryReceiveFailOnNumberOfOwnersNotEqualToTheNumberOfCoins(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	_, newOwnerPubHex := utils.CreateKeyPair()

	confs.Conf.Inflators = []string{inflatorPubHex}

	coin1, coin1Kp := newCoin(t, app, inflatorKp, inflatorPubHex, 1)
	coin2, coin2Kp := newCoin(t, app, inflatorKp, inflatorPubHex, 1)

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	coins := []string{coin1, coin2}
	proof, _, _, _ := dleq.NewDLEQProof(suite, g, h, x)
	transact(t, app, coins, []string{}, proof, []kyber.Scalar{coin1Kp.Private, coin2Kp.Private})

	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	d := models.Delivery{}
	d.Type = models.RECEIVE
	data := models.ReceiveData{}
	data.TransactionHash = hashHex
	data.NewOwners = map[string]string{}
	data.NewOwners[coin1] = newOwnerPubHex
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_NEW_OWNERS_NOT_EQUAL_TO_COINS, errors.New(resp.Log))
}

func TestDeliveryReceiveFailOnCoinDoesNotExistsInTransaction(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	_, newOwnerPubHex := utils.CreateKeyPair()

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

	d := models.Delivery{}
	d.Type = models.RECEIVE
	data := models.ReceiveData{}
	data.TransactionHash = hashHex
	fakeCoin := uuid.NewV4().String()
	data.NewOwners = map[string]string{}
	data.NewOwners[fakeCoin] = newOwnerPubHex
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_COIN_IS_NOT_IN_TRANSACTION(fakeCoin), errors.New(resp.Log))
}

func TestDeliveryReceiveFailOnOwnerExistsAlready(t *testing.T) {
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

	d := models.Delivery{}
	d.Type = models.RECEIVE
	data := models.ReceiveData{}
	data.TransactionHash = hashHex
	data.NewOwners = map[string]string{}
	// use the older owner to fail
	pub, _ := coinKp.Public.MarshalBinary()
	coinPubHex := hex.EncodeToString(pub)
	data.NewOwners[coin] = coinPubHex
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_OWNER_FROM_COINS_EXISTS_ALREADY(coinPubHex, coin), errors.New(resp.Log))
}

func TestDeliveryReceiveFailOnProofIsNotCorrect(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	_, newOwnerPubHex := utils.CreateKeyPair()

	confs.Conf.Inflators = []string{inflatorPubHex}

	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 1)

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	coins := []string{coin}
	proof, xG, xH, _ := dleq.NewDLEQProof(suite, g, h, x)
	transact(t, app, coins, []string{}, proof, []kyber.Scalar{coinKp.Private})

	// test that the serialization worked without problems
	pv := models.NewProofVerification(g, h, xG, xH)
	pvp, err := pv.GetProof()
	assert.Nil(t, err)
	err = proof.Verify(suite, pvp.G, pvp.H, pvp.XG, pvp.XH)
	assert.Nil(t, err)

	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	d := models.Delivery{}
	d.Type = models.RECEIVE
	data := models.ReceiveData{}
	data.TransactionHash = hashHex
	pv.GHex += "l" // adding a character to be an incorrect proof
	data.ProofVerification = pv
	data.NewOwners = map[string]string{}
	data.NewOwners[coin] = newOwnerPubHex
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_PROOF_VERIFICATION_IS_NOT_CORRECT, errors.New(resp.Log))
}

func TestDeliveryReceiveFailOnProofDoesNotVerify(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	_, newOwnerPubHex := utils.CreateKeyPair()

	confs.Conf.Inflators = []string{inflatorPubHex}

	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 1)

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	coins := []string{coin}
	proof, xG, xH, _ := dleq.NewDLEQProof(suite, g, h, x)
	transact(t, app, coins, []string{}, proof, []kyber.Scalar{coinKp.Private})

	fakeXG := suite.Point().Add(xG, xH)
	pv := models.NewProofVerification(g, h, fakeXG, xH)

	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	d := models.Delivery{}
	d.Type = models.RECEIVE
	data := models.ReceiveData{}
	data.TransactionHash = hashHex
	data.ProofVerification = pv
	data.NewOwners = map[string]string{}
	data.NewOwners[coin] = newOwnerPubHex
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_PROOF_VERIFICATION_IS_NOT_VALID, errors.New(resp.Log))
}

func TestDeliveryReceiveFailOnFailOnSignatureDoesNotValidate(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	_, newOwnerPubHex := utils.CreateKeyPair()

	confs.Conf.Inflators = []string{inflatorPubHex}

	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 1)

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	coins := []string{coin}
	proof, xG, xH, _ := dleq.NewDLEQProof(suite, g, h, x)
	transact(t, app, coins, []string{}, proof, []kyber.Scalar{coinKp.Private})

	pv := models.NewProofVerification(g, h, xG, xH)

	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	d := models.Delivery{}
	d.Type = models.RECEIVE
	data := models.ReceiveData{}
	data.TransactionHash = hashHex
	data.ProofVerification = pv
	data.NewOwners = map[string]string{}
	data.NewOwners[coin] = newOwnerPubHex
	d.Data = data
	dataB, _ := json.Marshal(data)
	// we will use inflators key to invalidate the signature
	d.Signature, _ = utils.MultiSignature([]kyber.Scalar{inflatorKp.Private}, dataB)
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_SIGNATURE_NOT_VALID, errors.New(resp.Log))
}

func TestDeliveryReceiveFailOnReceivingTwice(t *testing.T) {
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
	proof, xG, xH, _ := dleq.NewDLEQProof(suite, g, h, x)
	transact(t, app, coins, []string{}, proof, []kyber.Scalar{coinKp.Private})

	pv := models.NewProofVerification(g, h, xG, xH)

	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	resps := []types.ResponseDeliverTx{}
	for i := 0; i < 2; i++ {
		newOwnerKp, newOwnerPubHex := utils.CreateKeyPair()
		d := models.Delivery{}
		d.Type = models.RECEIVE
		data := models.ReceiveData{}
		data.TransactionHash = hashHex
		data.ProofVerification = pv
		data.NewOwners = map[string]string{}
		data.NewOwners[coin] = newOwnerPubHex
		d.Data = data
		dataB, _ := json.Marshal(data)
		d.Signature, _ = utils.MultiSignature([]kyber.Scalar{newOwnerKp.Private}, dataB)
		b, _ := json.Marshal(d)
		resp := app.DeliverTx(b)
		resps = append(resps, resp)
	}
	assert.Equal(t, models.CodeTypeOK, resps[0].Code)

	// we expect the second time will fail
	assert.Equal(t, models.CodeTypeUnauthorized, resps[1].Code)
	assert.Equal(t, validations.ERR_TRANSACTION_HAS_BEEN_RECEIVED, errors.New(resps[1].Log))
}

func TestDeliveryReceiveSuccess(t *testing.T) {
	app := NewTMApplication()

	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	newOwnerKp, newOwnerPubHex := utils.CreateKeyPair()

	confs.Conf.Inflators = []string{inflatorPubHex}

	coin, coinKp := newCoin(t, app, inflatorKp, inflatorPubHex, 1)

	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	coins := []string{coin}
	proof, xG, xH, _ := dleq.NewDLEQProof(suite, g, h, x)
	transact(t, app, coins, []string{}, proof, []kyber.Scalar{coinKp.Private})

	pv := models.NewProofVerification(g, h, xG, xH)

	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	d := models.Delivery{}
	d.Type = models.RECEIVE
	data := models.ReceiveData{}
	data.TransactionHash = hashHex
	data.ProofVerification = pv
	data.NewOwners = map[string]string{}
	data.NewOwners[coin] = newOwnerPubHex
	d.Data = data
	dataB, _ := json.Marshal(data)
	d.Signature, _ = utils.MultiSignature([]kyber.Scalar{newOwnerKp.Private}, dataB)
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeOK, resp.Code)

	// check that the coin is unlocked
	isLocked, err := app.state.IsCoinLocked(coin)
	assert.Nil(t, err)
	assert.False(t, isLocked)

	// old owner is deleted
	pub, _ := coinKp.Public.MarshalBinary()
	oldOwnerPubHex := hex.EncodeToString(pub)
	_, err = app.state.GetOwner(oldOwnerPubHex)
	assert.NotNil(t, err)

	// the coin has the new owner and the new owner has the coin
	sc, err := app.state.GetCoin(coin)
	assert.Nil(t, err)
	assert.Equal(t, newOwnerPubHex, sc.Owner)
	coinUuid, err := app.state.GetOwner(newOwnerPubHex)
	assert.Nil(t, err)
	assert.Equal(t, coin, coinUuid)

	// the transaction hash does not exists in the db
	st, err := app.state.GetTransaction(data.TransactionHash)
	assert.Nil(t, err)
	assert.True(t, st.IsCoinsReceived)
}
