package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
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
				Name:    "create",
				Aliases: []string{"new", "c", "n"},
				Usage:   "bima create <command>",
				Subcommands: []*cli.Command{
					{
						Name:    "app",
						Aliases: []string{"project", "skeleton"},
						Usage:   "bima create app <name>",
						Action: func(cCtx *cli.Context) error {
							name := cCtx.Args().First()
							if name == "" {
								fmt.Println("Usage: bima create app <name>")

								return nil
							}

							fmt.Println("Creating new application...")

							err := create(name)
							if err == nil {
								util := color.New(color.Bold)

								fmt.Print("Move to ")
								util.Print(name)
								fmt.Println(" folder and type ")
								util.Println("bima run")
							}

							return err
						},
					},
				},
			},
			{
				Name:    "module",
				Aliases: []string{"m"},
				Usage:   "module <command>",
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

							if err := dump(); err != nil {
								color.New(color.FgRed).Println("Error update DI container")

								return err
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

							err = register(generator, util, module)
							if err != nil {
								color.New(color.FgRed).Println(err)
							}

							if err = genproto(); err != nil {
								color.New(color.FgRed).Println("Error generate code from proto files")

								return err
							}

							if err = clean(); err != nil {
								color.New(color.FgRed).Println("Error cleaning dependencies")

								return err
							}

							if err = dump(); err != nil {
								color.New(color.FgRed).Println("Error update DI container")

								return err
							}

							if err = clean(); err != nil {
								color.New(color.FgRed).Println("Error cleaning dependencies")

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

							if err := dump(); err != nil {
								color.New(color.FgRed).Println("Error cleaning DI container")

								return err
							}

							util := color.New(color.FgCyan, color.Bold)

							unregister(util, module)
							if err := dump(); err != nil {
								color.New(color.FgRed).Println("Error update DI container")

								return err
							}

							if err := clean(); err != nil {
								color.New(color.FgRed).Println("Error cleaning dependencies")

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
					fmt.Println("Generate dic codes...")

					return dump()
				},
			},
			{
				Name:    "update",
				Aliases: []string{"u"},
				Usage:   "update",
				Action: func(*cli.Context) error {
					fmt.Println("Updating dependencies...")
					if err := update(); err != nil {
						color.New(color.FgRed).Println("Error update dependencies")

						return err
					}

					if err := dump(); err != nil {
						color.New(color.FgRed).Println("Error update DI container")

						return err
					}

					return nil
				},
			},
			{
				Name:    "clean",
				Aliases: []string{"c"},
				Usage:   "clean",
				Action: func(*cli.Context) error {
					fmt.Println("Cleaning dependencies...")
					if err := clean(); err != nil {
						color.New(color.FgRed).Println("Error cleaning dependencies")

						return err
					}

					if err := dump(); err != nil {
						color.New(color.FgRed).Println("Error update DI container")

						return err
					}

					return nil
				},
			},
			{
				Name:    "generate",
				Aliases: []string{"gen", "genproto", "g"},
				Usage:   "generate",
				Action: func(*cli.Context) error {
					fmt.Println("Generating protobuff...")
					if err := genproto(); err != nil {
						color.New(color.FgRed).Println("Error generate protobuff")

						return err
					}

					if err := clean(); err != nil {
						color.New(color.FgRed).Println("Error cleaning dependencies")

						return err
					}

					if err := dump(); err != nil {
						color.New(color.FgRed).Println("Error update DI container")

						return err
					}

					return nil
				},
			},
			{
				Name: "run",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "file",
						Value:       ".env",
						Usage:       "Config file",
						Destination: &file,
					},
				},
				Aliases: []string{"r"},
				Usage:   "run",
				Action: func(*cli.Context) error {
					return run(file)
				},
			},
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "version",
				Action: func(*cli.Context) error {
					wd, _ := os.Getwd()
					var path strings.Builder

					path.WriteString(wd)
					path.WriteString("/go.mod")

					mod, err := os.ReadFile(path.String())
					if err != nil {
						return err
					}

					f, err := modfile.Parse(path.String(), mod, nil)
					if err != nil {
						return err
					}

					version := "unknown"
					for _, v := range f.Require {
						if v.Mod.Path == "github.com/bimalabs/framework/v4" {
							version = v.Mod.Version

							break
						}
					}

					fmt.Printf("Framework: %s\n", version)
					fmt.Println("Cli: v1.0.3")

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
func create(name string) error {
	output, err := exec.Command("git", "clone", "--depth", "1", "https://github.com/bimalabs/skeleton.git", name).CombinedOutput()
	if err != nil {
		fmt.Println(string(output))

		return err
	}

	output, err = exec.Command("rm", "-rf", fmt.Sprintf("%s/.git", name)).CombinedOutput()
	if err != nil {
		fmt.Println(string(output))

		return err
	}

	output, err = exec.Command("cp", fmt.Sprintf("%s/.example.env", name), fmt.Sprintf("%s/.env", name)).CombinedOutput()
	if err != nil {
		fmt.Println(string(output))

		return err
	}

	wd, _ := os.Getwd()

	cmd := exec.Command("go", "get")
	cmd.Dir = fmt.Sprintf("%s/%s", wd, name)
	output, _ = cmd.CombinedOutput()

	cmd = exec.Command("go", "run", "dumper/main.go")
	cmd.Dir = fmt.Sprintf("%s/%s", wd, name)
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))

		return err
	}

	cmd = exec.Command("go", "get")
	cmd.Dir = fmt.Sprintf("%s/%s", wd, name)
	output, _ = cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))

		return err
	}

	return err
}

func dump() error {
	_, err := exec.Command("go", "run", "dumper/main.go").Output()

	return err
}

func clean() error {
	_, err := exec.Command("go", "mod", "tidy").Output()

	return err
}

func update() error {
	_, err := exec.Command("go", "get", "-u").Output()

	return err
}

func run(file string) error {
	cmd := exec.Command("go", "run", "cmd/main.go", file)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s\n", scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
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

func register(generator *generators.Factory, util *color.Color, name string) error {
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

	fmt.Print("Module ")
	util.Print(module)
	util.Println(" deleted")
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
