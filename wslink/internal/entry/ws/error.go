package ws

import "fmt"

var (
	errUnexpectedProtocol = fmt.Errorf("unexpected protocol")
	errUnexpectedFlag     = fmt.Errorf("unexpected flag")
)
