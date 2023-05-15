package controller

import (
	"PlatformBackEnd/data"
	"PlatformBackEnd/tools"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"
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
		glog.Errorf("Failed to check the crc of event file, the error is %v", data.ErrInvalidChecksum.Error())
		return nil, errors.Wrap(data.ErrInvalidChecksum, "length")
	}

	length := binary.LittleEndian.Uint64(header[0:8])
	buf.Reset()

	if _, err = io.CopyN(buf, f, int64(length)); err != nil {
		glog.Errorf("Failed to copy the header of event file, the error is %v", err.Error())
		return nil, err
	}

	if _, err = io.CopyN(buf, f, data.FooterSize); err != nil {
		glog.Errorf("Failed to copy the tail of event file, the error is %v", err.Error())
		return nil, err
	}

	payload := buf.Bytes()

	footer := payload[length:]
	crc = binary.LittleEndian.Uint32(footer)
	if !tools.VerifyChecksum(payload[:length], crc) {
		glog.Errorf("Failed to check the crc of event file, the error is %v", data.ErrInvalidChecksum.Error())
		return nil, errors.Wrap(data.ErrInvalidChecksum, "payload")
	}

	ev := &util.Event{}

	return ev, ev.Unmarshal(payload[0:length])
}

func GetData(c *gin.Context) {
	var evdata data.EVData
	err := c.ShouldBindJSON(&evdata)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.API_PARAMETER_ERROR,
			"message": fmt.Sprintf("Method GetData gets invalid request payload, err is %v", err.Error()),
		})
		glog.Errorf("Method GetData gets invalid request payload, the error is %v", err.Error())
		return
	}

	j := tools.NewJWT()
	tokenString := c.GetHeader("token")
	if tokenString == "" {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
			"message": "Failed to get token, because the token is empty!",
		})
		glog.Error("Failed to get token, because the token is empty!")
		return
	}
	token, err := j.Parse_Token(tokenString)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
			"message": fmt.Sprintf("Failed to parse token, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to parse token, the error is %v", err.Error())
		return
	}

	logdir := token.Path + "/log/" + evdata.Logdir

	f, err := os.Open(logdir)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to open log dir, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to open log dir, the error is %v", err.Error())
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
				glog.Errorf("Failed to read event file, the error is %v", err.Error())
			}
		}
		events = append(events, ev)
	}

	fullpath := token.Path + "/" + "model_data/" + evdata.Logdir + "/"
	err = tools.CreatePath(fullpath, 0777)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.SUCCESS,
			"message": fmt.Sprintf("Failed to create file model_data under workdir, the error is %v", err.Error()),
		})
		glog.Errorf("Failed to create file model_data under workdir, the error is %v", err.Error())
		return
	}

	test_loss, _ := os.Create(fullpath + data.TestLossFile)
	defer test_loss.Close()

	train_loss, _ := os.Create(fullpath + data.TrainLossFile)
	defer train_loss.Close()

	acc, _ := os.Create(fullpath + data.AccFile)
	defer acc.Close()

	for i := range events {
		s := events[i].GetSummary()
		if s != nil {
			for j := range s.Value {
				tag := s.Value[j].Tag
				if strings.HasPrefix(tag, data.TestPrefix) {
					_, err := test_loss.WriteString(tag + " " + tools.FloatToString(s.Value[j].GetSimpleValue()) + "\n")
					if err != nil {
						glog.Errorf("Failed to get test loss value from event, the error is %v", err.Error())
					}
				} else if strings.HasPrefix(tag, data.TrainPrefix) {
					_, err := train_loss.WriteString(tag + " " + tools.FloatToString(s.Value[j].GetSimpleValue()) + "\n")
					if err != nil {
						glog.Errorf("Failed to get train loss value from event, the error is %v", err.Error())
					}
				} else if strings.HasPrefix(tag, data.Accuracy) {
					_, err := acc.WriteString(tag + " " + tools.FloatToString(s.Value[j].GetSimpleValue()) + "\n")
					if err != nil {
						glog.Errorf("Failed to get acc value from event, the error is %v", err.Error())
					}
				}
			}
		}
	}

	err = tools.CalculateAvg(fullpath + data.TestLossFile)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to calculate test loss avg, the err is %v", err.Error()),
		})
		glog.Errorf("Failed to calculate test loss avg, the err is %v", err.Error())
		return
	}
	err = tools.CalculateAvg(fullpath + data.TrainLossFile)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    data.OPERATION_FAILURE,
			"message": fmt.Sprintf("Failed to calculate train loss avg, the err is %v", err.Error()),
		})
		glog.Errorf("Failed to calculate train loss avg, the err is %v", err.Error())
		return
	}

	acc_json := tools.TxtToJson(fullpath + data.AccFile)
	test_json := tools.TxtToJson(fullpath + data.TestLossFile)
	train_json := tools.TxtToJson(fullpath + data.TrainLossFile)

	data_jSON := data.MyJSON{
		JSON1: acc_json,
		JSON2: test_json,
		JSON3: train_json,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": data.SUCCESS,
		"data": data_jSON,
	})
}
