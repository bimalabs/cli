package engine

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/sarulabs/di/v2"
	"github.com/sarulabs/dingo/v4"

	providerPkg "github.com/bimalabs/cli/provider"

	generator "github.com/bimalabs/cli/generator"
)

// C retrieves a Container from an interface.
// The function panics if the Container can not be retrieved.
//
// The interface can be :
// - a *Container
// - an *http.Request containing a *Container in its context.Context
//   for the dingo.ContainerKey("dingo") key.
//
// The function can be changed to match the needs of your application.
var C = func(i interface{}) *Container {
	if c, ok := i.(*Container); ok {
		return c
	}
	r, ok := i.(*http.Request)
	if !ok {
		panic("could not get the container with dic.C()")
	}
	c, ok := r.Context().Value(dingo.ContainerKey("dingo")).(*Container)
	if !ok {
		panic("could not get the container from the given *http.Request in dic.C()")
	}
	return c
}

type builder struct {
	builder *di.Builder
}

// NewBuilder creates a builder that can be used to create a Container.
// You probably should use NewContainer to create the container directly.
// But using NewBuilder allows you to redefine some di services.
// This can be used for testing.
// But this behavior is not safe, so be sure to know what you are doing.
func NewBuilder(scopes ...string) (*builder, error) {
	if len(scopes) == 0 {
		scopes = []string{di.App, di.Request, di.SubRequest}
	}
	b, err := di.NewBuilder(scopes...)
	if err != nil {
		return nil, fmt.Errorf("could not create di.Builder: %v", err)
	}
	provider := &providerPkg.Generator{}
	if err := provider.Load(); err != nil {
		return nil, fmt.Errorf("could not load definitions with the Provider (Generator from github.com/bimalabs/cli/provider): %v", err)
	}
	for _, d := range getDiDefs(provider) {
		if err := b.Add(d); err != nil {
			return nil, fmt.Errorf("could not add di.Def in di.Builder: %v", err)
		}
	}
	return &builder{builder: b}, nil
}

// Add adds one or more definitions in the Builder.
// It returns an error if a definition can not be added.
func (b *builder) Add(defs ...di.Def) error {
	return b.builder.Add(defs...)
}

// Set is a shortcut to add a definition for an already built object.
func (b *builder) Set(name string, obj interface{}) error {
	return b.builder.Set(name, obj)
}

// Build creates a Container in the most generic scope.
func (b *builder) Build() *Container {
	return &Container{ctn: b.builder.Build()}
}

// NewContainer creates a new Container.
// If no scope is provided, di.App, di.Request and di.SubRequest are used.
// The returned Container has the most generic scope (di.App).
// The SubContainer() method should be called to get a Container in a more specific scope.
func NewContainer(scopes ...string) (*Container, error) {
	b, err := NewBuilder(scopes...)
	if err != nil {
		return nil, err
	}
	return b.Build(), nil
}

// Container represents a generated dependency injection container.
// It is a wrapper around a di.Container.
//
// A Container has a scope and may have a parent in a more generic scope
// and children in a more specific scope.
// Objects can be retrieved from the Container.
// If the requested object does not already exist in the Container,
// it is built thanks to the object definition.
// The following attempts to get this object will return the same object.
type Container struct {
	ctn di.Container
}

// Scope returns the Container scope.
func (c *Container) Scope() string {
	return c.ctn.Scope()
}

// Scopes returns the list of available scopes.
func (c *Container) Scopes() []string {
	return c.ctn.Scopes()
}

// ParentScopes returns the list of scopes wider than the Container scope.
func (c *Container) ParentScopes() []string {
	return c.ctn.ParentScopes()
}

// SubScopes returns the list of scopes that are more specific than the Container scope.
func (c *Container) SubScopes() []string {
	return c.ctn.SubScopes()
}

// Parent returns the parent Container.
func (c *Container) Parent() *Container {
	if p := c.ctn.Parent(); p != nil {
		return &Container{ctn: p}
	}
	return nil
}

// SubContainer creates a new Container in the next sub-scope
// that will have this Container as parent.
func (c *Container) SubContainer() (*Container, error) {
	sub, err := c.ctn.SubContainer()
	if err != nil {
		return nil, err
	}
	return &Container{ctn: sub}, nil
}

// SafeGet retrieves an object from the Container.
// The object has to belong to this scope or a more generic one.
// If the object does not already exist, it is created and saved in the Container.
// If the object can not be created, it returns an error.
func (c *Container) SafeGet(name string) (interface{}, error) {
	return c.ctn.SafeGet(name)
}

