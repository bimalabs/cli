package generator

import (
	"log"
	"os"
	"strings"
	engine "text/template"

	"gopkg.in/yaml.v2"
)

const c = "configs/modules.yaml"

type (
	Module struct {
		Config []string `yaml:"modules"`
	}
)

func (g *Module) Generate(template *Template, modulePath string, packagePath string, templatePath string) {
	var str strings.Builder
	str.WriteString(packagePath)
	str.WriteString("/")
	str.WriteString(templatePath)
	str.WriteString("/module.tpl")

	moduleTemplate, err := engine.ParseFiles(str.String())
	if err != nil {
		panic(err)
	}

	str.Reset()
	str.WriteString(modulePath)
	str.WriteString("/module.go")

	moduleFile, err := os.Create(str.String())
	if err != nil {
		panic(err)
	}

	str.Reset()
	str.WriteString("module:")
	str.WriteString(template.ModuleLowercase)

	workDir, _ := os.Getwd()
	g.Config = g.parse(workDir)
	g.Config = append(g.Config, str.String())
	g.Config = g.makeUnique(g.Config)

	modules, err := yaml.Marshal(g)
	if err != nil {
		panic(err)
	}

	str.Reset()
	str.WriteString(workDir)
	str.WriteString("/")
	str.WriteString(c)

	err = os.WriteFile(str.String(), modules, 0644)
	if err != nil {
		panic(err)
	}

	moduleTemplate.Execute(moduleFile, template)
}

func (g *Module) makeUnique(modules []string) []string {
	exists := make(map[string]bool)
	var result []string
	for e := range modules {
		if exists[modules[e]] != true {
			exists[modules[e]] = true

			result = append(result, modules[e])
		}
	}

	return result
}

func (g *Module) parse(dir string) []string {
	var path strings.Builder
	path.WriteString(dir)
	path.WriteString("/")
	path.WriteString(c)

	config, err := os.ReadFile(path.String())
	if err != nil {
		log.Println(err)

		return []string{}
	}

	err = yaml.Unmarshal(config, g)
	if err != nil {
		log.Println(err)

		return []string{}
	}

	return g.Config
}
