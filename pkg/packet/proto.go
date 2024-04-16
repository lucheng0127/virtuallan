package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	P_KEEPALIVE uint16 = 0x1b00 | 0x1
	P_AUTH      uint16 = 0x1b00 | (0x01 << 1)
	P_RAW       uint16 = 0x1b00 | (0x01 << 2)
)

type VLHeader struct {
	Type uint16
}

type VLBody interface {
	Encode() ([]byte, error)
	Decode([]byte) error
}

type VLPkt struct {
	VLHeader
	VLBody
}

// Encode vlpkt into byte array
func (pkt *VLPkt) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.BigEndian, pkt.Type)
	if err != nil {
		return nil, err
	}

	bodyBytes, err := pkt.VLBody.Encode()
	if err != nil {
		return nil, err
	}

	return append(buf.Bytes(), bodyBytes...), nil
}

// Decode byte array into vlpkt
func Decode(stream []byte) (*VLPkt, error) {
	if len(stream) < 2 {
		return nil, errors.New("invalidate vlpkt")
	}

	h := binary.BigEndian.Uint16(stream[:2])
	pkt := new(VLPkt)

	switch h {
	case P_AUTH:
		b := new(AuthBody)

		err := b.Decode(stream[2:])
		if err != nil {
			return nil, err
		}

		pkt.VLHeader.Type = P_AUTH
		pkt.VLBody = b

		return pkt, nil
	case P_RAW:
		b := new(RawBody)

		err := b.Decode(stream[2:])
		if err != nil {
			return nil, err
		}

		pkt.VLHeader.Type = P_RAW
		pkt.VLBody = b

		return pkt, nil
	default:
		return nil, errors.New("unsupported vl pkt")
	}
}
