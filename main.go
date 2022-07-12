package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bimalabs/cli/tool"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"golang.org/x/mod/modfile"
	"mvdan.cc/sh/interp"
	"mvdan.cc/sh/syntax"
)

var (
	version              = "v1.2.4"
	protocMinVersion     = 31900
	protocGoMinVersion   = 12800
	protocGRpcMinVersion = 10200
	spinerIndex          = 9
	duration             = 77 * time.Millisecond
)

func main() {
	file := ""
	app := &cli.App{
		Name:                 "Bima Cli",
		Usage:                "Bima Framework Toolkit",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:    "create",
				Aliases: []string{"new"},
				Usage:   "bima create <command>",
				Subcommands: []*cli.Command{
					{
						Name:        "project",
						Aliases:     []string{"app"},
						Usage:       "bima create app <name>",
						Description: "Create new application or project",
						Action: func(cCtx *cli.Context) error {
							name := cCtx.Args().First()
							if name == "" {
								fmt.Println("Usage: bima create app <name>")

								return nil
							}

							return tool.App(name).Create()
						},
					},
					{
						Name:        "middleware",
						Aliases:     []string{"mid"},
						Usage:       "bima create middleware <name>",
						Description: "Create new middleware",
						Action: func(cCtx *cli.Context) error {
							name := cCtx.Args().First()
							if name == "" {
								fmt.Println("Usage: bima create middleware <name>")

								return nil
							}

							return tool.Middleware(name).Create()
						},
					},
					{
						Name:        "driver",
						Aliases:     []string{"dvr"},
						Usage:       "bima create driver <name>",
						Description: "Create new driver",
						Action: func(cCtx *cli.Context) error {
							name := cCtx.Args().First()
							if name == "" {
								fmt.Println("Usage: bima create driver <name>")

								return nil
							}

							return tool.Driver(name).Create()
						},
					},
					{
						Name:        "adapter",
						Aliases:     []string{"adp"},
						Usage:       "bima create adapter <name>",
						Description: "Create new adapter",
						Action: func(cCtx *cli.Context) error {
							name := cCtx.Args().First()
							if name == "" {
								fmt.Println("Usage: bima create adapter <name>")

								return nil
							}

							return tool.Adapter(name).Create()
						},
					},
					{
						Name:        "route",
						Aliases:     []string{"rt"},
						Usage:       "bima create route <name>",
						Description: "Create new route",
						Action: func(cCtx *cli.Context) error {
							name := cCtx.Args().First()
							if name == "" {
								fmt.Println("Usage: bima create route <name>")

								return nil
							}

							return tool.Route(name).Create()
						},
					},
				},
			},
			{
				Name:    "module",
				Aliases: []string{"mod"},
				Usage:   "module <command>",
				Subcommands: []*cli.Command{
					{
						Name: "add",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "config",
								Aliases:     []string{"c"},
								Value:       ".env",
								Usage:       "Config file",
								Destination: &file,
							},
						},
						Aliases:     []string{"new"},
						Usage:       "module add <name> [<version> -c <config>]",
						Description: "Create new module <name> with <version> using <config> file",
						Action: func(cCtx *cli.Context) error {
							name := cCtx.Args().First()
							if name == "" {
								fmt.Println("Usage: bima module add <name> [<version> -c <config>]")

								return nil
							}

							version := "v1"
							if cCtx.NArg() > 1 {
								version = cCtx.Args().Get(1)
							}

							return tool.Module(name).Create(file, version)
						},
					},
					{
						Name:        "remove",
						Aliases:     []string{"rm", "rem"},
						Usage:       "module remove <name>",
						Description: "Remove module <name>",
						Action: func(cCtx *cli.Context) error {
							name := cCtx.Args().First()
							if name == "" {
								fmt.Println("Usage: bima module remove <name>")

								return nil
							}

							return tool.Module(name).Remove()
						},
					},
				},
			},
			{
				Name:        "dump",
				Aliases:     []string{"dmp"},
				Usage:       "dump",
				Description: "Generate service container",
				Action: func(*cli.Context) error {
					progress := spinner.New(spinner.CharSets[spinerIndex], duration)
					progress.Suffix = " Generate service container... "
					progress.Start()
					time.Sleep(1 * time.Second)

					err := tool.Call("dump")
					progress.Stop()

					return err
				},
			},
			{
				Name:        "build",
				Aliases:     []string{"install", "compile"},
				Usage:       "build <name>",
				Description: "Build application to binary",
				Action: func(cCtx *cli.Context) error {
					name := cCtx.Args().First()
					if name == "" {
						fmt.Println("Usage: bima build <name>")

						return nil
					}

					progress := spinner.New(spinner.CharSets[spinerIndex], duration)
					progress.Suffix = " Bundling application... "
					progress.Start()
					if err := tool.Call("clean"); err != nil {
						progress.Stop()
						color.New(color.FgRed).Println("Error cleaning dependencies")

						return err
					}

					if err := tool.Call("dump"); err != nil {
						progress.Stop()
						color.New(color.FgRed).Println("Error update DI container")

						return err
					}

					err := tool.Call("build", name, false)
					progress.Stop()

					return err
				},
			},
			{
				Name:        "update",
				Aliases:     []string{"upd"},
				Usage:       "update",
				Description: "Update project dependencies",
				Action: func(*cli.Context) error {
					progress := spinner.New(spinner.CharSets[spinerIndex], duration)
					progress.Suffix = " Updating dependencies... "
					progress.Start()
					if err := tool.Call("update"); err != nil {
						progress.Stop()
						color.New(color.FgRed).Println("Error update dependencies")

						return err
					}

					if err := tool.Call("dump"); err != nil {
						progress.Stop()
						color.New(color.FgRed).Println("Error update DI container")

						return err
					}

					progress.Stop()

					return nil
				},
			},
			{
				Name:        "clean",
				Aliases:     []string{"cln"},
				Usage:       "clean",
				Description: "Cleaning project dependencies",
				Action: func(*cli.Context) error {
					progress := spinner.New(spinner.CharSets[spinerIndex], duration)
					progress.Suffix = " Cleaning dependencies... "
					progress.Start()
					if err := tool.Call("clean"); err != nil {
						progress.Stop()
						color.New(color.FgRed).Println("Error cleaning dependencies")

						return err
					}

					if err := tool.Call("dump"); err != nil {
						progress.Stop()
						color.New(color.FgRed).Println("Error update DI container")

						return err
					}

					progress.Stop()

					return nil
				},
			},
			{
				Name:        "generate",
				Aliases:     []string{"gen", "genproto"},
				Usage:       "generate",
				Description: "Generate code from protobuf file(s)",
				Action: func(*cli.Context) error {
					progress := spinner.New(spinner.CharSets[spinerIndex], duration)
					progress.Suffix = " Generating protobuff... "
					progress.Start()
					if err := tool.Call("genproto"); err != nil {
						progress.Stop()
						color.New(color.FgRed).Println("Error generate protobuff")

						return err
					}

					if err := tool.Call("clean"); err != nil {
						progress.Stop()
						color.New(color.FgRed).Println("Error cleaning dependencies")

						return err
					}

					if err := tool.Call("dump"); err != nil {
						progress.Stop()
						color.New(color.FgRed).Println("Error update DI container")

						return err
					}

					progress.Stop()

					return nil
				},
			},
			{
				Name: "run",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "config",
						Aliases:     []string{"c"},
						Value:       ".env",
						Usage:       "Config file",
						Destination: &file,
					},
				},
				Aliases:     []string{"rn"},
				Usage:       "run <mode> [-c <config>]",
				Description: "Run application using <config> file",
				Action: func(cCtx *cli.Context) error {
					mode := cCtx.Args().First()
					if mode == "debug" {
						progress := spinner.New(spinner.CharSets[spinerIndex], duration)
						progress.Suffix = " Preparing debug mode... "
						progress.Start()

						err := tool.Call("build", "bima", true)
						if err != nil {
							progress.Stop()

							return err
						}

						progress.Stop()

						cmd, _ := syntax.NewParser().Parse(strings.NewReader(fmt.Sprintf("./bima run %s", file)), "")
						runner, _ := interp.New(interp.Env(nil), interp.StdIO(nil, os.Stdout, os.Stdout))

						return runner.Run(context.TODO(), cmd)
					}

					return tool.Call("run", file)
				},
			},
			{
				Name:        "debug",
				Aliases:     []string{"dbg"},
				Usage:       "debug",
				Description: "Debug application",
				Action: func(cCtx *cli.Context) error {
					content, err := os.ReadFile(".pid")
					if err != nil {
						color.New(color.FgRed).Println("Application not running")

						return nil
					}

					pid, err := strconv.Atoi(string(content))
					if err != nil {
						color.New(color.FgRed).Println("Invalid PID")

						return nil
					}

					return tool.Call("debug", pid)
				},
			},
			{
				Name:        "version",
				Aliases:     []string{"v"},
				Usage:       "version",
				Description: "Show Bima Cli version and Framework",
				Action: func(*cli.Context) error {
					wd, _ := os.Getwd()
					var path strings.Builder

					path.WriteString(wd)
					path.WriteString("/go.mod")

					framework := "unknown"
					mod, err := os.ReadFile(path.String())
					if err != nil {
						fmt.Printf("Framework: %s\n", framework)
						fmt.Printf("Cli: %s\n", version)

						return nil
					}

					f, err := modfile.Parse(path.String(), mod, nil)
					if err != nil {
						fmt.Printf("Framework: %s\n", framework)
						fmt.Printf("Cli: %s\n", version)

						return nil
					}

					for _, v := range f.Require {
						if v.Mod.Path == "github.com/bimalabs/framework/v4" {
							framework = v.Mod.Version

							break
						}
					}

					fmt.Printf("Framework: %s\n", framework)
					fmt.Printf("Cli: %s\n", version)

					return nil
				},
			},
			{
				Name:        "upgrade",
				Aliases:     []string{"upg"},
				Usage:       "upgrade",
				Description: "Upgrade Bima Cli to latest version",
				Action: func(*cli.Context) error {
					return tool.Call("upgrade", version)
				},
			},
			{
				Name:        "makesure",
				Aliases:     []string{"mks"},
				Usage:       "makesure",
				Description: "Check and install toolchain",
				Action: func(ctx *cli.Context) error {
					return tool.Call("makesure", protocMinVersion, protocGoMinVersion, protocGRpcMinVersion)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
