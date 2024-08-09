package server

import (
	"net"
	"reflect"
	"sync"
	"testing"

	"bou.ke/monkey"
	"github.com/lucheng0127/virtuallan/pkg/utils"
)

func TestServer_IPForUser(t *testing.T) {
	type args struct {
		username string
	}
	tests := []struct {
		name       string
		args       args
		want       net.IP
		wantErr    bool
		patchFunc  interface{}
		targetFunc interface{}
	}{
		{
			name:       "idx 1-1",
			args:       args{username: "whocares"},
			want:       net.ParseIP("192.168.123.101").To4(),
			wantErr:    false,
			patchFunc:  utils.IdxFromString,
			targetFunc: func(int, string) int { return 1 },
		},
		{
			name:       "idx 1-2",
			args:       args{username: "whocares"},
			want:       net.ParseIP("192.168.123.102").To4(),
			wantErr:    false,
			patchFunc:  utils.IdxFromString,
			targetFunc: func(int, string) int { return 1 },
		},
		{
			name:       "idx 1-3",
			args:       args{username: "whocares"},
			want:       net.ParseIP("192.168.123.100").To4(),
			wantErr:    false,
			patchFunc:  utils.IdxFromString,
			targetFunc: func(int, string) int { return 1 },
		},
		{
			name:       "idx 1-4",
			args:       args{username: "whocares"},
			want:       nil,
			wantErr:    true,
			patchFunc:  utils.IdxFromString,
			targetFunc: func(int, string) int { return 1 },
		},
	}

	ipStart := net.ParseIP("192.168.123.100").To4()
	svc := &Server{
		UsedIP:  make([]int, 0),
		IPStart: ipStart,
		IPCount: 3,
		MLock:   sync.Mutex{},
		Routes:  make(map[string]string),
	}

	for _, tt := range tests {
		if tt.targetFunc != nil {
			monkey.Patch(tt.patchFunc, tt.targetFunc)
		}

		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.IPForUser(tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.IPForUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.IPForUser() = %v, want %v", got, tt.want)
			}
		})
	}
}
