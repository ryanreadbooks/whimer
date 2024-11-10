package model_test

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
)

func TestSession_Serializer(t *testing.T) {
	sess := model.NewSession(utils.RandomString(64), time.Now().Unix())
	sess.Detail = utils.RandomString(256)

	jsonSer := model.JsonSessionSerializer{}
	b, _ := jsonSer.Serialize(sess)
	t.Log(len(b), len(base64.StdEncoding.EncodeToString(b)))

	msgSer := model.MsgpackSessionSerializer{}
	b, _ = msgSer.Serialize(sess)
	t.Log(len(b), len(base64.StdEncoding.EncodeToString(b)))
}

func BenchmarkJsonSerializer(b *testing.B) {
	sess := model.NewSession(utils.RandomString(64), time.Now().Unix())
	sess.Detail = utils.RandomString(256)
	ser := model.JsonSessionSerializer{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ser.Serialize(sess)
	}
}

func BenchmarkMsgpackSerializer(b *testing.B) {
	sess := model.NewSession(utils.RandomString(64), time.Now().Unix())
	sess.Detail = utils.RandomString(256)
	ser := model.MsgpackSessionSerializer{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ser.Serialize(sess)
	}
}