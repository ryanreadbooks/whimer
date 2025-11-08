package extension

// See: github.com/ryanreadbooks/whimer/apiextension/protobuf/options

import (
	"maps"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/ryanreadbooks/whimer/apiextension/protobuf/options"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type methodDescriptor struct {
	method string
	desc   protoreflect.MethodDescriptor
	option *options.WhimerMethodOption
	exist  bool
}

func nonExistingMethodDescriptor(method string) *methodDescriptor {
	return &methodDescriptor{
		method: method,
		exist:  false,
		desc:   nil,
	}
}

type methodDescriptorStore map[string]*methodDescriptor

type methodHolder struct {
	mu    sync.Mutex
	store atomic.Pointer[methodDescriptorStore]
}

func newMethodHolder() *methodHolder {
	m := make(methodDescriptorStore)
	mo := &methodHolder{}
	mo.store.Store(&m)

	return mo
}

func (mo *methodHolder) appendInvalidMethod(method string) {
	mo.mu.Lock()
	defer mo.mu.Unlock()

	// populate a non-existing descriptor
	cache := *mo.store.Load()
	newCache := make(methodDescriptorStore, len(cache)+1)
	maps.Copy(newCache, cache)
	newCache[method] = nonExistingMethodDescriptor(method)
	mo.store.Store(&newCache)
}

func (mo *methodHolder) appendValidMethod(method string,
	option *options.WhimerMethodOption, desc protoreflect.MethodDescriptor) *methodDescriptor {

	mo.mu.Lock()
	defer mo.mu.Unlock()

	md := &methodDescriptor{
		method: method,
		option: option,
		exist:  true,
		desc:   desc,
	}

	// populate a non-existing descriptor
	cache := *mo.store.Load()
	newCache := make(methodDescriptorStore, len(cache)+1)
	maps.Copy(newCache, cache)
	newCache[method] = md
	mo.store.Store(&newCache)

	return md
}

func (mo *methodHolder) loadOrBuild(method string) *options.WhimerMethodOption {
	if existingDesc, ok := (*mo.store.Load())[method]; ok {
		// found descriptor
		if existingDesc != nil && existingDesc.exist {
			return existingDesc.option
		} else {
			return nil
		}
	}

	exist := false
	defer func() {
		if !exist {
			mo.appendInvalidMethod(method)
		}
	}()

	canonicalName := buildCanonicalMethodName(method)
	if canonicalName == "" {
		return nil
	}

	// find in global registry
	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(canonicalName))
	if err != nil {
		// not found
		return nil
	}

	// found descriptor, but we still need to check if options exists
	methodDesc, ok := desc.(protoreflect.MethodDescriptor)
	if !ok {
		return nil
	}

	methodExt := proto.GetExtension(methodDesc.Options(), options.E_Method)
	if methodExt == nil {
		return nil
	}

	// method ext should be whimer method option
	targetOption, ok := methodExt.(*options.WhimerMethodOption)
	if !ok || targetOption == nil {
		return nil
	}

	// option is valid here
	mo.appendValidMethod(method, targetOption, methodDesc)
	exist = true
	return targetOption
}

// method has the format of this: /yourpackage.servicename/methodName
func buildCanonicalMethodName(method string) string {
	method = method[1:]
	splits := strings.SplitN(method, "/", 2)
	if len(splits) != 2 {
		return ""
	}

	name := strings.Join(splits, ".")
	return name
}
