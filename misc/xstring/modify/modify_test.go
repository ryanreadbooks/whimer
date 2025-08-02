package modify

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/xstring/rand"
	. "github.com/smartystreets/goconvey/convey"
)

var _ = json.Marshal

type Bar struct {
	Addr string `test:""`
}

type Foo struct {
	Name    string  `test:""`
	NamePtr *string `test:""`
	B       *Bar
}

func TestReflect(t *testing.T) {
	f := Foo{}
	vf := &Foo{}

	t.Log(reflect.TypeOf(f).Kind())
	t.Log(reflect.TypeOf(vf).Kind())
	t.Log(reflect.TypeOf(vf).Elem().Kind())
	t.Log(reflect.ValueOf(f).CanSet())
	t.Log(reflect.ValueOf(vf).Elem().CanSet())
	t.Log(reflect.ValueOf(vf).CanSet())
}

func TestModify(t *testing.T) {
	f := Foo{Name: "ryan"}

	Modify(f)
}

type Tres struct {
	Bar *Bar
}

var opt = &option{tag: "test"}

func TestModifyPointer(t *testing.T) {
	sPtr := new(string)
	*sPtr = "namePtr"
	vf := &Foo{
		Name:    "ryan",
		NamePtr: sPtr,
		B: &Bar{
			Addr: "earth",
		},
	}

	Convey("TestModifyPointer", t, func() {
		targets := search(reflect.ValueOf(vf), opt)
		t.Log(len(targets))
		So(len(targets), ShouldEqual, 3)
	})
}

func TestModifyNested(t *testing.T) {
	v := &Tres{
		Bar: &Bar{
			Addr: "hello addr",
		},
	}
	targets := search(reflect.ValueOf(v), opt)
	t.Log(len(targets))
}

func TestModifyNestedNoPointer(t *testing.T) {
	type Nested struct {
		Addr string `test:""`
	}
	v := &struct {
		Name string `test:""`
		Bar  Nested
	}{
		Name: "john",
		Bar: Nested{
			Addr: "planet",
		},
	}

	t.Logf("%+v\n", v)

	targets := search(reflect.ValueOf(v), opt)
	t.Log(len(targets))
	for _, target := range targets {
		t.Log(target)
		target.rv.SetString("12345")
	}
	t.Logf("%+v\n", v)
}

func TestModifySlice(t *testing.T) {
	sPtr := new(string)
	*sPtr = "ptr1"

	sPtr2 := new(string)
	*sPtr2 = "ptr2"
	vs := []*Foo{
		{
			Name:    "ryan",
			NamePtr: sPtr,
			B: &Bar{
				Addr: "earth",
			}},
		{
			Name:    "lily",
			NamePtr: sPtr2,
			B: &Bar{
				Addr: "mars",
			},
		},
	}

	Convey("TestModifySlice", t, func() {
		targets := search(reflect.ValueOf(vs), opt)
		So(len(targets), ShouldEqual, 6)
		t.Log(len(targets))
		for _, target := range targets {
			t.Log(target)
		}
	})
}

func TestModifySlice2(t *testing.T) {
	sPtr := new(string)
	*sPtr = "ptr1"

	sPtr2 := new(string)
	*sPtr2 = "ptr2"

	vs2 := []Foo{
		{
			Name:    "ryan",
			NamePtr: sPtr,
			B: &Bar{
				Addr: "earth",
			}},
		{
			Name:    "lily",
			NamePtr: sPtr2,
			B: &Bar{
				Addr: "mars",
			},
		},
	}

	Convey("TestModifySlice2", t, func() {
		targets := search(reflect.ValueOf(vs2), opt)
		So(len(targets), ShouldEqual, 6)
		t.Log(len(targets))
		for _, target := range targets {
			t.Log(target)
		}
	})
}

func TestNoEffect(t *testing.T) {
	s := "hello world"
	targets := search(reflect.ValueOf(s), opt)
	t.Log(len(targets))

	s1 := []string{"hello", "world"}

	targets = search(reflect.ValueOf(s1), opt)
	t.Log(len(targets))

	s2 := map[string]string{"1": "hello", "2": "world"}
	targets = search(reflect.ValueOf(s2), opt)
	t.Log(len(targets))
}

// 1
type Code struct {
	Version string `test:""`
}

func (c Code) IsAllDesc(s string) bool {
	return c.Version == s
}

// 2 + 2 + n + m
type Meta struct {
	Width       int
	Height      string  `test:""`
	Format      *string `test:""`
	Code        Code
	Coding      *Code
	CodeSlice   []Code
	CodingSlice []*Code
}

func (m *Meta) IsAllDest(s string) bool {
	r1 := m.Height == s && *m.Format == s
	r1 = r1 && m.Code.IsAllDesc(s) && m.Coding.IsAllDesc(s)
	for _, c := range m.CodeSlice {
		r1 = r1 && c.IsAllDesc(s)
	}
	for _, c := range m.CodingSlice {
		r1 = r1 && c.IsAllDesc(s)
	}

	return r1
}

type Person struct {
	Name         string  `test:""`
	Address      *string `test:""`
	Meta         Meta
	MetaPtr      *Meta
	MetaMap      map[int]*Meta
	MetaPtrMapV2 map[string]*Meta
	CodeSliceMap map[string][]*Code
	MetaMapSlice map[int][]*Meta
	MetaSlice    []*Meta
	MetaSliceMap []map[string]*Meta
}

