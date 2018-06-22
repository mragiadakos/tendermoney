package models

import "encoding/json"

type DeliveryType string

var CONSTANT_VALUES = []float64{500, 100, 50, 20, 10, 5, 2, 1, 0.50, 0.20, 0.10, 0.05, 0.02, 0.01}

const (
	INFLATE = DeliveryType("inflate")
	SUM     = DeliveryType("sum")
	TAX     = DeliveryType("tax")
	DIVIDE  = DeliveryType("devide")
	SEND    = DeliveryType("send")
	RECEIVE = DeliveryType("receive")
)

type TypeDeliveryInterface interface {
	GetType() DeliveryType
}

type Delivery struct {
	Type      DeliveryType
	Signature string
	Data      interface{}
}

func (d *Delivery) GetType() DeliveryType {
	return d.Type
}

func (d *Delivery) GetInflationData() InflationData {
	b, _ := json.Marshal(d.Data)
	i := InflationData{}
	json.Unmarshal(b, &i)
	return i
}

func (d *Delivery) GetSumData() SumData {
	b, _ := json.Marshal(d.Data)
	i := SumData{}
	json.Unmarshal(b, &i)
	return i
}

func (d *Delivery) GetDivitionData() DivitionData {
	b, _ := json.Marshal(d.Data)
	i := DivitionData{}
	json.Unmarshal(b, &i)
	return i
}

const (
	CodeTypeOK            uint32 = 0
	CodeTypeEncodingError uint32 = 1
	CodeTypeBadNonce      uint32 = 2
	CodeTypeUnauthorized  uint32 = 3
	CodeTypeServerError   uint32 = 4
)
