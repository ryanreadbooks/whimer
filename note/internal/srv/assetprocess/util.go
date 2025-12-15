package assetprocess

import (
	"net/url"
	"strconv"
)

func encodeCallbackUrl(raw string, noteId int64, extra map[string]string) string {
	url, _ := url.Parse(raw)
	query := url.Query()
	query.Add("note_id", strconv.FormatInt(noteId, 10))
	for k, v := range extra {
		query.Add(k, v)
	}
	url.RawQuery = query.Encode()
	return url.String()
}
