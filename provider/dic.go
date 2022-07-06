package provider

import (
	bima "github.com/bimalabs/framework/v4"
	"github.com/bimalabs/framework/v4/generators"
	"github.com/gertd/go-pluralize"
	"github.com/sarulabs/dingo/v4"
)

type Generator struct {
	dingo.BaseProvider
}

func (p *Generator) Load() error {
	if err := p.AddDefSlice(dic); err != nil {
		return err
	}

	return nil
}

var dic = []dingo.Def{
	{
		Name:  "bima:module:generator",
		Scope: bima.Generator,
		Build: func(
			dic generators.Generator,
			model generators.Generator,
			module generators.Generator,
			proto generators.Generator,
			provider generators.Generator,
			server generators.Generator,
			swagger generators.Generator,
		) (*generators.Factory, error) {
			return &generators.Factory{
				Pluralizer: pluralize.NewClient(),
				Template:   &generators.Template{},
				Generators: []generators.Generator{dic, model, module, proto, provider, server, swagger},
			}, nil
		},
		Params: dingo.Params{
			"0": dingo.Service("bima:generator:dic"),
			"1": dingo.Service("bima:generator:model"),
			"2": dingo.Service("bima:generator:module"),
			"3": dingo.Service("bima:generator:proto"),
			"4": dingo.Service("bima:generator:provider"),
			"5": dingo.Service("bima:generator:server"),
			"6": dingo.Service("bima:generator:swagger"),
		},
	},
	{
		Name:  "bima:generator:dic",
		Scope: bima.Generator,
		Build: (*generators.Dic)(nil),
	},
	{
		Name:  "bima:generator:model",
		Scope: bima.Generator,
		Build: (*generators.Model)(nil),
	},
	{
		Name:  "bima:generator:module",
		Scope: bima.Generator,
		Build: (*generators.Module)(nil),
	},
	{
		Name:  "bima:generator:proto",
		Scope: bima.Generator,
		Build: (*generators.Proto)(nil),
	},
	{
		Name:  "bima:generator:provider",
		Scope: bima.Generator,
		Build: (*generators.Provider)(nil),
	},
	{
		Name:  "bima:generator:server",
		Scope: bima.Generator,
		Build: (*generators.Server)(nil),
	},
	{
		Name:  "bima:generator:swagger",
		Scope: bima.Generator,
		Build: (*generators.Swagger)(nil),
	},
}
