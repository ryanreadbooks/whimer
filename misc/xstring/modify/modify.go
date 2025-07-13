package modify

import (
	"context"
	"reflect"
	"sync"

	"github.com/ryanreadbooks/whimer/misc/xslice"
)

// reflect.go implement a way to modifying string using reflect

type Func func(ctx context.Context, old string) (string, error)

func nopF(ctx context.Context, old string) (string, error) { return old, nil }

type option struct {
	f            Func
	tag          string
	parallel     bool
	abortOnError bool
}

type Option func(o *option)

func WithFunc(f Func) Option {
	return func(o *option) {
		o.f = f
	}
}

func WithTag(t string) Option {
	return func(o *option) {
		o.tag = t
	}
}

func WithParallel(b bool) Option {
	return func(o *option) {
		o.parallel = b
	}
}

func WithAbortOnError(b bool) Option {
	return func(o *option) {
		o.abortOnError = b
	}
}

func getOpt(opts ...Option) *option {
	opt := &option{f: nopF, tag: ""}

	for _, o := range opts {
		o(opt)
	}

	return opt
}

// Modify will modify the string members in a struct.
// The caller should make sure the struct is addressable
func Modify(value any, opts ...Option) error {
	opt := getOpt(opts...)
	targets := search(reflect.ValueOf(value), opt)

	return invokeFunc(context.Background(), targets, opt)
}

// ModifyCtx will modify the string members in a struct.
// The caller should make sure the struct is addressable
func ModifyCtx(ctx context.Context, value any, opts ...Option) error {
	opt := getOpt(opts...)
	targets := search(reflect.ValueOf(value), opt)

	return invokeFunc(ctx, targets, opt)
}

func invokeFunc(ctx context.Context, targets []*label, opt *option) error {
	if len(targets) == 0 {
		return nil
	}

	if opt.parallel {
		var wg sync.WaitGroup
		err := xslice.BatchAsyncExec(&wg, targets, 100, func(start, end int) error {
			for _, t := range targets[start:end] {
				if t != nil && t.rv.CanSet() {
					ret, err := opt.f(ctx, t.origin)
					if err != nil && opt.abortOnError {
						return err
					}

					t.rv.SetString(ret)
				}
			}

			return nil
		})

		return err
	} else {
		for _, t := range targets {
			ret, err := opt.f(ctx, t.origin)
			if err != nil && opt.abortOnError {
				return err
			}
			t.rv.SetString(ret)
		}
	}

	return nil
}

// label represents a scanning results of label fields
type label struct {
	origin string
	rv     reflect.Value
}

func canBeSearched(k reflect.Kind) bool {
	if k != reflect.Pointer &&
		k != reflect.Map &&
		k != reflect.Slice &&
		k != reflect.Struct {
		return false
	}

	return true
}

func searchContainer(v reflect.Value, opt *option) (ts []*label) {
	switch v.Kind() {
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			ts = append(ts, search(iter.Value(), opt)...)
		}
	case reflect.Slice, reflect.Array:
		for sIdx := 0; sIdx < v.Len(); sIdx++ {
			sliceItemValue := v.Index(sIdx)
			ts = append(ts, search(sliceItemValue, opt)...)
		}
	}

	return
}

func searchStruct(v reflect.Value, opt *option) (ts []*label) {
	valueType := v.Type()
	for idx := range v.NumField() {
		fieldValue := v.Field(idx)
		if vFieldKind := v.Field(idx).Kind(); vFieldKind == reflect.String {
			if _, ok := valueType.Field(idx).Tag.Lookup(opt.tag); ok {
				ts = append(ts, &label{
					origin: fieldValue.String(),
					rv:     fieldValue,
				})
			}
		} else {
			if vFieldKind == reflect.Pointer && fieldValue.Elem().Kind() == reflect.String {
				if _, ok := valueType.Field(idx).Tag.Lookup(opt.tag); ok {
					ts = append(ts, &label{
						origin: fieldValue.Elem().String(),
						rv:     fieldValue.Elem(),
					})
				}
			} else {
				ans := search(fieldValue, opt)
				ts = append(ts, ans...)
			}
		}
	}
	return
}

// please pay attention to map and slice item CanSet attribute, make sure they are addressable
func search(v reflect.Value, opt *option) (ts []*label) {
	if !v.IsValid() {
		return
	}

	if v.Kind() == reflect.Interface {
		v = v.Elem() // extract value underneath
	}

	// ensure v is addressable
	valueKind := v.Kind()
	if !canBeSearched(valueKind) {
		return
	}

	// ensure value is not nil
	if (valueKind == reflect.Pointer ||
		valueKind == reflect.Map ||
		valueKind == reflect.Slice) && v.IsNil() {
		return
	}

	if valueKind == reflect.Struct && !v.CanSet() {
		return
	}

	// valueType := v.Type()
	switch valueKind {
	case reflect.Map, reflect.Slice, reflect.Array:
		ts = append(ts, searchContainer(v, opt)...)
	case reflect.Struct:
		ts = append(ts, searchStruct(v, opt)...)
	case reflect.Pointer:
		elem := v.Elem()
		elemType := elem.Type()
		_ = elemType
		elemKind := elem.Kind()
		if elemKind != reflect.Struct {
			return search(elem, opt)
		}

		ts = append(ts, searchStruct(elem, opt)...)
	}

	return
}
