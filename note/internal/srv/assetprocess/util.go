package assetprocess

import (
	"net/url"
	"strconv"
)

func encodeCallbackUrl(raw string, noteId int64) string {
	url, _ := url.Parse(raw)
	query := url.Query()
	query.Add("note_id", strconv.FormatInt(noteId, 10))
	url.RawQuery = query.Encode()
	return url.String()
}
