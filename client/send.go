package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/dedis/kyber"
	client "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"

	"github.com/mragiadakos/tendermoney/server/ctrls/query"
	"github.com/mragiadakos/tendermoney/server/ctrls/utils"

	"github.com/dedis/kyber/group/edwards25519"
	"github.com/dedis/kyber/proof/dleq"
	"github.com/dedis/kyber/util/random"

	"github.com/mragiadakos/tendermoney/server/ctrls/models"
)

func send(coins, fee []string, vault string) (string, string, error) {
	cjs := []CoinJson{} // the json of the coins
	for _, coin := range coins {
		coinFile, err := ioutil.ReadFile(vault + "/" + coin)
		if err != nil {
			return "", "", errors.New("Error: The file for the coin " + coin + " is missing.")
		}
		cj := CoinJson{}
		err = json.Unmarshal(coinFile, &cj)
		if err != nil {
			return "", "", errors.New("Error: The file for the coin " + coin + " does not unmarshal: " + err.Error())
		}
		cjs = append(cjs, cj)
	}

	fjs := []CoinJson{} // the the json of the coins from the fee
	for _, coin := range fee {
		coinFile, err := ioutil.ReadFile(vault + "/" + coin)
		if err != nil {
			return "", "", errors.New("Error: The file for the coin " + coin + " is missing.")
		}
		cj := CoinJson{}
		err = json.Unmarshal(coinFile, &cj)
		if err != nil {
			return "", "", errors.New("Error: The file for the coin " + coin + " does not unmarshal: " + err.Error())
		}
		fjs = append(fjs, cj)
	}

	data := models.SendData{}
	data.Coins = coins
	data.Fee = fee
	rng := random.New()
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Create some random secrets and base points
	x := suite.Scalar().Pick(rng)
	g := suite.Point().Pick(rng)
	h := suite.Point().Pick(rng)

	proof, xG, xH, err := dleq.NewDLEQProof(suite, g, h, x)
	if err != nil {
		return "", "", errors.New("Error: failed to create a proof:" + err.Error())
	}
	data.Proof = models.NewProof(proof)

	msg, _ := json.Marshal(coins)
	hash := sha256.Sum256(msg)
	hashHex := hex.EncodeToString(hash[:])

	proofVerification := models.NewProofVerification(g, h, xG, xH)
	proofVerificationB, _ := json.Marshal(proofVerification)
	secret := base64.StdEncoding.EncodeToString(proofVerificationB)

	allCjs := append(cjs, fjs...)
	privks := []kyber.Scalar{}
	for _, cj := range allCjs {
		privB, err := hex.DecodeString(cj.OwnerPrivateKey)
		if err != nil {
			return "", "", errors.New("Error: The private key from the owner of the coin " + cj.UUID + ", is not a correct hexadecimal format.")
		}
		priv := suite.Scalar()
		err = priv.UnmarshalBinary(privB)
		if err != nil {
			return "", "", errors.New("Error: The private key from the owner of the coin " + cj.UUID + ", is not correct.")
		}
		privks = append(privks, priv)
	}

	d := models.Delivery{}
	d.Type = models.SEND
	d.Data = data
	dataB, _ := json.Marshal(data)
	d.Signature, _ = utils.MultiSignature(privks, dataB)

	dB, _ := json.Marshal(d)
	cli := client.NewHTTP(Confs.TendermintNode, "/websocket")
	btc, err := cli.BroadcastTxCommit(types.Tx(dB))
	if err != nil {
		return "", "", errors.New("Error: " + err.Error())
	}
	if btc.CheckTx.Code > models.CodeTypeOK {
		return "", "", errors.New("Error: " + btc.CheckTx.Log)
	}

	// remove all the coins
	for _, cj := range allCjs {
		os.Remove(vault + "/" + cj.UUID)
	}

	return hashHex, secret, nil
}

func getTransaction(hash string) (*query.QueryModelTransaction, error) {
	cli := client.NewHTTP(Confs.TendermintNode, "/websocket")
	q, err := cli.ABCIQuery("get_transaction?hash="+hash, nil)
	if err != nil {
		return nil, errors.New("Error:" + err.Error())
	}
	if q.Response.Code > models.CodeTypeOK {
		return nil, errors.New("Error: " + q.Response.Log)
	}

	qmt := query.QueryModelTransaction{}
	json.Unmarshal(q.Response.Value, &qmt)
	return &qmt, nil
}

func getCoin(uuid string) (*query.QueryModelCoin, error) {
	cli := client.NewHTTP(Confs.TendermintNode, "/websocket")
	q, err := cli.ABCIQuery("get_coin?coin="+uuid, nil)
	if err != nil {
		return nil, errors.New("Error:" + err.Error())
	}
	if q.Response.Code > models.CodeTypeOK {
		return nil, errors.New("Error: " + q.Response.Log)
	}

	qmc := query.QueryModelCoin{}
	json.Unmarshal(q.Response.Value, &qmc)
	return &qmc, nil
}

func getTransactionsWithUnreceivedFee(hash string) ([]query.QueryModelTransaction, error) {
	cli := client.NewHTTP(Confs.TendermintNode, "/websocket")
	q, err := cli.ABCIQuery("get_transactions_with_unreceived_fee", nil)
	if err != nil {
		return nil, errors.New("Error:" + err.Error())
	}
	if q.Response.Code > models.CodeTypeOK {
		return nil, errors.New("Error: " + q.Response.Log)
	}

	qmts := []query.QueryModelTransaction{}
	json.Unmarshal(q.Response.Value, &qmts)
	return qmts, nil
}
