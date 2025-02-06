package imgproxy

import (
	"encoding/hex"
	"fmt"
)

type Auth struct {
	Key  string `json:"key"`
	Salt string `json:"salt"`

	keyBin  []byte `json:"-" yaml:"-"`
	saltBin []byte `json:"-" yaml:"-"`
}

func (c *Auth) GetKey() []byte {
	return c.keyBin
}

func (c *Auth) GetSalt() []byte {
	return c.saltBin
}

func (c *Auth) Init() error {
	var err error
	c.keyBin, err = hex.DecodeString(c.Key)
	if err != nil {
		return fmt.Errorf("img proxy auth key is invalid: %w", err)
	}

	c.saltBin, err = hex.DecodeString(c.Salt)
	if err != nil {
		return fmt.Errorf("img proxy auth salt is invalid: %w", err)
	}

	return nil
}
