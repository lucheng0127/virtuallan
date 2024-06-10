package packet

func NewFinPkt() *VLPkt {
	b := NewRawPkt([]byte("time to say goodbye, see you next time."))
	return &VLPkt{
		VLHeader: VLHeader{Type: P_FIN},
		VLBody:   b,
	}
}
