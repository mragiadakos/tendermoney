package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/dedis/kyber"
	"github.com/dedis/kyber/group/edwards25519"
	"github.com/mragiadakos/tendermoney/app/ctrls/models"
	"github.com/mragiadakos/tendermoney/app/ctrls/query"
	"github.com/mragiadakos/tendermoney/app/ctrls/utils"
	client "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"
)

func tax(inflatorKpj KeyPairJson, percent int) error {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	inflatorPrivateKeyB, _ := hex.DecodeString(inflatorKpj.PrivateKey)
	inflatorPrivateKey := suite.Scalar()
	inflatorPrivateKey.UnmarshalBinary(inflatorPrivateKeyB)

	data := models.TaxData{}
	data.Inflator = inflatorKpj.PublicKey
	data.Percentage = percent

	d := models.Delivery{}
	d.Data = data
	d.Type = models.TAX
	msg, _ := json.Marshal(data)
	d.Signature, _ = utils.MultiSignature([]kyber.Scalar{inflatorPrivateKey}, msg)

	dB, _ := json.Marshal(d)

	cli := client.NewHTTP(Confs.TendermintNode, "/websocket")

	btc, err := cli.BroadcastTxCommit(types.Tx(dB))
	if err != nil {
		return errors.New("Error: " + err.Error())
	}
	if btc.CheckTx.Code > models.CodeTypeOK {
		return errors.New("Error: " + btc.CheckTx.Log)
	}

	return nil
}

func getLatestTax() (*query.QueryModelTax, error) {
	cli := client.NewHTTP(Confs.TendermintNode, "/websocket")
	q, err := cli.ABCIQuery("get_latest_tax", nil)
	if err != nil {
		return nil, errors.New("Error:" + err.Error())
	}
	if q.Response.Code > models.CodeTypeOK {
		return nil, errors.New("Error: " + q.Response.Log)
	}
	qmt := query.QueryModelTax{}
	json.Unmarshal(q.Response.Value, &qmt)
	return &qmt, nil
}
