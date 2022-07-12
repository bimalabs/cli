package tool

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bimalabs/cli/generated/generator"
	bima "github.com/bimalabs/framework/v4"
	"github.com/bimalabs/framework/v4/configs"
	"github.com/bimalabs/framework/v4/generators"
	"github.com/bimalabs/framework/v4/parsers"
	"github.com/bimalabs/framework/v4/utils"
	"github.com/fatih/color"
	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/copier"
	"github.com/vito/go-interact/interact"
	"golang.org/x/mod/modfile"
)

type Module string

func (m Module) Create(file string, version string) error {
	if err := Call("dump"); err != nil {
		color.New(color.FgRed).Println("Error update DI container")

		return err
	}

	env := configs.Env{}
	config(&env, file, filepath.Ext(file))

	container, err := generator.NewContainer(bima.Generator)
	if err != nil {
		color.New(color.FgRed).Println(err.Error())

		return err
	}

	generator := container.GetBimaModuleGenerator()
	generator.Driver = env.Db.Driver
	generator.ApiVersion = version

	termColor := color.New(color.FgGreen, color.Bold)
	err = create(generator, termColor, string(m))
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		m.Remove()

		return err
	}

	if err = Call("genproto"); err != nil {
		color.New(color.FgRed).Println("Error generate code from proto files")
		m.Remove()

		return err
	}

	if err = Call("clean"); err != nil {
		color.New(color.FgRed).Println("Error cleaning dependencies")
		m.Remove()

		return err
	}

	if err = Call("dump"); err != nil {
		color.New(color.FgRed).Println("Error update DI container")
		m.Remove()

		return err
	}

	if err = Call("clean"); err != nil {
		color.New(color.FgRed).Println("Error cleaning dependencies")
		m.Remove()

		return err
	}

	return nil
}

func (m Module) Remove() error {
	remove(string(m))
	if err := Call("dump"); err != nil {
		color.New(color.FgRed).Println("Error update DI container")

		return err
	}

	if err := Call("clean"); err != nil {
		color.New(color.FgRed).Println("Error cleaning dependencies")

		return err
	}

	return nil
}

func remove(module string) {
	util := color.New(color.FgGreen, color.Bold)
	workDir, _ := os.Getwd()
	pluralizer := pluralize.NewClient()
	moduleName := strcase.ToCamel(pluralizer.Singular(module))
	modulePlural := strcase.ToDelimited(pluralizer.Plural(moduleName), '_')
	moduleUnderscore := strcase.ToDelimited(module, '_')
	list := parsers.ParseModule(workDir)

	exist := false
	for _, v := range list {
		if v == fmt.Sprintf("module:%s", moduleUnderscore) {
			exist = true
			break
		}
	}

	if !exist {
		util.Println("Module is not registered")
		return
	}

	mod, err := os.ReadFile(fmt.Sprintf("%s/go.mod", workDir))
	if err != nil {
		panic(err)
	}

	jsonModules := fmt.Sprintf("%s/swaggers/modules.json", workDir)
	file, _ := os.ReadFile(jsonModules)
	modulesJson := []generators.ModuleJson{}
	registered := modulesJson
	json.Unmarshal(file, &modulesJson)
	for _, v := range modulesJson {
		if v.Name != moduleName {
			mUrl, _ := url.Parse(v.Url)
			query := mUrl.Query()

			query.Set("v", strconv.Itoa(int(time.Now().UnixMicro())))
			mUrl.RawQuery = query.Encode()
			v.Url = mUrl.String()
			registered = append(registered, v)
		}
	}

	registeredByte, _ := json.Marshal(registered)
	os.WriteFile(jsonModules, registeredByte, 0644)

	packageName := modfile.ModulePath(mod)
	yaml := fmt.Sprintf("%s/configs/modules.yaml", workDir)
	file, _ = os.ReadFile(yaml)
	modules := string(file)

	provider := fmt.Sprintf("%s/configs/provider.go", workDir)
	file, _ = os.ReadFile(provider)
	codeblock := string(file)

	modRegex := regexp.MustCompile(fmt.Sprintf("(?m)[\r\n]+^.*module:%s.*$", moduleUnderscore))
	modules = modRegex.ReplaceAllString(modules, "")
	os.WriteFile(yaml, []byte(modules), 0644)

	regex := regexp.MustCompile(fmt.Sprintf("(?m)[\r\n]+^.*%s.*$", fmt.Sprintf("%s/%s", packageName, modulePlural)))
	codeblock = regex.ReplaceAllString(codeblock, "")

	codeblock = modRegex.ReplaceAllString(codeblock, "")
	os.WriteFile(provider, []byte(codeblock), 0644)

	os.RemoveAll(fmt.Sprintf("%s/%s", workDir, modulePlural))
	os.Remove(fmt.Sprintf("%s/protos/%s.proto", workDir, moduleUnderscore))
	os.Remove(fmt.Sprintf("%s/protos/builds/%s_grpc.pb.go", workDir, moduleUnderscore))
	os.Remove(fmt.Sprintf("%s/protos/builds/%s.pb.go", workDir, moduleUnderscore))
	os.Remove(fmt.Sprintf("%s/protos/builds/%s.pb.gw.go", workDir, moduleUnderscore))
	os.Remove(fmt.Sprintf("%s/swaggers/%s.swagger.json", workDir, moduleUnderscore))

	fmt.Print("Module ")
	util.Print(module)
	util.Println(" deleted")
}

