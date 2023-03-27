package controller

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"
	"platform_back_end/data"
	"platform_back_end/tools"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/Applifier/go-tensorflow/types/tensorflow/core/util"
)

type Reader data.Reader

func NewReader(r io.Reader) *Reader {
	return &Reader{
		R:   r,
		Buf: bytes.NewBuffer(nil),
	}
}

func (r *Reader) Next() (*util.Event, error) {
	f := r.R
	buf := r.Buf
	buf.Reset()

	_, err := io.CopyN(buf, f, data.HeaderSize)
	if err != nil {
		glog.Error("Failed to copy the useless info of event file")
		return nil, err
	}

	header := buf.Bytes()

	crc := binary.LittleEndian.Uint32(header[8:12])
	if !tools.VerifyChecksum(header[0:8], crc) {
		glog.Error("Failed to check the crc of event file, the error is %v", data.ErrInvalidChecksum.Error())
		return nil, errors.Wrap(data.ErrInvalidChecksum, "length")
	}

	length := binary.LittleEndian.Uint64(header[0:8])
	buf.Reset()

	if _, err = io.CopyN(buf, f, int64(length)); err != nil {
		glog.Error("Failed to copy the header of event file")
		return nil, err
	}

	if _, err = io.CopyN(buf, f, data.FooterSize); err != nil {
		glog.Error("Failed to copy the tail of event file")
		return nil, err
	}

	payload := buf.Bytes()

	footer := payload[length:]
	crc = binary.LittleEndian.Uint32(footer)
	if !tools.VerifyChecksum(payload[:length], crc) {
		glog.Error("Failed to check the crc of event file, the error is %v", data.ErrInvalidChecksum.Error())
		return nil, errors.Wrap(data.ErrInvalidChecksum, "payload")
	}

	ev := &util.Event{}

	return ev, ev.Unmarshal(payload[0:length])
}

func GetData(c *gin.Context) {
	var evdata data.EVData
	err_bind := c.ShouldBindJSON(&evdata)
	if err_bind != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code: ":    http.StatusBadRequest,
			"message: ": fmt.Sprintf("Invalid request payload, err is %v", err_bind.Error()),
		})
		glog.Error("Method GetData gets invalid request payload")
		return
	}

	testfile := evdata.Logdir
	f, err_open := os.Open(testfile)
	if err_open != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code:":   http.StatusMethodNotAllowed,
			"message": err_open.Error(),
		})
		glog.Error("Failed to open log dir, the error is %v", err_open.Error())
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
				glog.Error("Failed to read event file, the error is %v", err.Error())
			}
		}
		events = append(events, ev)
	}

	test_loss, _ := os.Create(data.TestLossFile)
	defer test_loss.Close()

	train_loss, _ := os.Create(data.TrainLossFile)
	defer train_loss.Close()

	acc, _ := os.Create(data.AccFile)
	defer acc.Close()

	for i := range events {
		s := events[i].GetSummary()
		if s != nil {
			for j := range s.Value {
				tag := s.Value[j].Tag
				if strings.HasPrefix(tag, data.TestPrefix) {
					_, err := test_loss.WriteString(tag + " " + tools.FloatToString(s.Value[j].GetSimpleValue()) + "\n")
					if err != nil {
						glog.Error("Failed to get test loss value from event, the error is %v", err.Error())
					}
				} else if strings.HasPrefix(tag, data.TrainPrefix) {
					_, err := train_loss.WriteString(tag + " " + tools.FloatToString(s.Value[j].GetSimpleValue()) + "\n")
					if err != nil {
						glog.Error("Failed to get train loss value from event, the error is %v", err.Error())
					}
				} else if strings.HasPrefix(tag, data.Accuracy) {
					_, err := acc.WriteString(tag + " " + tools.FloatToString(s.Value[j].GetSimpleValue()) + "\n")
					if err != nil {
						glog.Error("Failed to get acc value from event, the error is %v", err.Error())
					}
				}
			}
		}
	}

	err := tools.CalculateAvg(data.TestLossFile)
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":   http.StatusMethodNotAllowed,
			"message:": err.Error(),
		})
		glog.Error("Failed to calculate test loss avg, the err is %v", err.Error())
		return
	}
	err = tools.CalculateAvg(data.TrainLossFile)
	if err != nil {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"code: ":   http.StatusMethodNotAllowed,
			"message:": err.Error(),
		})
		glog.Error("Failed to calculate train loss avg, the err is %v", err.Error())
		return
	}

}
