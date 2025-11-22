package extension

import (
	"maps"
	"sync"
	"sync/atomic"

	"github.com/ryanreadbooks/whimer/apiextension/protobuf/options"
)

type IExtensionOption interface {
	GetSkipMetadataUidCheck() bool
}

var (
	_ IExtensionOption = &options.WhimerMethodOption{}
	_ IExtensionOption = &options.WhimerServiceOption{}
)

type ExtensionOption struct {
	SkipMetadataUidCheck bool

	valid bool
}

type Extension struct {
	methods  *methodHolder
	services *serviceHolder

	mu    sync.Mutex
	cache atomic.Pointer[map[string]*ExtensionOption]
}

func newExtensionOption(o IExtensionOption) *ExtensionOption {
	valid := o != nil
	eo := &ExtensionOption{
		valid: valid,
	}

	if o != nil {
		eo.SkipMetadataUidCheck = o.GetSkipMetadataUidCheck()
	}
	return eo
}

// rpc服务的扩展
func NewExtension() *Extension {
	ext := &Extension{
		methods:  newMethodHolder(),
		services: newServiceHolder(),
	}
	m := make(map[string]*ExtensionOption)
	ext.cache.Store(&m)

	return ext
}

func (e *Extension) addInvalidOption(method string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	oldCache := *e.cache.Load()
	newCache := make(map[string]*ExtensionOption, len(oldCache)+1)
	maps.Copy(newCache, oldCache)
	newCache[method] = newExtensionOption(nil)

	e.cache.Store(&newCache)
}

func (e *Extension) addValidOption(method string, o *ExtensionOption) {
	e.mu.Lock()
	defer e.mu.Unlock()

	oldCache := *e.cache.Load()
	newCache := make(map[string]*ExtensionOption, len(oldCache)+1)
	maps.Copy(newCache, oldCache)
	newCache[method] = o

	e.cache.Store(&newCache)
}

// Get method extension
func (e *Extension) Get(method string) *ExtensionOption {
	return e.getOrBuild(method)
}

func (e *Extension) getOrBuild(method string) *ExtensionOption {
	option, ok := (*e.cache.Load())[method]
	if ok {
		return option
	}

	// not ok then we try to fetch from source
	methodOption := e.methods.loadOrBuild(method)
	if methodOption != nil {
		no := newExtensionOption(methodOption)
		e.addValidOption(method, no)
		return no
	}

	// we can not find it in methods, try in services
	serviceOption := e.services.loadOrBuild(method)
	if serviceOption != nil {
		no := newExtensionOption(serviceOption)
		e.addValidOption(method, no)
		return no
	}

	// not found at all, return invalid
	e.addInvalidOption(method)

	option, _ = (*e.cache.Load())[method]
	return option
}