// Get is similar to SafeGet but it does not return the error.
// Instead it panics.
func (c *Container) Get(name string) interface{} {
	return c.ctn.Get(name)
}

// Fill is similar to SafeGet but it does not return the object.
// Instead it fills the provided object with the value returned by SafeGet.
// The provided object must be a pointer to the value returned by SafeGet.
func (c *Container) Fill(name string, dst interface{}) error {
	return c.ctn.Fill(name, dst)
}

// UnscopedSafeGet retrieves an object from the Container, like SafeGet.
// The difference is that the object can be retrieved
// even if it belongs to a more specific scope.
// To do so, UnscopedSafeGet creates a sub-container.
// When the created object is no longer needed,
// it is important to use the Clean method to delete this sub-container.
func (c *Container) UnscopedSafeGet(name string) (interface{}, error) {
	return c.ctn.UnscopedSafeGet(name)
}

// UnscopedGet is similar to UnscopedSafeGet but it does not return the error.
// Instead it panics.
func (c *Container) UnscopedGet(name string) interface{} {
	return c.ctn.UnscopedGet(name)
}

// UnscopedFill is similar to UnscopedSafeGet but copies the object in dst instead of returning it.
func (c *Container) UnscopedFill(name string, dst interface{}) error {
	return c.ctn.UnscopedFill(name, dst)
}

// Clean deletes the sub-container created by UnscopedSafeGet, UnscopedGet or UnscopedFill.
func (c *Container) Clean() error {
	return c.ctn.Clean()
}

// DeleteWithSubContainers takes all the objects saved in this Container
// and calls the Close function of their Definition on them.
// It will also call DeleteWithSubContainers on each child and remove its reference in the parent Container.
// After deletion, the Container can no longer be used.
// The sub-containers are deleted even if they are still used in other goroutines.
// It can cause errors. You may want to use the Delete method instead.
func (c *Container) DeleteWithSubContainers() error {
	return c.ctn.DeleteWithSubContainers()
}

// Delete works like DeleteWithSubContainers if the Container does not have any child.
// But if the Container has sub-containers, it will not be deleted right away.
// The deletion only occurs when all the sub-containers have been deleted manually.
// So you have to call Delete or DeleteWithSubContainers on all the sub-containers.
func (c *Container) Delete() error {
	return c.ctn.Delete()
}

// IsClosed returns true if the Container has been deleted.
func (c *Container) IsClosed() bool {
	return c.ctn.IsClosed()
}

// SafeGetBimaGeneratorDic retrieves the "bima:generator:dic" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:dic"
// 	type: *generator.Dic
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetBimaGeneratorDic() (*generator.Dic, error) {
	i, err := c.ctn.SafeGet("bima:generator:dic")
	if err != nil {
		var eo *generator.Dic
		return eo, err
	}
	o, ok := i.(*generator.Dic)
	if !ok {
		return o, errors.New("could get 'bima:generator:dic' because the object could not be cast to *generator.Dic")
	}
	return o, nil
}

// GetBimaGeneratorDic retrieves the "bima:generator:dic" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:dic"
// 	type: *generator.Dic
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetBimaGeneratorDic() *generator.Dic {
	o, err := c.SafeGetBimaGeneratorDic()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetBimaGeneratorDic retrieves the "bima:generator:dic" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:dic"
// 	type: *generator.Dic
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// This method can be called even if generator is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetBimaGeneratorDic() (*generator.Dic, error) {
	i, err := c.ctn.UnscopedSafeGet("bima:generator:dic")
	if err != nil {
		var eo *generator.Dic
		return eo, err
	}
	o, ok := i.(*generator.Dic)
	if !ok {
		return o, errors.New("could get 'bima:generator:dic' because the object could not be cast to *generator.Dic")
	}
	return o, nil
}

// UnscopedGetBimaGeneratorDic retrieves the "bima:generator:dic" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:dic"
// 	type: *generator.Dic
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// This method can be called even if generator is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetBimaGeneratorDic() *generator.Dic {
	o, err := c.UnscopedSafeGetBimaGeneratorDic()
	if err != nil {
		panic(err)
	}
	return o
}

// BimaGeneratorDic retrieves the "bima:generator:dic" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:dic"
// 	type: *generator.Dic
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetBimaGeneratorDic method.
// If the container can not be retrieved, it panics.
func BimaGeneratorDic(i interface{}) *generator.Dic {
	return C(i).GetBimaGeneratorDic()
}

