package generator

import (
	"errors"

	"github.com/sarulabs/di/v2"
	"github.com/sarulabs/dingo/v4"

	generators "github.com/bimalabs/framework/v4/generators"
)

func getDiDefs(provider dingo.Provider) []di.Def {
	return []di.Def{
		{
			Name:  "bima:generator:dic",
			Scope: "generator",
			Build: func(ctn di.Container) (interface{}, error) {
				return &generators.Dic{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "bima:generator:model",
			Scope: "generator",
			Build: func(ctn di.Container) (interface{}, error) {
				return &generators.Model{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "bima:generator:module",
			Scope: "generator",
			Build: func(ctn di.Container) (interface{}, error) {
				var p0 []string
				return &generators.Module{
					Config: p0,
				}, nil
			},
			Unshared: false,
		},
		{
			Name:  "bima:generator:proto",
			Scope: "generator",
			Build: func(ctn di.Container) (interface{}, error) {
				return &generators.Proto{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "bima:generator:provider",
			Scope: "generator",
			Build: func(ctn di.Container) (interface{}, error) {
				return &generators.Provider{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "bima:generator:server",
			Scope: "generator",
			Build: func(ctn di.Container) (interface{}, error) {
				return &generators.Server{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "bima:generator:swagger",
			Scope: "generator",
			Build: func(ctn di.Container) (interface{}, error) {
				return &generators.Swagger{}, nil
			},
			Unshared: false,
		},
		{
			Name:  "bima:module:generator",
			Scope: "generator",
			Build: func(ctn di.Container) (interface{}, error) {
				d, err := provider.Get("bima:module:generator")
				if err != nil {
					var eo *generators.Factory
					return eo, err
				}
				pi0, err := ctn.SafeGet("bima:generator:dic")
				if err != nil {
					var eo *generators.Factory
					return eo, err
				}
				p0, ok := pi0.(generators.Generator)
				if !ok {
					var eo *generators.Factory
					return eo, errors.New("could not cast parameter 0 to generators.Generator")
				}
				pi1, err := ctn.SafeGet("bima:generator:model")
				if err != nil {
					var eo *generators.Factory
					return eo, err
				}
				p1, ok := pi1.(generators.Generator)
				if !ok {
					var eo *generators.Factory
					return eo, errors.New("could not cast parameter 1 to generators.Generator")
				}
				pi2, err := ctn.SafeGet("bima:generator:module")
				if err != nil {
					var eo *generators.Factory
					return eo, err
				}
				p2, ok := pi2.(generators.Generator)
				if !ok {
					var eo *generators.Factory
					return eo, errors.New("could not cast parameter 2 to generators.Generator")
				}
				pi3, err := ctn.SafeGet("bima:generator:proto")
				if err != nil {
					var eo *generators.Factory
					return eo, err
				}
				p3, ok := pi3.(generators.Generator)
				if !ok {
					var eo *generators.Factory
					return eo, errors.New("could not cast parameter 3 to generators.Generator")
				}
				pi4, err := ctn.SafeGet("bima:generator:provider")
				if err != nil {
					var eo *generators.Factory
					return eo, err
				}
				p4, ok := pi4.(generators.Generator)
				if !ok {
					var eo *generators.Factory
					return eo, errors.New("could not cast parameter 4 to generators.Generator")
				}
				pi5, err := ctn.SafeGet("bima:generator:server")
				if err != nil {
					var eo *generators.Factory
					return eo, err
				}
				p5, ok := pi5.(generators.Generator)
				if !ok {
					var eo *generators.Factory
					return eo, errors.New("could not cast parameter 5 to generators.Generator")
				}
				pi6, err := ctn.SafeGet("bima:generator:swagger")
				if err != nil {
					var eo *generators.Factory
					return eo, err
				}
				p6, ok := pi6.(generators.Generator)
				if !ok {
					var eo *generators.Factory
					return eo, errors.New("could not cast parameter 6 to generators.Generator")
				}
				b, ok := d.Build.(func(generators.Generator, generators.Generator, generators.Generator, generators.Generator, generators.Generator, generators.Generator, generators.Generator) (*generators.Factory, error))
				if !ok {
					var eo *generators.Factory
					return eo, errors.New("could not cast build function to func(generators.Generator, generators.Generator, generators.Generator, generators.Generator, generators.Generator, generators.Generator, generators.Generator) (*generators.Factory, error)")
				}
				return b(p0, p1, p2, p3, p4, p5, p6)
			},
			Unshared: false,
		},
	}
}
