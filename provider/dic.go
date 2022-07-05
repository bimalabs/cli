package provider

import (
	bima "github.com/bimalabs/framework/v4/dics"
	"github.com/sarulabs/dingo/v4"
)

type Generator struct {
	dingo.BaseProvider
}

func (p *Generator) Load() error {
	if err := p.AddDefSlice(bima.Generator); err != nil {
		return err
	}

	return nil
}
