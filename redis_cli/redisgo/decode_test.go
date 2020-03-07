package redisgo

import (
	"strconv"
	"strings"
	"testing"
)

func TestDecoder_Decode(t *testing.T) {
	type args struct {
		p *Resp
	}
	tests := []struct {
		name    string
		input   string
		args    args
		wantErr bool
	}{
		{"simple", "+OK\r\n", args{&Resp{}}, false},
		{"error", "-ERR unknown command 'foobar'\r\n", args{&Resp{}}, false},
		{"integer", ":1000\r\n", args{&Resp{}}, false},
		{"bluk", "$7\r\nfoo\nbar\r\n", args{&Resp{}}, false},
		{"array", "*3\r\n$3\r\nfoo\r\n$-1\r\n$3\r\nbar\r\n", args{&Resp{}}, false},
		{"array-in-array", "*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Foo\r\n-Bar\r\n", args{&Resp{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecoder(strings.NewReader(tt.input), 1024)
			if err := d.Decode(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("Decoder.Decode() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				t.Logf("%s", strconv.Quote(tt.args.p.String()))
			}
		})
	}
}
