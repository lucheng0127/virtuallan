package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type AuthBody struct {
	NLen   uint16
	Name   [16]byte
	PLen   uint16
	Passwd [64]byte
}

func (body *AuthBody) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.BigEndian, body.NLen)
	if err != nil {
		return nil, err
	}

	stream := buf.Bytes()
	stream = append(stream, body.Name[:]...)

	buf = new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, body.PLen)
	if err != nil {
		return nil, err
	}

	stream = append(stream, buf.Bytes()...)
	stream = append(stream, body.Passwd[:]...)

	return stream, nil
}

func (body *AuthBody) Decode(stream []byte) error {
	if len(stream) <= 28 {
		return errors.New("invalidate auth packet")
	}

	body.NLen = binary.BigEndian.Uint16(stream[:2])
	body.Name = [16]byte(stream[2:18])
	body.PLen = binary.BigEndian.Uint16(stream[18:20])
	body.Passwd = [64]byte(stream[20:])

	return nil
}

func NewAuthBody(user, passwd string) *AuthBody {
	body := new(AuthBody)

	ub := []byte(user)
	pb := []byte(passwd)
	body.NLen = uint16(len(ub))
	body.PLen = uint16(len(pb))

	copy(body.Name[:], ub)
	copy(body.Passwd[:], pb)

	return body
}

func (body *AuthBody) Parse() (string, string) {
	return string(body.Name[:body.NLen]), string(body.Passwd[:body.PLen])
}

func NewAuthPkt(user, passwd string) *VLPkt {
	return &VLPkt{
		VLHeader: VLHeader{Type: P_AUTH},
		VLBody:   NewAuthBody(user, passwd),
	}
}
