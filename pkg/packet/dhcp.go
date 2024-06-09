package packet

func NewDhcpPkt(addr string) (*VLPkt, error) {
	b, err := NewKeepaliveBody(addr)
	if err != nil {
		return nil, err
	}

	return &VLPkt{
		VLHeader: VLHeader{Type: P_DHCP},
		VLBody:   b,
	}, nil
}
