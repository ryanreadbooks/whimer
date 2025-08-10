package xelastic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const (
	IkMaxWord = "ik_max_word"
)

type IkMaxWordAnalyzer struct {
	Type string `json:"type,omitempty"`
}

func (i *IkMaxWordAnalyzer) UnmarshalJSON(data []byte) error {
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

func (i IkMaxWordAnalyzer) MarshalJSON() ([]byte, error) {
	type innerIkMaxWordAnalyzer IkMaxWordAnalyzer
	tmp := innerIkMaxWordAnalyzer{
		Type: i.Type,
	}

	tmp.Type = IkMaxWord

	return json.Marshal(tmp)
}

func NewIkMaxWordAnalyzer() *IkMaxWordAnalyzer {
	return &IkMaxWordAnalyzer{}
}