// SafeGetBimaGeneratorModel retrieves the "bima:generator:model" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:model"
// 	type: *generator.Model
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetBimaGeneratorModel() (*generator.Model, error) {
	i, err := c.ctn.SafeGet("bima:generator:model")
	if err != nil {
		var eo *generator.Model
		return eo, err
	}
	o, ok := i.(*generator.Model)
	if !ok {
		return o, errors.New("could get 'bima:generator:model' because the object could not be cast to *generator.Model")
	}
	return o, nil
}

// GetBimaGeneratorModel retrieves the "bima:generator:model" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:model"
// 	type: *generator.Model
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetBimaGeneratorModel() *generator.Model {
	o, err := c.SafeGetBimaGeneratorModel()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetBimaGeneratorModel retrieves the "bima:generator:model" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:model"
// 	type: *generator.Model
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// This method can be called even if generator is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetBimaGeneratorModel() (*generator.Model, error) {
	i, err := c.ctn.UnscopedSafeGet("bima:generator:model")
	if err != nil {
		var eo *generator.Model
		return eo, err
	}
	o, ok := i.(*generator.Model)
	if !ok {
		return o, errors.New("could get 'bima:generator:model' because the object could not be cast to *generator.Model")
	}
	return o, nil
}

// UnscopedGetBimaGeneratorModel retrieves the "bima:generator:model" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:model"
// 	type: *generator.Model
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// This method can be called even if generator is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetBimaGeneratorModel() *generator.Model {
	o, err := c.UnscopedSafeGetBimaGeneratorModel()
	if err != nil {
		panic(err)
	}
	return o
}

// BimaGeneratorModel retrieves the "bima:generator:model" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:model"
// 	type: *generator.Model
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetBimaGeneratorModel method.
// If the container can not be retrieved, it panics.
func BimaGeneratorModel(i interface{}) *generator.Model {
	return C(i).GetBimaGeneratorModel()
}

// SafeGetBimaGeneratorModule retrieves the "bima:generator:module" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:module"
// 	type: *generator.Module
// 	scope: "generator"
// 	build: struct
// 	params:
// 		- "Config": Value([]string)
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetBimaGeneratorModule() (*generator.Module, error) {
	i, err := c.ctn.SafeGet("bima:generator:module")
	if err != nil {
		var eo *generator.Module
		return eo, err
	}
	o, ok := i.(*generator.Module)
	if !ok {
		return o, errors.New("could get 'bima:generator:module' because the object could not be cast to *generator.Module")
	}
	return o, nil
}

// GetBimaGeneratorModule retrieves the "bima:generator:module" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:module"
// 	type: *generator.Module
// 	scope: "generator"
// 	build: struct
// 	params:
// 		- "Config": Value([]string)
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetBimaGeneratorModule() *generator.Module {
	o, err := c.SafeGetBimaGeneratorModule()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetBimaGeneratorModule retrieves the "bima:generator:module" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:module"
// 	type: *generator.Module
// 	scope: "generator"
// 	build: struct
// 	params:
// 		- "Config": Value([]string)
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// This method can be called even if generator is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetBimaGeneratorModule() (*generator.Module, error) {
	i, err := c.ctn.UnscopedSafeGet("bima:generator:module")
	if err != nil {
		var eo *generator.Module
		return eo, err
	}
	o, ok := i.(*generator.Module)
	if !ok {
		return o, errors.New("could get 'bima:generator:module' because the object could not be cast to *generator.Module")
	}
	return o, nil
}

// UnscopedGetBimaGeneratorModule retrieves the "bima:generator:module" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:module"
// 	type: *generator.Module
// 	scope: "generator"
// 	build: struct
// 	params:
// 		- "Config": Value([]string)
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// This method can be called even if generator is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetBimaGeneratorModule() *generator.Module {
	o, err := c.UnscopedSafeGetBimaGeneratorModule()
	if err != nil {
		panic(err)
	}
	return o
}

// BimaGeneratorModule retrieves the "bima:generator:module" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:module"
// 	type: *generator.Module
// 	scope: "generator"
// 	build: struct
// 	params:
// 		- "Config": Value([]string)
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetBimaGeneratorModule method.
// If the container can not be retrieved, it panics.
func BimaGeneratorModule(i interface{}) *generator.Module {
	return C(i).GetBimaGeneratorModule()
}

