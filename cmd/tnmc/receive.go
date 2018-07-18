package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/dedis/kyber"
	"github.com/mragiadakos/tendermoney/app/ctrls/utils"
	client "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"

	"github.com/mragiadakos/tendermoney/app/ctrls/models"
)

func receive(vault, hash, secret string) ([]string, error) {
	qmt, err := getTransaction(hash)
	if err != nil {
		return nil, err
	}
	if qmt.IsCoinsReceived {
		return nil, errors.New("Error: The coins has already been received.")
	}

	b, err := base64.RawStdEncoding.DecodeString(secret)
	if err != nil {
		return nil, errors.New("Error: The secret has problem with base64 encoding, " + err.Error())
	}
	pv := models.ProofVerification{}
	err = json.Unmarshal(b, &pv)
	if err != nil {
		return nil, errors.New("Error: The secret has problem with json encoding, " + err.Error())
	}

	newOwnerPubPerCoin := map[string]string{}
	newOwnersPrivHexPerCoin := map[string]string{}
	newOwnersPrivs := []kyber.Scalar{}
	for _, coin := range qmt.Coins {
		newOwnerKp, newOwnerPubHex := utils.CreateKeyPair()
		newOwnerPubPerCoin[coin] = newOwnerPubHex
		privB, _ := newOwnerKp.Private.MarshalBinary()
		newOwnerPrivHex := hex.EncodeToString(privB)
		newOwnersPrivHexPerCoin[coin] = newOwnerPrivHex
		newOwnersPrivs = append(newOwnersPrivs, newOwnerKp.Private)
	}

	data := models.ReceiveData{}
	data.NewOwners = newOwnerPubPerCoin
	data.TransactionHash = hash
	data.ProofVerification = pv

	d := models.Delivery{}
	d.Type = models.RECEIVE
	d.Data = data
	msg, _ := json.Marshal(data)
	d.Signature, _ = utils.MultiSignature(newOwnersPrivs, msg)

	dB, _ := json.Marshal(d)
	cli := client.NewHTTP(Confs.TendermintNode, "/websocket")
	btc, err := cli.BroadcastTxCommit(types.Tx(dB))
	if err != nil {
		return nil, errors.New("Error: " + err.Error())
	}
	if btc.CheckTx.Code > models.CodeTypeOK {
		return nil, errors.New("Error: " + btc.CheckTx.Log)
	}

	filenames := []string{}
	for coin, pub := range newOwnerPubPerCoin {
		priv := newOwnersPrivHexPerCoin[coin]
		qmc, err := getCoin(coin)
		if err != nil {
			fmt.Println("Error: The coin " + coin + " disappeared from the system: " + err.Error())
			fmt.Println("The public key of the coin: " + pub)
			fmt.Println("The private key of the coin: " + priv)
			continue
		}
		cj := CoinJson{}
		cj.UUID = coin
		cj.Value = qmc.Value
		cj.OwnerPublicKey = pub
		cj.OwnerPrivateKey = priv

		coinFileB, _ := json.Marshal(cj)
		filename := vault + "/" + cj.UUID
		err = ioutil.WriteFile(filename, coinFileB, 0644)
		if err != nil {
			fmt.Println("Error: The coin " + coin + " failed to be written in a file: " + err.Error())
			fmt.Println("The public key of the coin: " + pub)
			fmt.Println("The private key of the coin: " + priv)
			continue
		}
		filenames = append(filenames, filename)
	}

	return filenames, nil
}
