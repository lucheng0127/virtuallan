package packet

type RawBody struct {
	Payload []byte
}

func (body *RawBody) Encode() ([]byte, error) {
	return body.Payload, nil
}

func (body *RawBody) Decode(stream []byte) error {
	body.Payload = stream
	return nil
}

func NewRawPkt(stream []byte) *VLPkt {
	return &VLPkt{
		VLHeader: VLHeader{Type: P_RAW},
		VLBody:   &RawBody{Payload: stream},
	}
}
