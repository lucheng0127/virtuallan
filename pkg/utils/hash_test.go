package utils

import "testing"

func TestIdxFromString(t *testing.T) {
	type args struct {
		step int
		str  string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "step 100 user1",
			want: 32,
			args: args{
				step: 100,
				str:  "shawn",
			},
		},
		{
			name: "step 100 user2",
			want: 59,
			args: args{
				step: 100,
				str:  "guest",
			},
		},
		{
			name: "step 30 user1",
			want: 2,
			args: args{
				step: 30,
				str:  "shawn",
			},
		},
		{
			name: "step 30 user2",
			want: 9,
			args: args{
				step: 30,
				str:  "guest",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IdxFromString(tt.args.step, tt.args.str); got != tt.want {
				t.Errorf("IdxFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
