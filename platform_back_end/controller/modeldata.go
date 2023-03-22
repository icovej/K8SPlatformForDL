package controller

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"io"
	"net/http"
	"os"
	"platform_back_end/tools"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/Applifier/go-tensorflow/types/tensorflow/core/util"
)

const (
	maskDelta = 0xa282ead8

	headerSize  = 12
	footerSize  = 4
	testPrefix  = "test_loss"
	trainPrefix = "train_loss"
	accuracy    = "test_accuracy"

	testLossFile  = "test_loss_all.txt"
	trainLossFile = "train_loss_all.txt"
	accFile       = "acc.txt"
)

var crc32c = crc32.MakeTable(crc32.Castagnoli)

// 当event file头部或者尾部的检验值非法时的返回值
var ErrInvalidChecksum = errors.New("invalid crc")

// 读取TFEvents的结构体
type Reader struct {
	r   io.Reader
	buf *bytes.Buffer
}

type EVData struct {
	Logdir string `json:"logdir"`
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
	var evdata EVData
	err_bind := c.ShouldBindJSON(&evdata)
	if err_bind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err_bind.Error()})
		glog.Error("Failed to parse data form request, the error is %s", err_bind)
		return
	}

	testfile := evdata.Logdir
	f, err_open := os.Open(testfile)
	if err_open != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err_open.Error()})
		glog.Error("Failed to open log dir, the error is ", err_open.Error())
		return
	}
	defer f.Close()

	r := NewReader(f)
	events := make([]*util.Event, 0, 99)

	for {
		ev, err := r.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else if err != nil {
				glog.Error("Failed to read event file, the error is ", err.Error())
			}
		}
		events = append(events, ev)
	}

	test_loss, _ := os.Create(testLossFile)
	defer test_loss.Close()

	train_loss, _ := os.Create(trainLossFile)
	defer train_loss.Close()

	acc, _ := os.Create(accFile)
	defer acc.Close()

	for i := range events {
		s := events[i].GetSummary()
		if s != nil {
			for j := range s.Value {
				tag := s.Value[j].Tag
				if strings.HasPrefix(tag, testPrefix) {
					_, err := test_loss.WriteString(tag + " " + tools.FloatToString(s.Value[j].GetSimpleValue()) + "\n")
					if err != nil {
						glog.Error("Failed to get test loss value from event, the error is ", err.Error())
					}
				} else if strings.HasPrefix(tag, trainPrefix) {
					_, err := train_loss.WriteString(tag + " " + tools.FloatToString(s.Value[j].GetSimpleValue()) + "\n")
					if err != nil {
						glog.Error("Failed to get train loss value from event, the error is ", err.Error())
					}
				} else if strings.HasPrefix(tag, accuracy) {
					_, err := acc.WriteString(tag + " " + tools.FloatToString(s.Value[j].GetSimpleValue()) + "\n")
					if err != nil {
						glog.Error("Failed to get acc value from event, the error is ", err.Error())
					}
				}
			}
		}
	}

	err := tools.CalculateAvg(testLossFile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		glog.Error("Failed to calculate test loss avg, the err is ", err.Error())
		return
	}
	err = tools.CalculateAvg(trainLossFile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		glog.Error("Failed to calculate train loss avg, the err is ", err.Error())
		return
	}

}