func create(generator *generators.Factory, util *color.Color, name string) error {
	module := generators.ModuleTemplate{}
	field := generators.FieldTemplate{}
	mapType := utils.NewType()

	util.Println("Welcome to Bima Skeleton Module Generator")
	module.Name = name

	index := 2
	more := true
	for more {
		err := interact.NewInteraction("Add new column?").Resolve(&more)
		if err != nil {
			color.New(color.FgRed).Println(err.Error())

			return err
		}

		if more {
			column(util, &field, mapType)

			field.Name = strings.Replace(field.Name, " ", "", -1)
			column := generators.FieldTemplate{}

			copier.Copy(&column, field)

			column.Index = index
			column.Name = strings.Title(column.Name)
			column.NameUnderScore = strcase.ToDelimited(column.Name, '_')
			module.Fields = append(module.Fields, &column)

			field.Name = ""
			field.ProtobufType = ""

			index++
		}
	}

	if len(module.Fields) < 1 {
		return errors.New("You must have at least one column in table")
	}

	generator.Generate(module)

	workDir, _ := os.Getwd()
	fmt.Print("Module ")
	util.Print(name)
	fmt.Printf(" registered in %s/modules.yaml\n", workDir)

	return nil
}

func column(util *color.Color, field *generators.FieldTemplate, mapType utils.Type) {
	err := interact.NewInteraction("Input column name?").Resolve(&field.Name)
	if err != nil {
		util.Println(err.Error())
		column(util, field, mapType)
	}

	if field.Name == "" {
		util.Println("Column name is required")
		column(util, field, mapType)
	}

	field.ProtobufType = "string"
	err = interact.NewInteraction("Input data type?",
		interact.Choice{Display: "string", Value: "string"},
		interact.Choice{Display: "bool", Value: "bool"},
		interact.Choice{Display: "int32", Value: "int32"},
		interact.Choice{Display: "int64", Value: "int64"},
		interact.Choice{Display: "bytes", Value: "bytes"},
		interact.Choice{Display: "double", Value: "double"},
		interact.Choice{Display: "float", Value: "float"},
		interact.Choice{Display: "uint32", Value: "uint32"},
		interact.Choice{Display: "sint32", Value: "sint32"},
		interact.Choice{Display: "sint64", Value: "sint64"},
		interact.Choice{Display: "fixed32", Value: "fixed32"},
		interact.Choice{Display: "fixed64", Value: "fixed64"},
		interact.Choice{Display: "sfixed32", Value: "sfixed32"},
		interact.Choice{Display: "sfixed64", Value: "sfixed64"},
	).Resolve(&field.ProtobufType)
	if err != nil {
		util.Println(err.Error())
		column(util, field, mapType)
	}

	field.GolangType = mapType.Value(field.ProtobufType)
	field.IsRequired = true
	err = interact.NewInteraction("Is column required?").Resolve(&field.IsRequired)
	if err != nil {
		util.Println(err.Error())
		column(util, field, mapType)
	}
}
