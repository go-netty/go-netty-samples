package redisgo

import (
	"bytes"
	"testing"
)

func TestEncode(t *testing.T) {
	type args struct {
		v Value
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{"null", args{Null()}, "$-1\r\n", false},
		{"simple", args{Simple("hello")}, "+hello\r\n", false},
		{"error", args{Error("ERR oops")}, "-ERR oops\r\n", false},
		{"bluk", args{Bluk([]byte("hello"))}, "$5\r\nhello\r\n", false},
		{"int", args{Int(233)}, ":233\r\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if err := Encode(w, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Encode() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestEncodeMulti(t *testing.T) {
	type args struct {
		vals []Value
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{"multi", args{[]Value{Null(), Simple("hello"), Int(233)}}, "*3\r\n$-1\r\n+hello\r\n:233\r\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if err := EncodeMulti(w, tt.args.vals...); (err != nil) != tt.wantErr {
				t.Errorf("EncodeMulti() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("EncodeMulti() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestEncodeResp(t *testing.T) {
	type args struct {
		r *Resp
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{"array-in-array", args{&Resp{
			Kind: ArrayKind,
			Array: []Resp{
				{
					Kind: ArrayKind,
					Array: []Resp{
						Resp{
							Kind: IntegerKind,
							Data: "1",
						},
						Resp{
							Kind: IntegerKind,
							Data: "2",
						},
						Resp{
							Kind: IntegerKind,
							Data: "3",
						},
					},
				},
				{
					Kind: ArrayKind,
					Array: []Resp{
						{
							Kind: SimpleKind,
							Data: "Foo",
						},
						{
							Kind: ErrorKind,
							Data: "Bar",
						},
					},
				},
			},
		}}, "*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Foo\r\n-Bar\r\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if err := EncodeResp(w, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("EncodeResp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("EncodeResp() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