var getPtrString = func(s string) *string {
	o := new(string)
	*o = s

	return o
}

var getMeta = func() Meta {
	return Meta{
		Width:  100,
		Height: "height_" + rand.Random(6),
		Format: getPtrString("format_" + rand.Random(6)),
		Code:   Code{Version: "vmeta_" + rand.Random(5)},
		Coding: &Code{Version: "vmetaing_" + rand.Random(2)},
		CodeSlice: []Code{
			{"vmetas_" + rand.Random(3)},
			{"vmetas_" + rand.Random(3)}},
		CodingSlice: []*Code{
			{"vmetacd_" + rand.Random(4)},
			{"vmetacd_" + rand.Random(4)}},
	}
}

var getMetaPtr = func() *Meta {
	return &Meta{
		Width:  200,
		Height: "heightptr_" + rand.Random(6),
		Format: getPtrString("format_" + rand.Random(6)),
		Code:   Code{Version: "v_" + rand.Random(5)},
		Coding: &Code{Version: "ving_" + rand.Random(2)},
		CodeSlice: []Code{{"vslice_" + rand.Random(3)},
			{"vslice_" + rand.Random(3)}},
		CodingSlice: []*Code{{"vsp_" + rand.Random(4)},
			{"vsp_" + rand.Random(4)}},
	}
}

func TestModifyNestedPointer(t *testing.T) {
	type My struct {
		Meta Meta
	}
	m := My{
		Meta: getMeta(),
	}

	targets := search(reflect.ValueOf(&m), opt)
	t.Log(len(targets))
	dest := "world"
	for _, tar := range targets {
		t.Log(tar.origin)
		tar.rv.SetString(dest)
	}
	Convey("TestModifyNestedPointer", t, func() {
		So(m.Meta.IsAllDest(dest), ShouldBeTrue)
	})
}

func TestMofidyMap(t *testing.T) {
	type Maps struct {
		Members map[string]Code
	}

	p := Maps{
		Members: map[string]Code{"1": {"v1"}, "2": {"v2"}},
	}

	targets := search(reflect.ValueOf(&p), opt)
	t.Log(len(targets))
	for _, target := range targets {
		t.Log(target)
	}
}

func TestSearch(t *testing.T) {

	p := Person{
		Name:         "name_" + rand.Random(4),
		Address:      getPtrString("address_" + rand.Random(3)),
		Meta:         getMeta(),
		MetaPtr:      getMetaPtr(),
		MetaMap:      map[int]*Meta{1: getMetaPtr(), 2: getMetaPtr(), 3: getMetaPtr()}, // CanSet = false
		MetaPtrMapV2: map[string]*Meta{"a": getMetaPtr(), "b": getMetaPtr()},
		CodeSliceMap: map[string][]*Code{
			"q": {{"q1"}, {"q2"}},
			"w": {{"w1"}, {"w2"}},
		},
		MetaMapSlice: map[int][]*Meta{
			100: {getMetaPtr(), getMetaPtr()},
			200: {getMetaPtr(), getMetaPtr(), getMetaPtr()},
		},
		MetaSlice: []*Meta{
			getMetaPtr(),
			getMetaPtr(),
			getMetaPtr(),
		},
		MetaSliceMap: []map[string]*Meta{
			{"abc": getMetaPtr(), "bcd": getMetaPtr()},
			{"er": getMetaPtr(), "pp": getMetaPtr(), "oi": getMetaPtr()},
		},
	}

	Convey("TestSearch", t, func() {
		dest := "helloworld"
		targets := search(reflect.ValueOf(&p), opt)
		for _, target := range targets {
			target.rv.SetString(dest)
		}

		// b, _ := json.MarshalIndent(p, "", "  ")
		// t.Log(string(b))

		// check every string in person variable
		So(p.Name, ShouldEqual, dest)
		So(*p.Address, ShouldEqual, dest)
		So(p.Meta.IsAllDest(dest), ShouldBeTrue)
		So(p.MetaPtr.IsAllDest(dest), ShouldBeTrue)
		for _, m := range p.MetaMap {
			So(m.IsAllDest(dest), ShouldBeTrue)
		}
		for _, m := range p.MetaPtrMapV2 {
			So(m.IsAllDest(dest), ShouldBeTrue)
		}
		for _, m := range p.CodeSliceMap {
			for _, c := range m {
				So(c.IsAllDesc(dest), ShouldBeTrue)
			}
		}
		for _, m := range p.MetaMapSlice {
			for _, c := range m {
				So(c.IsAllDest(dest), ShouldBeTrue)
			}
		}
		for _, m := range p.MetaSlice {
			So(m.IsAllDest(dest), ShouldBeTrue)
		}
		for _, m := range p.MetaSliceMap {
			for _, c := range m {
				So(c.IsAllDest(dest), ShouldBeTrue)
			}
		}

	})
}

func TestModifyCtx(t *testing.T) {
	type Type struct {
		Meta *Meta
	}

	p := Type{
		Meta: getMetaPtr(),
	}

	Convey("TestModifyCtx", t, func() {
		b, _ := json.MarshalIndent(p, "", " ")
		t.Log(string(b))
		err := ModifyCtx(context.Background(), &p, WithTag("test"), WithFunc(
			func(ctx context.Context, old string) (string, error) {
				return strings.ToUpper(old), nil
			},
		))
		t.Log("-----")
		So(err, ShouldBeNil)
		b, _ = json.MarshalIndent(p, "", " ")
		t.Log(string(b))
	})
}
