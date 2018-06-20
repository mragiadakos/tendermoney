package models

import "encoding/json"

type DeliveryType string

const (
	INFLATE = DeliveryType("inflate")
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

const (
	CodeTypeOK            uint32 = 0
	CodeTypeEncodingError uint32 = 1
	CodeTypeBadNonce      uint32 = 2
	CodeTypeUnauthorized  uint32 = 3
	CodeTypeServerError   uint32 = 4
)
