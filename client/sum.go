package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/dedis/kyber"
	"github.com/dedis/kyber/group/edwards25519"
	"github.com/mragiadakos/tendermoney/server/ctrls/models"
	"github.com/mragiadakos/tendermoney/server/ctrls/utils"
	"github.com/satori/go.uuid"
	client "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"
)

func sum(coins []string, vault string) (string, error) {
	cjs := []CoinJson{}
	for _, coin := range coins {
		coinFile, err := ioutil.ReadFile(vault + "/" + coin)
		if err != nil {
			return "", errors.New("Error: The file for the coin " + coin + " is missing.")
		}
		cj := CoinJson{}
		err = json.Unmarshal(coinFile, &cj)
		if err != nil {
			return "", errors.New("Error: The file for the coin " + coin + " does not unmarshal: " + err.Error())
		}
		cjs = append(cjs, cj)
	}

	newOwnerKp, newOwnerPubHex := utils.CreateKeyPair()

	data := models.SumData{}
	data.Coins = coins
	data.NewCoin = uuid.NewV4().String()
	data.NewOwner = newOwnerPubHex

	suite := edwards25519.NewBlakeSHA256Ed25519()
	privks := []kyber.Scalar{newOwnerKp.Private}
	sumNumber := 0.0
	for _, cj := range cjs {
		privB, err := hex.DecodeString(cj.OwnerPrivateKey)
		if err != nil {
			return "", errors.New("Error: The private key from the owner of the coin " + cj.UUID + ", is not a correct hexadecimal format.")
		}
		priv := suite.Scalar()
		err = priv.UnmarshalBinary(privB)
		if err != nil {
			return "", errors.New("Error: The private key from the owner of the coin " + cj.UUID + ", is not correct.")
		}
		privks = append(privks, priv)
		sumNumber += cj.Value
	}

	d := models.Delivery{}
	d.Type = models.SUM
	d.Data = data
	dataB, _ := json.Marshal(data)
	d.Signature, _ = utils.MultiSignature(privks, dataB)
	dB, _ := json.Marshal(d)
	cli := client.NewHTTP(Confs.TendermintNode, "/websocket")
	btc, err := cli.BroadcastTxCommit(types.Tx(dB))
	if err != nil {
		return "", errors.New("Error: " + err.Error())
	}
	if btc.CheckTx.Code > models.CodeTypeOK {
		return "", errors.New("Error: " + btc.CheckTx.Log)
	}

	// after the success, we save the coin in the vault
	newCoin := CoinJson{}
	newCoin.UUID = data.NewCoin
	newCoin.Value = sumNumber
	newOwnerPrivB, _ := newOwnerKp.Private.MarshalBinary()
	newOwnerPrivHex := hex.EncodeToString(newOwnerPrivB)
	newCoin.OwnerPrivateKey = newOwnerPrivHex
	newCoin.OwnerPublicKey = newOwnerPubHex

	coinFileB, _ := json.Marshal(newCoin)
	filename := vault + "/" + newCoin.UUID
	err = ioutil.WriteFile(filename, coinFileB, 0644)
	if err != nil {
		return "", errors.New("Error: " + err.Error())
	}

	// remove the old ones
	for _, coin := range coins {
		os.Remove(vault + "/" + coin)
	}

	return filename, nil
}
