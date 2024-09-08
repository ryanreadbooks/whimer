package xsql

import "fmt"

type SQLState string

var (
	SQLStateOutOfRange SQLState = "22003" // [5]uint8{50, 50, 48, 48, 51}
)

func SQLStateEqual(src [5]byte, target SQLState) bool {
	return fmt.Sprintf("%s", src) == string(target)
}
