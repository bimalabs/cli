package engine

import (
	"errors"

	"github.com/sarulabs/di/v2"
	"github.com/sarulabs/dingo/v4"

	generator "github.com/bimalabs/cli/generator"
)

func getDiDefs(provider dingo.Provider) []di.Def {
	return []di.Def{
		{
			Name:  "bima:generator:dic",
			Scope: "generator",
			Build: func(ctn di.Container) (interface{}, error) {
				return &generator.Dic{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "bima:generator:model",
			Scope: "generator",
			Build: func(ctn di.Container) (interface{}, error) {
				return &generator.Model{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "bima:generator:module",
			Scope: "generator",
			Build: func(ctn di.Container) (interface{}, error) {
				var p0 []string
				return &generator.Module{
					Config: p0,
				}, nil
			},
			Unshared: false,
		},
		{
			Name:  "bima:generator:proto",
			Scope: "generator",
			Build: func(ctn di.Container) (interface{}, error) {
				return &generator.Proto{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "bima:generator:provider",
			Scope: "generator",
			Build: func(ctn di.Container) (interface{}, error) {
				return &generator.Provider{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "bima:generator:server",
			Scope: "generator",
			Build: func(ctn di.Container) (interface{}, error) {
				return &generator.Server{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "bima:generator:swagger",
			Scope: "generator",
			Build: func(ctn di.Container) (interface{}, error) {
				return &generator.Swagger{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "bima:module:generator",
			Scope: "generator",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("bima:module:generator")
				if err != nil {
					var eo *generator.Factory
					return eo, err
				}
				pi0, err := ctn.SafeGet("bima:generator:dic")
				if err != nil {
					var eo *generator.Factory
					return eo, err
				}
				p0, ok := pi0.(generator.Generator)
				if !ok {
					var eo *generator.Factory
					return eo, errors.New("could not cast parameter 0 to generator.Generator")
				}
				pi1, err := ctn.SafeGet("bima:generator:model")
				if err != nil {
					var eo *generator.Factory
					return eo, err
				}
				p1, ok := pi1.(generator.Generator)
				if !ok {
					var eo *generator.Factory
					return eo, errors.New("could not cast parameter 1 to generator.Generator")
				}
				pi2, err := ctn.SafeGet("bima:generator:module")
				if err != nil {
					var eo *generator.Factory
					return eo, err
				}
				p2, ok := pi2.(generator.Generator)
				if !ok {
					var eo *generator.Factory
					return eo, errors.New("could not cast parameter 2 to generator.Generator")
				}
				pi3, err := ctn.SafeGet("bima:generator:proto")
				if err != nil {
					var eo *generator.Factory
					return eo, err
				}
				p3, ok := pi3.(generator.Generator)
				if !ok {
					var eo *generator.Factory
					return eo, errors.New("could not cast parameter 3 to generator.Generator")
				}
				pi4, err := ctn.SafeGet("bima:generator:provider")
				if err != nil {
					var eo *generator.Factory
					return eo, err
				}
				p4, ok := pi4.(generator.Generator)
				if !ok {
					var eo *generator.Factory
					return eo, errors.New("could not cast parameter 4 to generator.Generator")
				}
				pi5, err := ctn.SafeGet("bima:generator:server")
				if err != nil {
					var eo *generator.Factory
					return eo, err
				}
				p5, ok := pi5.(generator.Generator)
				if !ok {
					var eo *generator.Factory
					return eo, errors.New("could not cast parameter 5 to generator.Generator")
				}
				pi6, err := ctn.SafeGet("bima:generator:swagger")
				if err != nil {
					var eo *generator.Factory
					return eo, err
				}
				p6, ok := pi6.(generator.Generator)
				if !ok {
					var eo *generator.Factory
					return eo, errors.New("could not cast parameter 6 to generator.Generator")
				}
				b, ok := d.Build.(func(generator.Generator, generator.Generator, generator.Generator, generator.Generator, generator.Generator, generator.Generator, generator.Generator) (*generator.Factory, error))
				if !ok {
					var eo *generator.Factory
					return eo, errors.New("could not cast build function to func(generator.Generator, generator.Generator, generator.Generator, generator.Generator, generator.Generator, generator.Generator, generator.Generator) (*generator.Factory, error)")
				}
				return b(p0, p1, p2, p3, p4, p5, p6)
			},
			Unshared: false,
		},
	}
}
