package {{.ModulePluralLowercase}}

import "github.com/bimalabs/framework/v4"

type {{.Module}} struct {
	*bima.GormModel
{{range .Columns}}
    {{.Name}} {{.GolangType}} {{if .IsRequired}}`validate:"required"`{{end}}
{{end}}
}

func (m *{{.Module}}) TableName() string {
	return "{{.ModuleLowercase}}"
}

func (m *{{.Module}}) IsSoftDelete() bool {
	return true
}