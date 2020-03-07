package redisgo

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"unsafe"
)

const (
	maxBlukSize = 512 * 1024 * 1024
)

type Decoder struct {
	r *bufio.Reader
}

func (d *Decoder) readLine() ([]byte, error) {
	ln, err := d.r.ReadSlice('\n')
	if err != nil {
		return nil, err
	}
	if len(ln) < 2 || ln[len(ln)-2] != '\r' {
		return nil, fmt.Errorf("expect terminated with CRLF")
	}
	return ln[:len(ln)-2], nil
}

func (d *Decoder) Decode(r *Resp) error {
	ch, err := d.r.ReadByte()
	if err != nil {
		return err
	}
	switch RespKind(ch) {
	case SimpleKind, ErrorKind, IntegerKind:
		ln, err := d.readLine()
		if err != nil {
			return err
		}
		*r = Resp{
			Kind: RespKind(ch),
			Data: string(ln),
		}
	case BlukKind:
		ln, err := d.readLine()
		if err != nil {
			return err
		}
		n, err := strconv.Atoi(*(*string)(unsafe.Pointer(&ln)))
		if err != nil {
			return err
		}
		if n < -1 || n > maxBlukSize {
			return fmt.Errorf("invalid bluk length: %d", n)
		}
		if n == -1 {
			*r = Resp{
				Kind: RespKind(ch),
				Null: true,
			}
		} else {
			data := make([]byte, n+2)
			_, err := io.ReadFull(d.r, data)
			if err != nil {
				return err
			}
			if data[len(data)-2] != '\r' || data[len(data)-1] != '\n' {
				return fmt.Errorf("expect terminated with CRLF")
			}
			*r = Resp{
				Kind: RespKind(ch),
				Data: string(data[:len(data)-2]),
			}
		}
	case ArrayKind:
		ln, err := d.readLine()
		if err != nil {
			return err
		}
		n, err := strconv.Atoi(*(*string)(unsafe.Pointer(&ln)))
		if err != nil {
			return err
		}
		if n < 0 {
			return fmt.Errorf("invalid array length: %d", n)
		}
		array := make([]Resp, n)
		for i := 0; i < n; i++ {
			err = d.Decode(&array[i])
			if err != nil {
				return err
			}
		}
		*r = Resp{
			Kind:  RespKind(ch),
			Array: array,
		}
	default:
		return fmt.Errorf("unrecognized kind: %c", ch)
	}
	return nil
}

func NewDecoder(r io.Reader, maxLineSize int) *Decoder {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReaderSize(r, maxLineSize)
	}
	return &Decoder{
		r: br,
	}
}
