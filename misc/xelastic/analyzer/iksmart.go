package xelastic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const (
	IkSmart = "ik_smart"
)

type IkSmartAnalyzer struct {
	Type string `json:"type,omitempty"`
}

func (i *IkSmartAnalyzer) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))

	for {
		t, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		switch t {
		case "type":
			if err := dec.Decode(&i.Type); err != nil {
				return fmt.Errorf("%s | %w", "Type", err)
			}
		}
	}

	return nil
}

func (i IkSmartAnalyzer) MarshalJSON() ([]byte, error) {
	type innerIkSmartAnalyzer IkSmartAnalyzer
	tmp := innerIkSmartAnalyzer{
		Type: i.Type,
	}

	tmp.Type = IkSmart

	return json.Marshal(tmp)
}

func NewIkSmartAnalyzer() *IkSmartAnalyzer {
	return &IkSmartAnalyzer{}
}
