package packet

func NewFinPKt(addr string) (*VLPkt, error) {
	b, err := NewKeepaliveBody(addr)
	if err != nil {
		return nil, err
	}

	return &VLPkt{
		VLHeader: VLHeader{Type: P_FIN},
		VLBody:   b,
	}, nil
}