// SafeGetBimaGeneratorProto retrieves the "bima:generator:proto" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:proto"
// 	type: *generator.Proto
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetBimaGeneratorProto() (*generator.Proto, error) {
	i, err := c.ctn.SafeGet("bima:generator:proto")
	if err != nil {
		var eo *generator.Proto
		return eo, err
	}
	o, ok := i.(*generator.Proto)
	if !ok {
		return o, errors.New("could get 'bima:generator:proto' because the object could not be cast to *generator.Proto")
	}
	return o, nil
}

// GetBimaGeneratorProto retrieves the "bima:generator:proto" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:proto"
// 	type: *generator.Proto
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetBimaGeneratorProto() *generator.Proto {
	o, err := c.SafeGetBimaGeneratorProto()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetBimaGeneratorProto retrieves the "bima:generator:proto" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:proto"
// 	type: *generator.Proto
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// This method can be called even if generator is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetBimaGeneratorProto() (*generator.Proto, error) {
	i, err := c.ctn.UnscopedSafeGet("bima:generator:proto")
	if err != nil {
		var eo *generator.Proto
		return eo, err
	}
	o, ok := i.(*generator.Proto)
	if !ok {
		return o, errors.New("could get 'bima:generator:proto' because the object could not be cast to *generator.Proto")
	}
	return o, nil
}

// UnscopedGetBimaGeneratorProto retrieves the "bima:generator:proto" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:proto"
// 	type: *generator.Proto
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// This method can be called even if generator is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetBimaGeneratorProto() *generator.Proto {
	o, err := c.UnscopedSafeGetBimaGeneratorProto()
	if err != nil {
		panic(err)
	}
	return o
}

// BimaGeneratorProto retrieves the "bima:generator:proto" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:proto"
// 	type: *generator.Proto
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetBimaGeneratorProto method.
// If the container can not be retrieved, it panics.
func BimaGeneratorProto(i interface{}) *generator.Proto {
	return C(i).GetBimaGeneratorProto()
}

// SafeGetBimaGeneratorProvider retrieves the "bima:generator:provider" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:provider"
// 	type: *generator.Provider
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetBimaGeneratorProvider() (*generator.Provider, error) {
	i, err := c.ctn.SafeGet("bima:generator:provider")
	if err != nil {
		var eo *generator.Provider
		return eo, err
	}
	o, ok := i.(*generator.Provider)
	if !ok {
		return o, errors.New("could get 'bima:generator:provider' because the object could not be cast to *generator.Provider")
	}
	return o, nil
}

// GetBimaGeneratorProvider retrieves the "bima:generator:provider" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:provider"
// 	type: *generator.Provider
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetBimaGeneratorProvider() *generator.Provider {
	o, err := c.SafeGetBimaGeneratorProvider()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetBimaGeneratorProvider retrieves the "bima:generator:provider" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:provider"
// 	type: *generator.Provider
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// This method can be called even if generator is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetBimaGeneratorProvider() (*generator.Provider, error) {
	i, err := c.ctn.UnscopedSafeGet("bima:generator:provider")
	if err != nil {
		var eo *generator.Provider
		return eo, err
	}
	o, ok := i.(*generator.Provider)
	if !ok {
		return o, errors.New("could get 'bima:generator:provider' because the object could not be cast to *generator.Provider")
	}
	return o, nil
}

// UnscopedGetBimaGeneratorProvider retrieves the "bima:generator:provider" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:provider"
// 	type: *generator.Provider
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// This method can be called even if generator is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetBimaGeneratorProvider() *generator.Provider {
	o, err := c.UnscopedSafeGetBimaGeneratorProvider()
	if err != nil {
		panic(err)
	}
	return o
}

// BimaGeneratorProvider retrieves the "bima:generator:provider" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:provider"
// 	type: *generator.Provider
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetBimaGeneratorProvider method.
// If the container can not be retrieved, it panics.
func BimaGeneratorProvider(i interface{}) *generator.Provider {
	return C(i).GetBimaGeneratorProvider()
}

// SafeGetBimaGeneratorServer retrieves the "bima:generator:server" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:server"
// 	type: *generator.Server
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetBimaGeneratorServer() (*generator.Server, error) {
	i, err := c.ctn.SafeGet("bima:generator:server")
	if err != nil {
		var eo *generator.Server
		return eo, err
	}
	o, ok := i.(*generator.Server)
	if !ok {
		return o, errors.New("could get 'bima:generator:server' because the object could not be cast to *generator.Server")
	}
	return o, nil
}

