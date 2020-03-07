package redisgo

import (
	"fmt"
	"io"
	"strconv"
	"unsafe"
)

const (
	blukchar    = '$'
	integerchar = ':'
	arraychar   = '*'
)

var (
	crlf       = []byte{'\r', '\n'}
	simplechar = []byte{'+'}
	errorchar  = []byte{'-'}
	nullbluk   = []byte("$-1\r\n")
)

type kind uint

const (
	nullKind kind = iota
	simpleKind
	errorKind
	blukKind
	intKind
	int8Kind
	int16Kind
	int32Kind
	int64Kind
	uintKind
	uint8Kind
	uint16Kind
	uint32Kind
	uint64Kind
)

type Value struct {
	kind kind
	u64  uint64
	s    string
}

func Null() Value {
	return Value{kind: nullKind}
}

func Simple(s string) Value {
	return Value{kind: simpleKind, s: s}
}

func Error(err string) Value {
	return Value{kind: errorKind, s: err}
}

func BlukString(s string) Value {
	return Value{kind: blukKind, s: s}
}

func Bluk(p []byte) Value {
	return BlukString(*(*string)(unsafe.Pointer(&p)))
}

func Int(v int) Value {
	return Value{kind: intKind, u64: uint64(v)}
}

func Int8(v int8) Value {
	return Value{kind: int8Kind, u64: uint64(v)}
}

func Int16(v int16) Value {
	return Value{kind: int16Kind, u64: uint64(v)}
}

func Int32(v int32) Value {
	return Value{kind: int32Kind, u64: uint64(v)}
}

func Int64(v int64) Value {
	return Value{kind: int64Kind, u64: uint64(v)}
}

func Uint(v uint) Value {
	return Value{kind: uintKind, u64: uint64(v)}
}

func Uint8(v uint8) Value {
	return Value{kind: uint8Kind, u64: uint64(v)}
}

func Uint16(v uint16) Value {
	return Value{kind: uint16Kind, u64: uint64(v)}
}

func Uint32(v uint32) Value {
	return Value{kind: uint32Kind, u64: uint64(v)}
}

func Uint64(v uint64) Value {
	return Value{kind: uint64Kind, u64: uint64(v)}
}

type slice struct {
	data uintptr
	n1   int
	n2   int
}

func sb(s string) []byte {
	b := slice{
		data: *(*uintptr)(unsafe.Pointer(&s)),
		n1:   len(s),
		n2:   len(s),
	}
	return *(*[]byte)(unsafe.Pointer(&b))
}

func Encode(w io.Writer, v Value) (err error) {
	var buf [32]byte
	b := crlf
	switch v.kind {
	case nullKind:
		b = nullbluk
	case simpleKind:
		_, err = w.Write(simplechar)
		if err != nil {
			return
		}
		_, err = w.Write(sb(v.s))
		if err != nil {
			return
		}
	case errorKind:
		_, err = w.Write(errorchar)
		if err != nil {
			return
		}
		_, err = w.Write(sb(v.s))
		if err != nil {
			return
		}
	case blukKind:
		t := append(buf[:0], blukchar)
		t = strconv.AppendInt(t, int64(len(v.s)), 10)
		t = append(t, '\r', '\n')
		_, err = w.Write(t)
		if err != nil {
			return
		}
		_, err = w.Write(sb(v.s))
		if err != nil {
			return
		}
	case intKind:
		b = append(buf[:0], integerchar)
		b = strconv.AppendInt(b, int64(int(v.u64)), 10)
		b = append(b, '\r', '\n')
	case int8Kind:
		b = append(buf[:0], integerchar)
		b = strconv.AppendInt(b, int64(int8(v.u64)), 10)
		b = append(b, '\r', '\n')
	case int16Kind:
		b = append(buf[:0], integerchar)
		b = strconv.AppendInt(b, int64(int16(v.u64)), 10)
		b = append(b, '\r', '\n')
	case int32Kind:
		b = append(buf[:0], integerchar)
		b = strconv.AppendInt(b, int64(int32(v.u64)), 10)
		b = append(b, '\r', '\n')
	case int64Kind:
		b = append(buf[:0], integerchar)
		b = strconv.AppendInt(b, int64(int64(v.u64)), 10)
		b = append(b, '\r', '\n')
	case uintKind:
		b = append(buf[:0], integerchar)
		b = strconv.AppendUint(b, uint64(uint(v.u64)), 10)
		b = append(b, '\r', '\n')
	case uint8Kind:
		b = append(buf[:0], integerchar)
		b = strconv.AppendUint(b, uint64(uint8(v.u64)), 10)
		b = append(b, '\r', '\n')
	case uint16Kind:
		b = append(buf[:0], integerchar)
		b = strconv.AppendUint(b, uint64(uint16(v.u64)), 10)
		b = append(b, '\r', '\n')
	case uint32Kind:
		b = append(buf[:0], integerchar)
		b = strconv.AppendUint(b, uint64(uint32(v.u64)), 10)
		b = append(b, '\r', '\n')
	case uint64Kind:
		b = append(buf[:0], integerchar)
		b = strconv.AppendUint(b, uint64(uint64(v.u64)), 10)
		b = append(b, '\r', '\n')
	default:
		panic("never get here")
	}
	_, err = w.Write(b)
	return
}

func EncodeMulti(w io.Writer, vals ...Value) error {
	var buf [32]byte
	t := append(buf[:0], arraychar)
	t = strconv.AppendInt(t, int64(len(vals)), 10)
	t = append(t, '\r', '\n')
	_, err := w.Write(t)
	if err != nil {
		return err
	}
	for _, v := range vals {
		err = Encode(w, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func EncodeResp(w io.Writer, r *Resp) (err error) {
	var buf [32]byte
	b := crlf
	switch r.Kind {
	case SimpleKind, ErrorKind, IntegerKind:
		if 3+len(r.Data) <= 32 {
			t := append(buf[:0], byte(r.Kind))
			t = append(t, r.Data...)
			b = append(t, '\r', '\n')
		} else {
			_, err = w.Write([]byte{byte(r.Kind)})
			if err != nil {
				return
			}
			_, err = w.Write(sb(r.Data))
			if err != nil {
				return
			}
		}
	case BlukKind:
		if r.Null {
			b = nullbluk
		} else {
			t := append(buf[:0], '$')
			t = strconv.AppendInt(t, int64(len(r.Data)), 10)
			t = append(t, '\r', '\n')
			_, err = w.Write(t)
			if err != nil {
				return
			}
			_, err = w.Write(sb(r.Data))
			if err != nil {
				return
			}
		}
	case ArrayKind:
		t := append(buf[:0], '*')
		t = strconv.AppendInt(t, int64(len(r.Array)), 10)
		t = append(t, '\r', '\n')
		_, err = w.Write(t)
		if err != nil {
			return
		}
		for i := range r.Array {
			err = EncodeResp(w, &r.Array[i])
			if err != nil {
				return
			}
		}
		return
	default:
		return fmt.Errorf("unrecognized kind: %c", r.Kind)
	}
	_, err = w.Write(b)
	return
}
