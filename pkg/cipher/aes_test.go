package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"errors"
	"io"
	"reflect"
	"strconv"
	"testing"

	"bou.ke/monkey"
)

func TestAESCipher(t *testing.T) {
	key16 := "0123456789abcdef"
	hexData, _ := hex.DecodeString("130913f")
	ac, _ := NewAESCipher(key16)

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "test string",
			data:    []byte("lucheng test"),
			wantErr: false,
		},
		{
			name:    "test hex",
			data:    hexData,
			wantErr: false,
		},
		{
			name:    "test int",
			data:    []byte(strconv.Itoa(19960127)),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encodeData, err := ac.Encrypt(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("AESCipher.Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			decodeData, err := ac.Decrypt(encodeData)
			if (err != nil) != tt.wantErr {
				t.Errorf("AESCipher.Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(decodeData, tt.data) {
				t.Errorf("AESCipher decrypt encrypt data not match")
			}
		})
	}
}

func TestAESCipher_Encrypt(t *testing.T) {
	key16 := "0123456789abcdef"
	type fields struct {
		key string
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       []byte
		wantErr    bool
		targetFunc interface{}
		patchFunc  interface{}
	}{
		{
			name: "ERR: io",
			fields: fields{
				key: key16,
			},
			args: args{
				[]byte("io"),
			},
			want:       make([]byte, 0),
			wantErr:    true,
			targetFunc: io.ReadFull,
			patchFunc: func(io.Reader, []byte) (int, error) {
				return 0, errors.New("io err")
			},
		},
		{
			name: "ERR: cipher",
			fields: fields{
				key: key16,
			},
			args: args{
				[]byte("cipher"),
			},
			want:       make([]byte, 0),
			wantErr:    true,
			targetFunc: aes.NewCipher,
			patchFunc: func([]byte) (cipher.Block, error) {
				return nil, errors.New("cipher err")
			},
		},
		{
			name: "ERR: empty data",
			fields: fields{
				key: key16,
			},
			args: args{
				make([]byte, 0),
			},
			want:    make([]byte, 0),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.patchFunc != nil {
				monkey.Patch(tt.targetFunc, tt.patchFunc)
			}

			ac := &AESCipher{
				key: tt.fields.key,
			}
			got, err := ac.Encrypt(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("AESCipher.Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AESCipher.Encrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAESCipher_Decrypt(t *testing.T) {
	key16 := "0123456789abcdef"
	type fields struct {
		key string
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       []byte
		wantErr    bool
		targetFunc interface{}
		patchFunc  interface{}
	}{
		{
			name: "ERR: cipher",
			fields: fields{
				key: key16,
			},
			args: args{
				[]byte("any string but large then block size"),
			},
			want:       make([]byte, 0),
			wantErr:    true,
			targetFunc: aes.NewCipher,
			patchFunc: func([]byte) (cipher.Block, error) {
				return nil, errors.New("cipher err")
			},
		},
		{
			name: "ERR: wrong data length",
			fields: fields{
				key: key16,
			},
			args: args{
				make([]byte, 0),
			},
			want:       make([]byte, 0),
			wantErr:    true,
			targetFunc: nil,
			patchFunc:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.patchFunc != nil {
				monkey.Patch(tt.targetFunc, tt.patchFunc)
			}

			ac := &AESCipher{
				key: tt.fields.key,
			}
			got, err := ac.Decrypt(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("AESCipher.Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AESCipher.Decrypt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAESCipher(t *testing.T) {
	key16 := "0123456789abcdef"
	key24 := "0123456789abcdef01234567"
	key32 := "0123456789abcdef0123456789abcdef"
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    Cipher
		wantErr bool
	}{
		{
			name:    "16",
			args:    args{key: key16},
			want:    &AESCipher{key: key16},
			wantErr: false,
		},
		{
			name:    "24",
			args:    args{key: key24},
			want:    &AESCipher{key: key24},
			wantErr: false,
		},
		{
			name:    "32",
			args:    args{key: key32},
			want:    &AESCipher{key: key32},
			wantErr: false,
		},
		{
			name:    "ERR: invalid key",
			args:    args{key: "123"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAESCipher(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAESCipher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAESCipher() = %v, want %v", got, tt.want)
			}
		})
	}
}
