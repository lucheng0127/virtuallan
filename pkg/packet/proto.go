package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	HeadLen = 16

	P_KEEPALIVE uint16 = 0x1b00 | 0x1
	P_AUTH      uint16 = 0x1b00 | (0x01 << 1)
	P_DATA      uint16 = 0x1b00 | (0x01 << 2)
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

func Decode(stream []byte) (*VLPkt, error) {
	if len(stream) < 16 {
		return nil, errors.New("invalidate vlpkt")
	}

	h := binary.BigEndian.Uint16(stream[:2])

	switch h {
	case P_AUTH:
		b := new(AuthBody)

		err := b.Decode(stream[2:])
		if err != nil {
			return nil, err
		}

		pkt := new(VLPkt)
		pkt.VLHeader.Type = P_AUTH
		pkt.VLBody = b

		return pkt, nil
	default:
		return nil, errors.New("unsupported vl pkt")
	}
}
