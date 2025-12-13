package assetprocess

import (
	"net/url"
	"testing"
)

func TestGetCallbackUrl(t *testing.T) {
	myUrl := "http://localhost:8090/api/v1/callback/devtest"
	url, err := url.Parse(myUrl)
	t.Log(err)

	vals := url.Query()
	vals.Add("noteId", "100")
	vals.Add("taskId", "1234567890")
	vals.Add("namespace", "& &wole")
	url.RawQuery = vals.Encode()

	callbackUrl := url.String()
	t.Log(callbackUrl)
}

func TestEncodeCallbackUrl(t *testing.T) {
	rawUrl := "http://localhost:8090/api/v1/callback/devtest"
	noteId := int64(100)
	callbackUrl := encodeCallbackUrl(rawUrl, noteId)
	t.Log(callbackUrl)
}