package xslice

import (
	"testing"
)

func TestExtract(t *testing.T) {
	type user struct {
		name string
		age  int
	}

	users := []*user{{name: "hello", age: 10}, {name: "1212", age: 2}}
	names := Extract(users, func(t *user) string {
		return t.name
	})

	t.Log(names)

}
