//go:build linux

package packet

import (
	"testing"

	"github.com/lucheng0127/virtuallan/pkg/cipher"
)

func TestAuthPktEncodeAndDecode(t *testing.T) {
	user := "shawn"
	passwd := "secret1111"
	cipher.SetAESKey("0123456789abcdef")

	pkt := &VLPkt{
		VLHeader: VLHeader{Type: P_AUTH},
		VLBody:   NewAuthBody(user, passwd),
	}

	stream, err := pkt.Encode()
	if err != nil {
		t.Error("Encode auth pkt failed")
	}

	aPkt, err := Decode(stream)
	if err != nil {
		t.Error("Decode auth pkt failed")
	}

	u, p := aPkt.VLBody.(*AuthBody).Parse()
	if u != user || p != passwd {
		t.Errorf("Parse auth pkt want: %s, %s\ngot: %s %s\n", user, passwd, u, p)
	}
}
