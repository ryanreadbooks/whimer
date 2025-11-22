package extension

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

type serviceDescriptor struct {
	service string
	desc    protoreflect.ServiceDescriptor
	option  *options.WhimerServiceOption
	exist   bool
}

type serviceDescriptorStore map[string]*serviceDescriptor

type serviceHolder struct {
	mu    sync.Mutex
	store atomic.Pointer[serviceDescriptorStore]
}

func newServiceHolder() *serviceHolder {
	m := make(serviceDescriptorStore)
	so := &serviceHolder{}
	so.store.Store(&m)

	return so
}

func (so *serviceHolder) appendInvalidService(service string) {
	so.mu.Lock()
	defer so.mu.Unlock()

	cache := *so.store.Load()
	newCache := make(serviceDescriptorStore, len(cache)+1)
	maps.Copy(newCache, cache)
	newCache[service] = &serviceDescriptor{exist: false, service: service}
	so.store.Store(&newCache)
}

func (so *serviceHolder) appendValidService(service string,
	option *options.WhimerServiceOption,
	desc protoreflect.ServiceDescriptor) *serviceDescriptor {
	so.mu.Lock()
	defer so.mu.Unlock()

	cache := *so.store.Load()
	newCache := make(serviceDescriptorStore, len(cache)+1)
	maps.Copy(newCache, cache)
	sd := &serviceDescriptor{
		exist:   true,
		desc:    desc,
		option:  option,
		service: service}
	newCache[service] = sd
	so.store.Store(&newCache)

	return sd
}

func (so *serviceHolder) loadOrBuild(methodName string) *options.WhimerServiceOption {
	serviceName := buildCanonicalServiceName(methodName)
	if serviceName == "" {
		return nil
	}

	if existingDesc, ok := (*so.store.Load())[serviceName]; ok {
		if existingDesc != nil && existingDesc.exist {
			return existingDesc.option
		} else {
			return nil
		}
	}

	exist := false
	defer func() {
		if !exist {
			so.appendInvalidService(serviceName)
		}
	}()

	// find in global registry
	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(serviceName))
	if err != nil {
		// not found
		return nil
	}

	// found descriptor
	serviceDesc, ok := desc.(protoreflect.ServiceDescriptor)
	if !ok {
		return nil
	}

	serviceExt := proto.GetExtension(serviceDesc.Options(), options.E_Service)
	if serviceExt == nil {
		return nil
	}

	targetOption, ok := serviceExt.(*options.WhimerServiceOption)
	if !ok || targetOption == nil {
		return nil
	}

	// option is valid here
	so.appendValidService(serviceName, targetOption, serviceDesc)
	exist = true
	return targetOption
}

func buildCanonicalServiceName(methodName string) string {
	mn := buildCanonicalMethodName(methodName)
	if i := strings.LastIndexByte(string(mn), '.'); i >= 0 {
		return mn[:i]
	}

	return ""
}
