package xsql

import (
	"reflect"
	"slices"
	"strings"
)

func getFields(v any, keep func(v string) bool) (string, []string) {
	var tags []string
	val := reflect.TypeOf(v)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 只处理结构体类型
	if val.Kind() != reflect.Struct {
		return "", nil
	}

	for i := range val.NumField() {
		field := val.Field(i)
		if field.PkgPath != "" {
			continue
		}

		// 获取 db tag
		tag := field.Tag.Get("db")
		if tag == "" {
			continue
		}

		// 可以跳过某些字段
		if keep(tag) {
			tags = append(tags, "`"+tag+"`")
		}
	}

	return strings.Join(tags, ","), tags
}

func GetFields(v any, skip ...string) string {
	s, _ := getFields(v, func(tag string) bool {
		return !slices.Contains(skip, tag)
	})

	return s
}

func getQuest(fields string) string {
	fieldsList := strings.Split(fields, ",")
	fl := len(fieldsList)
	if fl == 0 {
		return ""
	}

	var bd strings.Builder
	bd.Grow(fl * 2)
	for i := range fl {
		bd.WriteByte('?')
		if i != fl-1 {
			bd.WriteByte(',')
		}
	}

	return bd.String()
}

// 同时返回参数化参数?
func GetFields2WithSkip(v any, skip ...string) (string, string) {
	fields := strings.TrimSpace(GetFields(v, skip...))
	if len(fields) == 0 {
		return fields, ""
	}

	return fields, getQuest(fields)
}

func SelectFields(v any, targets ...string) string {
	// _, list:= getFields(v, func(tag string) bool {
	// 	return slices.Contains(targets, tag)
	// })

	// 按照targets顺序返回
	return strings.Join(targets, ",")
}

func SelectFields2(v any, targets ...string) (string, string) {
	fields := strings.TrimSpace(SelectFields(v, targets...))
	if len(fields) == 0 {
		return fields, ""
	}

	return fields, getQuest(fields)
}
