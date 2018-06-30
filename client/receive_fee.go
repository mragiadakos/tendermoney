package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/mragiadakos/tendermoney/server/ctrls/models"
	"github.com/mragiadakos/tendermoney/server/ctrls/utils"
	client "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"

	"github.com/dedis/kyber"
	"github.com/dedis/kyber/group/edwards25519"
)

func receiveFee(inflatorKpj KeyPairJson, vault, hash string) ([]string, error) {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	inflatorPrivateKeyB, _ := hex.DecodeString(inflatorKpj.PrivateKey)
	inflatorPrivateKey := suite.Scalar()
	inflatorPrivateKey.UnmarshalBinary(inflatorPrivateKeyB)

	qmt, err := getTransaction(hash)
	if err != nil {
		return nil, err
	}
	if qmt.IsFeeReceived {
		return nil, errors.New("Error: The fee has already been received.")
	}

	newOwnerPubPerCoin := map[string]string{}
	newOwnersPrivHexPerCoin := map[string]string{}
	privs := []kyber.Scalar{inflatorPrivateKey}
	for _, coin := range qmt.Fee {
		newOwnerKp, newOwnerPubHex := utils.CreateKeyPair()
		newOwnerPubPerCoin[coin] = newOwnerPubHex
		privB, _ := newOwnerKp.Private.MarshalBinary()
		newOwnerPrivHex := hex.EncodeToString(privB)
		newOwnersPrivHexPerCoin[coin] = newOwnerPrivHex
		privs = append(privs, newOwnerKp.Private)
	}

	data := models.RetrieveData{}
	data.Inflator = inflatorKpj.PublicKey
	data.TransactionHash = hash
	data.NewOwners = newOwnerPubPerCoin

	d := models.Delivery{}
	d.Type = models.RETRIEVE_FEE
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
