package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/dedis/kyber"
	"github.com/dedis/kyber/group/edwards25519"
	"github.com/mragiadakos/tendermoney/server/ctrls/models"
	"github.com/mragiadakos/tendermoney/server/ctrls/utils"
	uuid "github.com/satori/go.uuid"
	client "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"
)

func inflate(inflatorKpj KeyPairJson, vault string, value float64) (string, error) {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	inflatorPrivateKeyB, _ := hex.DecodeString(inflatorKpj.PrivateKey)
	inflatorPrivateKey := suite.Scalar()
	inflatorPrivateKey.UnmarshalBinary(inflatorPrivateKeyB)

	ownerKp, ownerPubHex := utils.CreateKeyPair()
	data := models.InflationData{}
	data.Inflator = inflatorKpj.PublicKey
	data.Coin = uuid.NewV4().String()
	data.Value = value
	data.Owner = ownerPubHex
	d := models.Delivery{}
	d.Type = models.INFLATE
	d.Data = data
	msg, _ := json.Marshal(data)
	d.Signature, _ = utils.MultiSignature([]kyber.Scalar{inflatorPrivateKey, ownerKp.Private}, msg)

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
	newCoin.UUID = data.Coin
	newCoin.Value = data.Value
	ownerPrivB, _ := ownerKp.Private.MarshalBinary()
	ownerPrivHex := hex.EncodeToString(ownerPrivB)
	newCoin.OwnerPrivateKey = ownerPrivHex
	newCoin.OwnerPublicKey = ownerPubHex

	coinFileB, _ := json.Marshal(newCoin)
	filename := vault + "/" + newCoin.UUID
	err = ioutil.WriteFile(filename, coinFileB, 0644)
	if err != nil {
		return "", errors.New("Error: " + err.Error())
	}
	return filename, nil
}
