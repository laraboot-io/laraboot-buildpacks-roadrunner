package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit/postal"
)

type DependencyService struct {
	DeliverCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Dependency   postal.Dependency
			CnbPath      string
			LayerPath    string
			PlatformPath string
		}
		Returns struct {
			Error error
		}
		Stub func(postal.Dependency, string, string, string) error
	}
	InstallCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Dependency postal.Dependency
			CnbPath    string
			LayerPath  string
		}
		Returns struct {
			Error error
		}
		Stub func(postal.Dependency, string, string) error
	}
	ResolveCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Path    string
			Name    string
			Version string
			Stack   string
		}
		Returns struct {
			Dependency postal.Dependency
			Error      error
		}
		Stub func(string, string, string, string) (postal.Dependency, error)
	}
}

func (f *DependencyService) Deliver(dependency postal.Dependency, cnbPath, layerPath string, platformPath string) error {

	f.DeliverCall.Lock()
	defer f.DeliverCall.Unlock()
	f.DeliverCall.CallCount++
	f.DeliverCall.Receives.Dependency = dependency
	f.DeliverCall.Receives.CnbPath = cnbPath
	f.DeliverCall.Receives.LayerPath = layerPath
	f.DeliverCall.Receives.PlatformPath = platformPath
	if f.DeliverCall.Stub != nil {
		return f.DeliverCall.Stub(dependency, cnbPath, layerPath, platformPath)
	}
	return f.DeliverCall.Returns.Error

}

func (f *DependencyService) Install(param1 postal.Dependency, param2 string, param3 string) error {
	f.InstallCall.Lock()
	defer f.InstallCall.Unlock()
	f.InstallCall.CallCount++
	f.InstallCall.Receives.Dependency = param1
	f.InstallCall.Receives.CnbPath = param2
	f.InstallCall.Receives.LayerPath = param3
	if f.InstallCall.Stub != nil {
		return f.InstallCall.Stub(param1, param2, param3)
	}
	return f.InstallCall.Returns.Error
}
func (f *DependencyService) Resolve(param1 string, param2 string, param3 string, param4 string) (postal.Dependency, error) {
	f.ResolveCall.Lock()
	defer f.ResolveCall.Unlock()
	f.ResolveCall.CallCount++
	f.ResolveCall.Receives.Path = param1
	f.ResolveCall.Receives.Name = param2
	f.ResolveCall.Receives.Version = param3
	f.ResolveCall.Receives.Stack = param4
	if f.ResolveCall.Stub != nil {
		return f.ResolveCall.Stub(param1, param2, param3, param4)
	}
	return f.ResolveCall.Returns.Dependency, f.ResolveCall.Returns.Error
}
