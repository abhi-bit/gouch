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

func encode_raw_12_28_split(top uint32, bottom uint32) []byte {
	topbuf := new(bytes.Buffer)
	binary.Write(topbuf, binary.BigEndian, top)
	topbytes := topbuf.Bytes()

	newtoptop := topbytes[len(topbytes)-2] & 0x0f << 4
	newtopbottom := topbytes[len(topbytes)-1] & 0xf0 >> 4
	newtop := newtoptop | newtopbottom

	newbottomtop := topbytes[len(topbytes)-1] & 0x0f << 4

	bottombuf := new(bytes.Buffer)
	binary.Write(bottombuf, binary.BigEndian, bottom)
	bottombytes := bottombuf.Bytes()

	newbottombottom := bottombytes[len(bottombytes)-4] & 0x0f

	newbottom := newbottomtop | newbottombottom

	resultbuf := bytes.NewBuffer([]byte{newtop, newbottom})
	resultbuf.Write(bottombytes[len(bottombytes)-3:])
	return resultbuf.Bytes()
}

func decode_raw_12_28_split(data []byte) (top uint32, bottom uint32) {
	kFirstByte := (data[0] & 0xf0) >> 4
	kSecondByteTop := (data[0] & 0xf0) << 4
	kSecondByteBottom := (data[1] & 0xf0) >> 4
	kSecondByte := kSecondByteTop | kSecondByteBottom

	buf := bytes.NewBuffer([]byte{0x00, 0x00, kFirstByte, kSecondByte})
	binary.Read(buf, binary.BigEndian, &top)

	buf = bytes.NewBuffer([]byte{data[1] & 0x0f})
	buf.Write(data[2:])
	binary.Read(buf, binary.BigEndian, &bottom)
	return
}

func valueTopBit(in byte) bool {
	if in&TOP_BIT_MASK != 0 {
		return true
	}
	return false
}

func decode_raw_1_47_split(raw []byte) (bool, uint64) {
	var rint uint64
	rbool := valueTopBit(raw[0])
	topByte := maskOutTopBit(raw[0])
	buf := bytes.NewBuffer([]byte{0, 0, topByte})
	buf.Write(raw[1:6])
	binary.Read(buf, binary.BigEndian, &rint)
	return rbool, rint
}

func encode_raw_1_47_split(topBit bool, rest uint64) []byte {
	// encode the rest portion first
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, rest)
	bufbytes := buf.Bytes()
	// then overwrite the top bit
	if topBit {
		bufbytes[len(bufbytes)-6] = bufbytes[len(bufbytes)-6] | TOP_BIT_MASK
	} else {
		bufbytes[len(bufbytes)-6] = bufbytes[len(bufbytes)-6] &^ TOP_BIT_MASK
	}
	return bufbytes[len(bufbytes)-6:]
}
