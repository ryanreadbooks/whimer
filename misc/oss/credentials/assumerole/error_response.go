package assumerole

import (
	"bytes"
	"encoding/xml"
	"io"
)

// xmlDecoder provide decoded value in xml.
func xmlDecoder(body io.Reader, v any) error {
	d := xml.NewDecoder(body)
	return d.Decode(v)
}

// xmlDecodeAndBody reads the whole body up to 1MB and
// tries to XML decode it into v.
// The body that was read and any error from reading or decoding is returned.
func xmlDecodeAndBody(bodyReader io.Reader, v interface{}) ([]byte, error) {
	// read the whole body (up to 1MB)
	const maxBodyLength = 1 << 20
	body, err := io.ReadAll(io.LimitReader(bodyReader, maxBodyLength))
	if err != nil {
		return nil, err
	}
	return bytes.TrimSpace(body), xmlDecoder(bytes.NewReader(body), v)
}
