package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/lucheng0127/virtuallan/pkg/utils"
)

type KeepaliveBody struct {
	Len   uint16
	IP    []byte // Store ip/netmask string bytes
	Noise []byte
}

func (body *KeepaliveBody) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.BigEndian, body.Len)
	if err != nil {
		return nil, err
	}

	stream := buf.Bytes()
	stream = append(stream, body.IP...)
	stream = append(stream, body.Noise...)
	return stream, nil
}

func (body *KeepaliveBody) Decode(stream []byte) error {
	if len(stream) < MINI_LEN {
		return errors.New("invalidate keepalive packet")
	}

	body.Len = binary.BigEndian.Uint16(stream[:2])
	body.IP = stream[2 : 2+body.Len]
	return nil
}

func NewKeepaliveBody(addr string) (*KeepaliveBody, error) {
	if !utils.ValidateIPv4WithNetmask(addr) {
		return nil, fmt.Errorf("invalidate ipv4 address %s", addr)
	}

	body := new(KeepaliveBody)
	body.IP = []byte(addr)
	body.Len = uint16(len(body.IP))

	noiseLen := MINI_LEN - 2 - int(body.Len)
	if noiseLen > 0 {
		body.Noise = make([]byte, noiseLen)
	} else {
		body.Noise = make([]byte, 0)
	}

	return body, nil
}

func NewKeepalivePkt(addr string) (*VLPkt, error) {
	b, err := NewKeepaliveBody(addr)
	if err != nil {
		return nil, err
	}

	return &VLPkt{
		VLHeader: VLHeader{Type: P_KEEPALIVE},
		VLBody:   b,
	}, nil
}

func (body *KeepaliveBody) Parse() string {
	return string(body.IP)
}
