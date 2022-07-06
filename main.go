package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
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
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
	"github.com/vito/go-interact/interact"
	"golang.org/x/mod/modfile"
	"gopkg.in/yaml.v2"
)

func main() {
	var file string
	app := &cli.App{
		Name:  "Bima Cli",
		Usage: "Bima Framework Toolkit",
		Commands: []*cli.Command{
			{
				Name:    "module",
				Aliases: []string{"m"},
				Usage:   "module",
				Subcommands: []*cli.Command{
					{
						Name: "add",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "file",
								Value:       ".env",
								Usage:       "Config file",
								Destination: &file,
							},
						},
						Aliases: []string{"a"},
						Usage:   "module add <name>",
						Action: func(cCtx *cli.Context) error {
							module := cCtx.Args().First()
							if module == "" {
								fmt.Println("Usage: bima module add <name>")

								return nil
							}

							config := configs.Env{}
							env(&config, file, filepath.Ext(file))

							container, err := generator.NewContainer(bima.Generator)
							if err != nil {
								return err
							}

							generator := container.GetBimaModuleGenerator()
							generator.Driver = config.Db.Driver
							generator.ApiVersion = "v1"
							if cCtx.NArg() > 1 {
								generator.ApiVersion = cCtx.Args().Get(1)
							}

							util := color.New(color.FgCyan, color.Bold)

							register(generator, util, module)

							if err := genproto(); err != nil {
								color.New(color.FgRed).Println("Error generate code from proto files")
								os.Exit(1)
							}

							if err := clean(); err != nil {
								color.New(color.FgRed).Println("Error update dependencies")

								return err
							}

							if err := dump(); err != nil {
								color.New(color.FgRed).Println("Error update DI container")

								return err
							}

							util = color.New(color.FgGreen)
							util.Print("By: ")
							util.Println("Bimalabs")

							return nil
						},
					},
					{
						Name:    "remove",
						Aliases: []string{"r"},
						Usage:   "module remove <name>",
						Action: func(cCtx *cli.Context) error {
							module := cCtx.Args().First()
							if module == "" {
								fmt.Println("Usage: bima module add <name>")

								return nil
							}

							util := color.New(color.FgCyan, color.Bold)

							unregister(util, module)
							if err := dump(); err != nil {
								color.New(color.FgRed).Println("Error update DI container")

								return err
							}

							if err := clean(); err != nil {
								color.New(color.FgRed).Println("Error update dependencies")

								return err
							}

							util = color.New(color.FgGreen)
							util.Print("By: ")
							util.Println("Bimalabs")

							return nil
						},
					},
				},
			},
			{
				Name:    "dump",
				Aliases: []string{"d"},
				Usage:   "dump",
				Action: func(*cli.Context) error {
					return dump()
				},
			},
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "version",
				Action: func(*cli.Context) error {
					fmt.Printf("Framework: %s\n", bima.Version)
					fmt.Println("Cli: v1.0.0")

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func dump() error {
	_, err := exec.Command("go", "run", "dumper/main.go").Output()

	return err
}

func clean() error {
	_, err := exec.Command("go", "mod", "tidy").Output()

	return err
}

func genproto() error {
	_, err := exec.Command("sh", "proto_gen.sh").Output()

	return err
}

func env(config *configs.Env, filePath string, ext string) {
	switch ext {
	case ".env":
		godotenv.Load()
		parse(config)
	case ".yaml":
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalln(err.Error())
		}

		err = yaml.Unmarshal(content, config)
		if err != nil {
			log.Fatalln(err.Error())
		}
	case ".json":
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalln(err.Error())
		}

		err = json.Unmarshal(content, config)
		if err != nil {
			log.Fatalln(err.Error())
		}
	}

	if config.Secret == "" {
		hasher := sha256.New()
		hasher.Write([]byte(time.Now().Format(time.RFC3339)))

		config.Secret = base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	}
}

func parse(config *configs.Env) {
	config.Secret = os.Getenv("APP_SECRET")
	config.Debug, _ = strconv.ParseBool(os.Getenv("APP_DEBUG"))
	config.HttpPort, _ = strconv.Atoi(os.Getenv("APP_PORT"))
	config.RpcPort, _ = strconv.Atoi(os.Getenv("GRPC_PORT"))

	sName := os.Getenv("APP_NAME")
	config.Service = configs.Service{
		Name:           sName,
		ConnonicalName: strcase.ToDelimited(sName, '_'),
	}

	dbPort, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	config.Db = configs.Db{
		Host:     os.Getenv("DB_HOST"),
		Port:     dbPort,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		Driver:   os.Getenv("DB_DRIVER"),
	}

	config.CacheLifetime, _ = strconv.Atoi(os.Getenv("CACHE_LIFETIME"))
}

func register(generator *generators.Factory, util *color.Color, name string) {
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
			util.Println(err.Error())
			os.Exit(1)
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
		util.Println("You must have at least one column in table")
		os.Exit(1)
	}

	generator.Generate(module)

	workDir, _ := os.Getwd()
	util.Println(fmt.Sprintf("Module registered in %s/modules.yaml", workDir))
}

func unregister(util *color.Color, module string) {
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

	util.Println("Module deleted")
}

func column(util *color.Color, field *generators.FieldTemplate, mapType utils.Type) {
	err := interact.NewInteraction("Input column name?").Resolve(&field.Name)
	if err != nil {
		util.Println(err.Error())
		os.Exit(1)
	}

	if field.Name == "" {
		util.Println("Column name is required")
		column(util, field, mapType)
	}

	field.ProtobufType = "string"
	err = interact.NewInteraction("Input data type?",
		interact.Choice{Display: "double", Value: "double"},
		interact.Choice{Display: "float", Value: "float"},
		interact.Choice{Display: "int32", Value: "int32"},
		interact.Choice{Display: "int64", Value: "int64"},
		interact.Choice{Display: "uint32", Value: "uint32"},
		interact.Choice{Display: "sint32", Value: "sint32"},
		interact.Choice{Display: "sint64", Value: "sint64"},
		interact.Choice{Display: "fixed32", Value: "fixed32"},
		interact.Choice{Display: "fixed64", Value: "fixed64"},
		interact.Choice{Display: "sfixed32", Value: "sfixed32"},
		interact.Choice{Display: "sfixed64", Value: "sfixed64"},
		interact.Choice{Display: "bool", Value: "bool"},
		interact.Choice{Display: "string", Value: "string"},
		interact.Choice{Display: "bytes", Value: "bytes"},
	).Resolve(&field.ProtobufType)
	if err != nil {
		util.Println(err.Error())
		os.Exit(1)
	}

	field.GolangType = mapType.Value(field.ProtobufType)
	field.IsRequired = true
	err = interact.NewInteraction("Is column required?").Resolve(&field.IsRequired)
	if err != nil {
		util.Println(err.Error())
		os.Exit(1)
	}
}
