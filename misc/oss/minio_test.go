package oss

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/policy"
	"github.com/minio/minio-go/v7/pkg/set"
)

var (
	testCli *minio.Client
)

func TestMain(m *testing.M) {
	keyId := os.Getenv("AWS_ACCESS_KEY_ID")
	sk := os.Getenv("AWS_SECRET_ACCESS_KEY")

	var err error
	testCli, err = minio.New("127.0.0.1:9000", &minio.Options{
		Creds: credentials.NewStaticV4(keyId, sk, ""),
	})
	if err != nil {
		panic(err)
	}

	m.Run()
}

func TestPresignPostPolicy(t *testing.T) {
	policy := minio.NewPostPolicy()
	err := policy.SetContentType("image/jpg")
	if err != nil {
		t.Fatal(err)
	}
	err = policy.SetContentLengthRange(1, 1024*1024)
	if err != nil {
		t.Fatal(err)
	}
	err = policy.SetExpires(time.Now().Add(time.Minute * 5))
	if err != nil {
		t.Fatal(err)
	}

	policy.SetBucket("testdev-bucket")
	policy.SetKey("helloworld")

	t.Log(policy.String())
	url, form, err := testCli.PresignedPostPolicy(t.Context(), policy)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("curl ")
	for k, v := range form {
		fmt.Printf("-F %s=%s ", k, v)
	}
	fmt.Printf("-F file=@/etc/bash.bashrc ")
	fmt.Printf("%s\n", url)

	// url, err = cli.PresignedPutObject(t.Context(), "testdev-bucket", "meme", time.Minute*5)

	// if err != nil {
	// 	t.Fatal(err)
	// }

	// t.Log(url)
}

func TestSts(t *testing.T) {
	st := policy.Statement{
		Resources:  set.CreateStringSet(),
		Actions:    set.CreateStringSet(),
		Effect:     "Allow",
		Conditions: make(policy.ConditionMap),
	}
	st.Resources.Add("arn:aws:s3:::testdev-bucket/yesandno")
	st.Actions.Add("s3:PutObject")
	bp := policy.BucketAccessPolicy{
		Version:    "2012-10-17",
		Statements: []policy.Statement{st},
	}
	policyBytes, _ := json.MarshalIndent(bp, " ", " ")
	t.Log(string(policyBytes))
	cred, err := credentials.NewSTSAssumeRole("http://127.0.0.1:9000",
		credentials.STSAssumeRoleOptions{
			AccessKey: "whimer",
			SecretKey: "whimer-secret",
			Policy:    string(policyBytes),
		})

	if err != nil {
		t.Fatal(err)
	}

	val, err := cred.Get()
	if err != nil {
		t.Fatal(err)
	}

	vv, _ := json.MarshalIndent(val, " ", " ")
	t.Log(string(vv))

	uploadCli, err := minio.New("127.0.0.1:9000", &minio.Options{
		Creds: credentials.NewStaticV4(val.AccessKeyID, val.SecretAccessKey, val.SessionToken),
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := uploadCli.FPutObject(t.Context(), "testdev-bucket", "yesandno", "./conf.go", minio.PutObjectOptions{})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("out = %v\n", out)

	out, err = uploadCli.FPutObject(t.Context(), "testdev-bucket", "yesandno2", "./utils.go", minio.PutObjectOptions{})
	t.Log(err)
	t.Logf("out = %v\n", out)
}
