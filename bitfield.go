package gouch

import (
	"bytes"
	"encoding/binary"
)

// TopBitMask used for masking the first bit
const TopBitMask byte = 0x80

func encodeRaw08(val interface{}) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, val)
	bufbytes := buf.Bytes()
	return bufbytes[len(bufbytes)-1:]
}

func encodeRaw16(val interface{}) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, val)
	bufbytes := buf.Bytes()
	return bufbytes[len(bufbytes)-2:]
}

func encodeRaw31Highestbiton(val interface{}) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, val)
	bufbytes := buf.Bytes()
	bufbytes[len(bufbytes)-4] = bufbytes[len(bufbytes)-4] | TopBitMask
	return bufbytes[len(bufbytes)-4:]
}

func encodeRaw32(val interface{}) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, val)
	bufbytes := buf.Bytes()
	return bufbytes[len(bufbytes)-4:]
}

func encodeRaw40(val interface{}) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, val)
	bufbytes := buf.Bytes()
	return bufbytes[len(bufbytes)-5:]
}

func encodeRaw48(val interface{}) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, val)
	bufbytes := buf.Bytes()
	return bufbytes[len(bufbytes)-6:]
}

func decodeRaw08(raw []byte) uint8 {
	var rv uint8
	buf := bytes.NewBuffer(raw)
	binary.Read(buf, binary.BigEndian, &rv)
	return rv
}

func decodeRaw16(raw []byte) uint16 {
	var rv uint16

	//Using sync.Pool
	if b, ok := twoByte.Get().([]byte); ok {
		copy(b, raw)
		rv = binary.BigEndian.Uint16(b)
		twoByte.Put(b)
	}
	return rv
}

// just like raw32 but mask out the top bit
func decodeRaw31(raw []byte) uint32 {
	var rv uint32
	if b, ok := fourByte.Get().([]byte); ok {
		topByte := maskOutTopBit(raw[0])
		b[0] = topByte
		k := 1
		for _, v := range raw[1:4] {
			b[k] = v
			k++
		}
		rv = binary.BigEndian.Uint32(b)
		fourByte.Put(b)
	}
	return rv
}

func decodeRaw32(raw []byte) uint32 {
	var rv uint32
	if b, ok := fourByte.Get().([]byte); ok {
		copy(b, raw)
		rv = binary.BigEndian.Uint32(b)
		fourByte.Put(b)
	}
	return rv
}

func decodeRaw40(raw []byte) uint64 {
	var rv uint64
	if b, ok := eightByte.Get().([]byte); ok {
		b[0] = 0x00
		b[1] = 0x00
		b[2] = 0x00
		k := 3
		for _, v := range raw[0:5] {
			b[k] = v
			k++
		}
		rv = binary.BigEndian.Uint64(b)
		eightByte.Put(b)
	}
	return rv
}

func decodeRaw48(raw []byte) uint64 {
	var rv uint64
	if b, ok := eightByte.Get().([]byte); ok {
		b[0] = 0x00
		b[1] = 0x00
		k := 2
		for _, v := range raw[:6] {
			b[k] = v
			k++
		}
		rv = binary.BigEndian.Uint64(b)
		eightByte.Put(b)
	}
	return rv
}

func decodeRaw64(raw []byte) uint64 {
	var rv uint64
	if b, ok := eightByte.Get().([]byte); ok {
		copy(b, raw)
		rv = binary.BigEndian.Uint64(b)
		eightByte.Put(b)
	}
	return rv
}

func valueTopBit(in byte) bool {
	if in&TopBitMask != 0 {
		return true
	}
	return false
}

func maskOutTopBit(in byte) byte {
	return in &^ TopBitMask
}

// this decodes a common structure with 1 bit, followed by 47 bits
func decodeRaw1_47Split(raw []byte) (bool, uint64) {
	var rint uint64
	rbool := valueTopBit(raw[0])
	topByte := maskOutTopBit(raw[0])
	buf := bytes.NewBuffer([]byte{0, 0, topByte})
	buf.Write(raw[1:6])
	binary.Read(buf, binary.BigEndian, &rint)
	return rbool, rint
}

func encodeRaw1_47Split(topBit bool, rest uint64) []byte {
	// encode the rest portion first
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, rest)
	bufbytes := buf.Bytes()
	// then overwrite the top bit
	if topBit {
		bufbytes[len(bufbytes)-6] = bufbytes[len(bufbytes)-6] | TopBitMask
	} else {
		bufbytes[len(bufbytes)-6] = bufbytes[len(bufbytes)-6] &^ TopBitMask
	}
	return bufbytes[len(bufbytes)-6:]
}

func decodeRaw12_28Split(data []byte) (top uint32, bottom uint32) {
	FirstByte := (data[0] & 0xf0) >> 4
	SecondByteTop := (data[0] & 0x0f) << 4
	SecondByteBottom := (data[1] & 0xf0) >> 4
	SecondByte := SecondByteTop | SecondByteBottom

	//Using sync.Pool
	if b, ok := fourByte.Get().([]byte); ok {
		b[0] = 0x00
		b[1] = 0x00
		b[2] = FirstByte
		b[3] = SecondByte

		top = binary.BigEndian.Uint32(b)

		fourByte.Put(b)
	}

	if b, ok := fourByte.Get().([]byte); ok {
		b[0] = data[1] & 0x0f
		k := 1
		for _, v := range data[2:] {
			b[k] = v
			k++
		}
		bottom = binary.BigEndian.Uint32(b)
		fourByte.Put(b)
	}

	return
}

func encodeRaw12_28Split(top uint32, bottom uint32) []byte {
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
