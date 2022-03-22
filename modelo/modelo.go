package modelo

import "encoding/json"

// Structs
// Binding from JSON
type EntelRequest struct {
	Msisdn string `form:"msisdn" json:"msisdn" binding:"required"`
	Uli    string `form:"uli" json:"uli" binding:"required"`
	Apn    string `form:"apn" json:"apn" binding:"required"`
}

type EntelNumberRequest struct {
	Msisdn string `form:"msisdn" json:"msisdn" binding:"required"`
}

type EntelResponseUlify struct {
	Uli    string `form:"uli" json:"uli" binding:"required"`
	Celda  int64  `form:"celda" json:"celda" binding:"required"`
	Sector int64  `form:"sector" json:"sector" binding:"required"`
}

func DecodeData(raw []byte) (data EntelResponseUlify, err error) {
	err = json.Unmarshal(raw, &data)
	return data, err
}
