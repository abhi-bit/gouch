package gouch

import (
	"bytes"
	"encoding/binary"
)

const TOP_BIT_MASK = 0x80

func encode_raw08(val interface{}) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, val)
	bufBytes := buf.Bytes()
	return bufBytes[len(bufBytes)-1:]
}

func decode_raw08(raw []byte) uint8 {
	var rv uint8
	buf := bytes.NewBuffer(raw)
	binary.Read(buf, binary.BigEndian, &rv)
	return rv
}

func encode_raw16(val interface{}) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, val)
	bufBytes := buf.Bytes()
	return bufBytes[len(bufBytes)-2:]
}

func decode_raw16(raw []byte) uint16 {
	var rv uint16
	buf := bytes.NewBuffer(raw)
	binary.Read(buf, binary.BigEndian, &rv)
	return rv
}

func encode_raw24(val interface{}) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, val)
	bufBytes := buf.Bytes()
	return bufBytes[len(bufBytes)-3:]
}

func decode_raw24(raw []byte) uint32 {
	var rv uint32
	buf := bytes.NewBuffer([]byte{0})
	buf.Write(raw)
	binary.Read(buf, binary.BigEndian, &rv)
	return rv
}

func maskOutTopBit(in byte) byte {
	return in &^ TOP_BIT_MASK
}

func decode_raw31(raw []byte) uint32 {
	var rv uint32
	topByte := maskOutTopBit(raw[0])
	buf := bytes.NewBuffer([]byte{topByte})
	buf.Write(raw[1:4])
	binary.Read(buf, binary.BigEndian, &rv)
	return rv
}

func encode_raw32(val interface{}) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, val)
	bufBytes := buf.Bytes()
	return bufBytes[len(bufBytes)-4:]
}

func decode_raw32(raw []byte) uint32 {
	var rv uint32
	buf := bytes.NewBuffer(raw)
	binary.Read(buf, binary.BigEndian, &rv)
	return rv
}

func encode_raw40(val interface{}) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, val)
	bufBytes := buf.Bytes()
	return bufBytes[len(bufBytes)-5:]
}

func decode_raw40(raw []byte) uint64 {
	var rv uint64
	buf := bytes.NewBuffer([]byte{0, 0, 0})
	buf.Write(raw)
	binary.Read(buf, binary.BigEndian, &rv)
	return rv
}

func encode_raw48(val interface{}) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, val)
	bufBytes := buf.Bytes()
	return bufBytes[len(bufBytes)-6:]
}
func decode_raw48(raw []byte) uint64 {
	var rv uint64
	buf := bytes.NewBuffer([]byte{0, 0})
	buf.Write(raw)
	binary.Read(buf, binary.BigEndian, &rv)
	return rv
}

func encode_raw64(val interface{}) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, val)
	bufBytes := buf.Bytes()
	return bufBytes[len(bufBytes)-8:]
}

func decode_raw64(raw []byte) uint64 {
	var rv uint64
	buf := bytes.NewBuffer(raw)
	binary.Read(buf, binary.BigEndian, &rv)
	return rv
}
