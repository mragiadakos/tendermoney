package ctrls

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/mragiadakos/tendermoney/server/confs"
	"github.com/mragiadakos/tendermoney/server/ctrls/models"
	"github.com/mragiadakos/tendermoney/server/ctrls/utils"
	"github.com/mragiadakos/tendermoney/server/ctrls/validations"

	"github.com/stretchr/testify/assert"
)

func TestDeliveryTaxFailOnNegativePercentage(t *testing.T) {
	app := NewTMApplication()
	d := models.Delivery{}
	d.Type = models.TAX
	data := models.TaxData{}
	data.Percentage = -1
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_TAX_NEGATIVE, errors.New(resp.Log))
}

func TestDeliveryTaxFailOnOverPercentage(t *testing.T) {
	app := NewTMApplication()
	d := models.Delivery{}
	d.Type = models.TAX
	data := models.TaxData{}
	data.Percentage = 101
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_TAX_OVER_ONE_PERCENT, errors.New(resp.Log))
}

func TestDeliveryTaxFailOnNotInflator(t *testing.T) {
	app := NewTMApplication()
	d := models.Delivery{}
	d.Type = models.TAX
	data := models.TaxData{}
	data.Percentage = 23
	d.Data = data
	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)
	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_INFLATOR_NOT_IN_LIST, errors.New(resp.Log))
}

func TestDeliveryTaxFailOnSignatureValidate(t *testing.T) {
	app := NewTMApplication()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}

	d := models.Delivery{}
	d.Type = models.TAX
	data := models.TaxData{}
	data.Percentage = 23
	data.Inflator = inflatorPubHex
	msg, _ := json.Marshal(data)
	d.Signature, _ = utils.Sign(inflatorKp.Private, msg)

	data.Percentage += 1
	d.Data = data

	b, _ := json.Marshal(d)
	resp := app.DeliverTx(b)

	assert.Equal(t, models.CodeTypeUnauthorized, resp.Code)
	assert.Equal(t, validations.ERR_SIGNATURE_NOT_VALID, errors.New(resp.Log))
}

func TestDeliveryTaxSuccess(t *testing.T) {
	app := NewTMApplication()
	inflatorKp, inflatorPubHex := utils.CreateKeyPair()
	confs.Conf.Inflators = []string{inflatorPubHex}
	datas := []models.TaxData{}
	for i := 0; i < 3; i++ {
		d := models.Delivery{}
		d.Type = models.TAX
		data := models.TaxData{}
		data.Percentage = 23
		data.Inflator = inflatorPubHex
		msg, _ := json.Marshal(data)
		d.Signature, _ = utils.Sign(inflatorKp.Private, msg)
		d.Data = data

		b, _ := json.Marshal(d)
		resp := app.DeliverTx(b)
		assert.Equal(t, models.CodeTypeOK, resp.Code)
		datas = append(datas, data)
	}
	tax := app.state.GetTax()
	assert.Equal(t, tax, datas[2])
}

func TestDeliveryTaxFeeBasedOnConstantValue(t *testing.T) {
	data := models.TaxData{}
	data.Percentage = 23
	fee := data.GetFeeFromTransaction(5)
	assert.Equal(t, 1.15, fee)

	fee = data.GetFeeFromTransaction(1)
	assert.Equal(t, 0.23, fee)

	fee = data.GetFeeFromTransaction(0.50)
	assert.Equal(t, 0.12, fee)

	fee = data.GetFeeFromTransaction(0.10)
	assert.Equal(t, 0.02, fee)

	fee = data.GetFeeFromTransaction(0.02)
	assert.Equal(t, 0.01, fee)

	fee = data.GetFeeFromTransaction(0.01)
	assert.Equal(t, 0.01, fee)

	data = models.TaxData{}
	data.Percentage = 0
	fee = data.GetFeeFromTransaction(0.01)
	assert.Equal(t, 0.0, fee)
}
