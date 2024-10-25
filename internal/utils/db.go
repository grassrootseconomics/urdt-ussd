package utils

import (
	"encoding/binary"
)

type DataTyp uint16

const (
	DATA_ACCOUNT DataTyp = iota
	DATA_ACCOUNT_CREATED
	DATA_TRACKING_ID
	DATA_PUBLIC_KEY
	DATA_CUSTODIAL_ID
	DATA_ACCOUNT_PIN
	DATA_ACCOUNT_STATUS
	DATA_FIRST_NAME
	DATA_FAMILY_NAME
	DATA_YOB
	DATA_LOCATION
	DATA_GENDER
	DATA_OFFERINGS
	DATA_RECIPIENT
	DATA_AMOUNT
	DATA_TEMPORARY_PIN
	DATA_ADDRESS
)

func typToBytes(typ DataTyp) []byte {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], uint16(typ))
	return b[:]
}

func PackKey(typ DataTyp, data []byte) []byte {
	v := typToBytes(typ)
	return append(v, data...)
}
