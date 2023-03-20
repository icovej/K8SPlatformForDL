package controller

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Applifier/go-tensorflow/types/tensorflow/core/util"
)

const (
	maskDelta = 0xa282ead8

	headerSize = 12
	footerSize = 4
)

var crc32c = crc32.MakeTable(crc32.Castagnoli)

// 当event file头部或者尾部的检验值非法时的返回值
var ErrInvalidChecksum = errors.New("invalid crc")

// 读取TFEvents的结构体
type Reader struct {
	r   io.Reader
	buf *bytes.Buffer
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		r:   r,
		buf: bytes.NewBuffer(nil),
	}
}

func (r *Reader) Next() (*util.Event, error) {
	f := r.r
	buf := r.buf
	buf.Reset()

	_, err := io.CopyN(buf, f, headerSize)
	if err != nil {
		return nil, err
	}

	header := buf.Bytes()

	crc := binary.LittleEndian.Uint32(header[8:12])
	if !verifyChecksum(header[0:8], crc) {
		return nil, errors.Wrap(ErrInvalidChecksum, "length")
	}

	length := binary.LittleEndian.Uint64(header[0:8])
	buf.Reset()

	if _, err = io.CopyN(buf, f, int64(length)); err != nil {
		return nil, err
	}

	if _, err = io.CopyN(buf, f, footerSize); err != nil {
		return nil, err
	}

	payload := buf.Bytes()

	footer := payload[length:]
	crc = binary.LittleEndian.Uint32(footer)
	if !verifyChecksum(payload[:length], crc) {
		return nil, errors.Wrap(ErrInvalidChecksum, "payload")
	}

	ev := &util.Event{}

	return ev, ev.Unmarshal(payload[0:length])
}

func verifyChecksum(data []byte, crcMasked uint32) bool {
	rot := crcMasked - maskDelta
	unmaskedCrc := ((rot >> 17) | (rot << 15))

	crc := crc32.Checksum(data, crc32c)

	return crc == unmaskedCrc
}

func GetData(c *gin.Context) {

}