// GetBimaGeneratorServer retrieves the "bima:generator:server" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:server"
// 	type: *generator.Server
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetBimaGeneratorServer() *generator.Server {
	o, err := c.SafeGetBimaGeneratorServer()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetBimaGeneratorServer retrieves the "bima:generator:server" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:server"
// 	type: *generator.Server
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// This method can be called even if generator is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetBimaGeneratorServer() (*generator.Server, error) {
	i, err := c.ctn.UnscopedSafeGet("bima:generator:server")
	if err != nil {
		var eo *generator.Server
		return eo, err
	}
	o, ok := i.(*generator.Server)
	if !ok {
		return o, errors.New("could get 'bima:generator:server' because the object could not be cast to *generator.Server")
	}
	return o, nil
}

// UnscopedGetBimaGeneratorServer retrieves the "bima:generator:server" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:server"
// 	type: *generator.Server
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// This method can be called even if generator is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetBimaGeneratorServer() *generator.Server {
	o, err := c.UnscopedSafeGetBimaGeneratorServer()
	if err != nil {
		panic(err)
	}
	return o
}

// BimaGeneratorServer retrieves the "bima:generator:server" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:server"
// 	type: *generator.Server
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetBimaGeneratorServer method.
// If the container can not be retrieved, it panics.
func BimaGeneratorServer(i interface{}) *generator.Server {
	return C(i).GetBimaGeneratorServer()
}

// SafeGetBimaGeneratorSwagger retrieves the "bima:generator:swagger" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:swagger"
// 	type: *generator.Swagger
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetBimaGeneratorSwagger() (*generator.Swagger, error) {
	i, err := c.ctn.SafeGet("bima:generator:swagger")
	if err != nil {
		var eo *generator.Swagger
		return eo, err
	}
	o, ok := i.(*generator.Swagger)
	if !ok {
		return o, errors.New("could get 'bima:generator:swagger' because the object could not be cast to *generator.Swagger")
	}
	return o, nil
}

// GetBimaGeneratorSwagger retrieves the "bima:generator:swagger" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:swagger"
// 	type: *generator.Swagger
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetBimaGeneratorSwagger() *generator.Swagger {
	o, err := c.SafeGetBimaGeneratorSwagger()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetBimaGeneratorSwagger retrieves the "bima:generator:swagger" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:swagger"
// 	type: *generator.Swagger
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// This method can be called even if generator is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetBimaGeneratorSwagger() (*generator.Swagger, error) {
	i, err := c.ctn.UnscopedSafeGet("bima:generator:swagger")
	if err != nil {
		var eo *generator.Swagger
		return eo, err
	}
	o, ok := i.(*generator.Swagger)
	if !ok {
		return o, errors.New("could get 'bima:generator:swagger' because the object could not be cast to *generator.Swagger")
	}
	return o, nil
}

// UnscopedGetBimaGeneratorSwagger retrieves the "bima:generator:swagger" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:swagger"
// 	type: *generator.Swagger
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// This method can be called even if generator is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetBimaGeneratorSwagger() *generator.Swagger {
	o, err := c.UnscopedSafeGetBimaGeneratorSwagger()
	if err != nil {
		panic(err)
	}
	return o
}

// BimaGeneratorSwagger retrieves the "bima:generator:swagger" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:generator:swagger"
// 	type: *generator.Swagger
// 	scope: "generator"
// 	build: struct
// 	params: nil
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetBimaGeneratorSwagger method.
// If the container can not be retrieved, it panics.
func BimaGeneratorSwagger(i interface{}) *generator.Swagger {
	return C(i).GetBimaGeneratorSwagger()
}

// SafeGetBimaModuleGenerator retrieves the "bima:module:generator" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:module:generator"
// 	type: *generator.Factory
// 	scope: "generator"
// 	build: func
// 	params:
// 		- "0": Service(generator.Generator) ["bima:generator:dic"]
// 		- "1": Service(generator.Generator) ["bima:generator:model"]
// 		- "2": Service(generator.Generator) ["bima:generator:module"]
// 		- "3": Service(generator.Generator) ["bima:generator:proto"]
// 		- "4": Service(generator.Generator) ["bima:generator:provider"]
// 		- "5": Service(generator.Generator) ["bima:generator:server"]
// 		- "6": Service(generator.Generator) ["bima:generator:swagger"]
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// If the object can not be retrieved, it returns an error.
func (c *Container) SafeGetBimaModuleGenerator() (*generator.Factory, error) {
	i, err := c.ctn.SafeGet("bima:module:generator")
	if err != nil {
		var eo *generator.Factory
		return eo, err
	}
	o, ok := i.(*generator.Factory)
	if !ok {
		return o, errors.New("could get 'bima:module:generator' because the object could not be cast to *generator.Factory")
	}
	return o, nil
}

