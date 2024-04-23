package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type RspBody struct {
	Code  uint16
	Noise []byte
}

func (body *RspBody) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.BigEndian, body.Code)
	if err != nil {
		return nil, err
	}

	stream := buf.Bytes()
	stream = append(stream, body.Noise...)
	return stream, nil
}

func (body *RspBody) Decode(stream []byte) error {
	if len(stream) < MINI_LEN {
		return errors.New("invalidate response packet")
	}

	body.Code = binary.BigEndian.Uint16(stream[:2])
	return nil
}

func NewResponseBody(code uint16) *RspBody {
	body := new(RspBody)
	body.Code = code
	noiseLen := MINI_LEN - 2
	body.Noise = make([]byte, noiseLen)

	return body
}

func NewResponsePkt(code uint16) *VLPkt {
	b := NewResponseBody(code)

	return &VLPkt{
		VLHeader: VLHeader{Type: P_RESPONSE},
		VLBody:   b,
	}
}
