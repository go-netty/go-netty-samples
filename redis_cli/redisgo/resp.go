package redisgo

import "bytes"

type RespKind byte

const (
	SimpleKind  RespKind = '+'
	ErrorKind            = '-'
	IntegerKind          = ':'
	BlukKind             = '$'
	ArrayKind            = '*'
)

type Resp struct {
	Kind  RespKind
	Null  bool
	Data  string
	Array []Resp
}

func (r *Resp) String() string {
	w := &bytes.Buffer{}
	EncodeResp(w, r)
	return w.String()
}