// GetBimaModuleGenerator retrieves the "bima:module:generator" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:module:generator"
// 	type: *generator.Factory
// 	scope: "generator"
// 	build: func
// 	params:
// 		- "0": Service(generator.Generator) ["bima:generator:dic"]
// 		- "1": Service(generator.Generator) ["bima:generator:model"]
// 		- "2": Service(generator.Generator) ["bima:generator:module"]
// 		- "3": Service(generator.Generator) ["bima:generator:proto"]
// 		- "4": Service(generator.Generator) ["bima:generator:provider"]
// 		- "5": Service(generator.Generator) ["bima:generator:server"]
// 		- "6": Service(generator.Generator) ["bima:generator:swagger"]
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// If the object can not be retrieved, it panics.
func (c *Container) GetBimaModuleGenerator() *generator.Factory {
	o, err := c.SafeGetBimaModuleGenerator()
	if err != nil {
		panic(err)
	}
	return o
}

// UnscopedSafeGetBimaModuleGenerator retrieves the "bima:module:generator" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:module:generator"
// 	type: *generator.Factory
// 	scope: "generator"
// 	build: func
// 	params:
// 		- "0": Service(generator.Generator) ["bima:generator:dic"]
// 		- "1": Service(generator.Generator) ["bima:generator:model"]
// 		- "2": Service(generator.Generator) ["bima:generator:module"]
// 		- "3": Service(generator.Generator) ["bima:generator:proto"]
// 		- "4": Service(generator.Generator) ["bima:generator:provider"]
// 		- "5": Service(generator.Generator) ["bima:generator:server"]
// 		- "6": Service(generator.Generator) ["bima:generator:swagger"]
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// This method can be called even if generator is a sub-scope of the container.
// If the object can not be retrieved, it returns an error.
func (c *Container) UnscopedSafeGetBimaModuleGenerator() (*generator.Factory, error) {
	i, err := c.ctn.UnscopedSafeGet("bima:module:generator")
	if err != nil {
		var eo *generator.Factory
		return eo, err
	}
	o, ok := i.(*generator.Factory)
	if !ok {
		return o, errors.New("could get 'bima:module:generator' because the object could not be cast to *generator.Factory")
	}
	return o, nil
}

// UnscopedGetBimaModuleGenerator retrieves the "bima:module:generator" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:module:generator"
// 	type: *generator.Factory
// 	scope: "generator"
// 	build: func
// 	params:
// 		- "0": Service(generator.Generator) ["bima:generator:dic"]
// 		- "1": Service(generator.Generator) ["bima:generator:model"]
// 		- "2": Service(generator.Generator) ["bima:generator:module"]
// 		- "3": Service(generator.Generator) ["bima:generator:proto"]
// 		- "4": Service(generator.Generator) ["bima:generator:provider"]
// 		- "5": Service(generator.Generator) ["bima:generator:server"]
// 		- "6": Service(generator.Generator) ["bima:generator:swagger"]
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// This method can be called even if generator is a sub-scope of the container.
// If the object can not be retrieved, it panics.
func (c *Container) UnscopedGetBimaModuleGenerator() *generator.Factory {
	o, err := c.UnscopedSafeGetBimaModuleGenerator()
	if err != nil {
		panic(err)
	}
	return o
}

// BimaModuleGenerator retrieves the "bima:module:generator" object from the generator scope.
//
// ---------------------------------------------
// 	name: "bima:module:generator"
// 	type: *generator.Factory
// 	scope: "generator"
// 	build: func
// 	params:
// 		- "0": Service(generator.Generator) ["bima:generator:dic"]
// 		- "1": Service(generator.Generator) ["bima:generator:model"]
// 		- "2": Service(generator.Generator) ["bima:generator:module"]
// 		- "3": Service(generator.Generator) ["bima:generator:proto"]
// 		- "4": Service(generator.Generator) ["bima:generator:provider"]
// 		- "5": Service(generator.Generator) ["bima:generator:server"]
// 		- "6": Service(generator.Generator) ["bima:generator:swagger"]
// 	unshared: false
// 	close: false
// ---------------------------------------------
//
// It tries to find the container with the C method and the given interface.
// If the container can be retrieved, it calls the GetBimaModuleGenerator method.
// If the container can not be retrieved, it panics.
func BimaModuleGenerator(i interface{}) *generator.Factory {
	return C(i).GetBimaModuleGenerator()
}
