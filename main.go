package main

import (
	"context"
	"errors"
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
	version              = "v1.2.16"
	protocMinVersion     = 31900
	protocGoMinVersion   = 12800
	protocGRpcMinVersion = 10200
	spinerIndex          = 9
	duration             = 77 * time.Millisecond
)

func main() {
	file := ""
	app := &cli.App{
		Name:                 "Bima cli",
		Usage:                "Bima Framework Toolkit",
		Description:          "bima version",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:        "create",
				Aliases:     []string{"new"},
				Usage:       "Create something with bima",
				Description: "bima create <command>",
				Subcommands: []*cli.Command{
					{
						Name:        "project",
						Aliases:     []string{"app"},
						Description: "bima create app <name>",
						Usage:       "Create new application or project",
						Action: func(ctx *cli.Context) error {
							name := ctx.Args().First()
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
						Description: "bima create middleware <name>",
						Usage:       "Create new middleware",
						Action: func(ctx *cli.Context) error {
							name := ctx.Args().First()
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
						Description: "bima create driver <name>",
						Usage:       "Create new driver",
						Action: func(ctx *cli.Context) error {
							name := ctx.Args().First()
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
						Description: "bima create adapter <name>",
						Usage:       "Create new adapter",
						Action: func(ctx *cli.Context) error {
							name := ctx.Args().First()
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
						Description: "bima create route <name>",
						Usage:       "Create new route",
						Action: func(ctx *cli.Context) error {
							name := ctx.Args().First()
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
				Name:        "module",
				Aliases:     []string{"mod"},
				Usage:       "Create or remove module",
				Description: "module <command>",
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
						Description: "module add <name> [<version> -c <config>]",
						Usage:       "Create new module <name> with <version> using <config> file",
						Action: func(ctx *cli.Context) error {
							name := ctx.Args().First()
							if name == "" {
								fmt.Println("Usage: bima module add <name> [<version> -c <config>]")

								return nil
							}

							version := "v1"
							if ctx.NArg() > 1 {
								version = ctx.Args().Get(1)
							}

							return tool.Module(name).Create(file, version)
						},
					},
					{
						Name:        "remove",
						Aliases:     []string{"rm", "rem"},
						Description: "module remove <name>",
						Usage:       "Remove module <name>",
						Action: func(ctx *cli.Context) error {
							name := ctx.Args().First()
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
				Description: "dump",
				Usage:       "Dump service container",
				Action: func(*cli.Context) error {
					progress := spinner.New(spinner.CharSets[spinerIndex], duration)
					progress.Suffix = " Dumping service container... "
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
				Description: "build <name>",
				Usage:       "Build application to binary",
				Action: func(ctx *cli.Context) error {
					name := ctx.Args().First()
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
						color.New(color.FgRed).Println("Error updating services container")

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
				Description: "update",
				Usage:       "Update project dependencies",
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
						color.New(color.FgRed).Println("Error updating services container")

						return err
					}

					progress.Stop()

					return nil
				},
			},
			{
				Name:        "clean",
				Aliases:     []string{"cln"},
				Description: "clean",
				Usage:       "Cleaning project dependencies",
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
						color.New(color.FgRed).Println("Error updating services container")

						return err
					}

					progress.Stop()

					return nil
				},
			},
			{
				Name:        "generate",
				Aliases:     []string{"gen", "genproto"},
				Description: "generate",
				Usage:       "Generate code from protobuf file(s)",
				Action: func(*cli.Context) error {
					progress := spinner.New(spinner.CharSets[spinerIndex], duration)
					progress.Suffix = " Generating codes from protobuff file(s)... "
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
						color.New(color.FgRed).Println("Error updating services container")

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
				Description: "run <mode> [-c <config>]",
				Usage:       "Run application using <config> file",
				Action: func(ctx *cli.Context) error {
					mode := ctx.Args().First()
					if mode == "debug" {
						progress := spinner.New(spinner.CharSets[spinerIndex], duration)
						progress.Suffix = " Preparing debug mode... "
						progress.Start()

						os.Remove(".pid")

						err := tool.Call("build", "bima", true)
						if err != nil {
							progress.Stop()

							return err
						}

						progress.Stop()

						cmd, _ := syntax.NewParser().Parse(strings.NewReader(fmt.Sprintf("./bima run %s true", file)), "")
						runner, _ := interp.New(interp.Env(nil), interp.StdIO(nil, os.Stdout, os.Stdout))

						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()
						go func() {
							runner.Run(ctx, cmd)
						}()

						var pid = 0
						for {
							if pid != 0 {
								break
							}

							content, err := os.ReadFile(".pid")
							if err != nil {
								time.Sleep(50 * time.Millisecond)

								continue
							}

							pid, err = strconv.Atoi(string(content))
							if err != nil {
								color.New(color.FgRed).Println("Invalid PID")

								break
							}
						}

						if pid == 0 {
							return errors.New("PID not exists")
						}

						return tool.Debug(ctx, pid)
					}

					progress := spinner.New(spinner.CharSets[spinerIndex], duration)
					progress.Suffix = " Preparing run mode... "
					progress.Start()
					if err := tool.Call("dump"); err != nil {
						progress.Stop()
						color.New(color.FgRed).Println("Error updating services container")

						return err
					}

					progress.Stop()

					return tool.Call("run", file)
				},
			},
			{
				Name:        "version",
				Aliases:     []string{"v"},
				Description: "version",
				Usage:       "Show cli and framework version",
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
				Description: "upgrade",
				Usage:       "Upgrade cli to latest version",
				Action: func(*cli.Context) error {
					return tool.Call("upgrade", version)
				},
			},
			{
				Name:        "makesure",
				Aliases:     []string{"mks"},
				Description: "makesure",
				Usage:       "Check and install toolchain when it possible",
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
