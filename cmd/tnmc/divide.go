package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"

	"github.com/mragiadakos/tendermoney/server/ctrls/models"
	client "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"

	"github.com/mragiadakos/tendermoney/server/ctrls/utils"
	"github.com/satori/go.uuid"

	"github.com/dedis/kyber"
	"github.com/dedis/kyber/group/edwards25519"

	"io/ioutil"
)

func divide(vault, coin string, values []float64) ([]string, error) {
	coinFile, err := ioutil.ReadFile(vault + "/" + coin)
	if err != nil {
		return nil, errors.New("Error: The file for the coin " + coin + " is missing.")
	}
	cj := CoinJson{}
	err = json.Unmarshal(coinFile, &cj)
	if err != nil {
		return nil, errors.New("Error: The file for the coin " + coin + " does not unmarshal: " + err.Error())
	}

	suite := edwards25519.NewBlakeSHA256Ed25519()
	oldOwnerPrivB, err := hex.DecodeString(cj.OwnerPrivateKey)
	if err != nil {
		return nil, errors.New("Error: The coin " + coin + " has not a correct private key in hexadecimal format.")
	}
	oldOwnerPriv := suite.Scalar()
	err = oldOwnerPriv.UnmarshalBinary(oldOwnerPrivB)
	if err != nil {
		return nil, errors.New("Error: The coin " + coin + " has not a correct private key.")
	}
	ncjs := []CoinJson{}
	privs := []kyber.Scalar{oldOwnerPriv}
	for _, val := range values {
		newOwnerKp, newOwnerPubHex := utils.CreateKeyPair()
		ncj := CoinJson{}
		ncj.UUID = uuid.NewV4().String()
		ncj.OwnerPublicKey = newOwnerPubHex
		newOwnerPrivB, _ := newOwnerKp.Private.MarshalBinary()
		newOwnerPrivHex := hex.EncodeToString(newOwnerPrivB)
		ncj.OwnerPrivateKey = newOwnerPrivHex
		ncj.Value = val
		ncjs = append(ncjs, ncj)
		privs = append(privs, newOwnerKp.Private)
	}

	data := models.DivitionData{}
	data.Coin = coin
	data.NewCoins = map[string]models.Coin{}
	for _, ncj := range ncjs {
		c := models.Coin{}
		c.Owner = ncj.OwnerPublicKey
		c.Value = ncj.Value
		data.NewCoins[ncj.UUID] = c
	}

	d := models.Delivery{}
	d.Type = models.DIVIDE
	d.Data = data
	msg, _ := json.Marshal(data)
	d.Signature, _ = utils.MultiSignature(privs, msg)

	dB, _ := json.Marshal(d)

	cli := client.NewHTTP(Confs.TendermintNode, "/websocket")
	btc, err := cli.BroadcastTxCommit(types.Tx(dB))
	if err != nil {
		return nil, errors.New("Error: " + err.Error())
	}
	if btc.CheckTx.Code > models.CodeTypeOK {
		return nil, errors.New("Error: " + btc.CheckTx.Log)
	}

	// save the new coins
	filenames := []string{}
	for _, newCoin := range ncjs {
		coinFileB, _ := json.Marshal(newCoin)
		filename := vault + "/" + newCoin.UUID
		err = ioutil.WriteFile(filename, coinFileB, 0644)
		if err != nil {
			return nil, errors.New("Error: " + err.Error())
		}
		filenames = append(filenames, filename)
	}

	// delete the previous coin
	os.Remove(vault + "/" + coin)

	return filenames, nil
}
