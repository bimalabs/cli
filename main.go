package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	Version              = "v1.1.17"
	ProtocMinVersion     = 31900
	ProtocGoMinVersion   = 12800
	ProtocGRpcMinVersion = 10200
	SpinerIndex          = 9
	Duration             = 77 * time.Millisecond
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
						Name:    "project",
						Aliases: []string{"app"},
						Usage:   "bima create app <name>",
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
						Name:    "middleware",
						Aliases: []string{"mid"},
						Usage:   "bima create middleware <name>",
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
						Name:    "driver",
						Aliases: []string{"dvr"},
						Usage:   "bima create driver <name>",
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
						Name:    "adapter",
						Aliases: []string{"adp"},
						Usage:   "bima create adapter <name>",
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
						Name:    "route",
						Aliases: []string{"rt"},
						Usage:   "bima create route <name>",
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
						Aliases: []string{"new"},
						Usage:   "module add <name>",
						Action: func(cCtx *cli.Context) error {
							name := cCtx.Args().First()
							if name == "" {
								fmt.Println("Usage: bima module add <name> [<version>]")

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
						Name:    "remove",
						Aliases: []string{"rm", "rem"},
						Usage:   "module remove <name>",
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
				Name:    "dump",
				Aliases: []string{"dmp"},
				Usage:   "dump",
				Action: func(*cli.Context) error {
					progress := spinner.New(spinner.CharSets[SpinerIndex], Duration)
					progress.Suffix = " Generate service container... "
					progress.Start()
					time.Sleep(1 * time.Second)

					err := tool.Call("dump")
					progress.Stop()

					return err
				},
			},
			{
				Name:    "build",
				Aliases: []string{"install", "compile"},
				Usage:   "build <name>",
				Action: func(cCtx *cli.Context) error {
					name := cCtx.Args().First()
					if name == "" {
						fmt.Println("Usage: bima build <name>")

						return nil
					}

					progress := spinner.New(spinner.CharSets[SpinerIndex], Duration)
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
				Name:    "update",
				Aliases: []string{"upd"},
				Usage:   "update",
				Action: func(*cli.Context) error {
					progress := spinner.New(spinner.CharSets[SpinerIndex], Duration)
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
				Name:    "clean",
				Aliases: []string{"cln"},
				Usage:   "clean",
				Action: func(*cli.Context) error {
					progress := spinner.New(spinner.CharSets[SpinerIndex], Duration)
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
				Name:    "generate",
				Aliases: []string{"gen", "genproto"},
				Usage:   "generate",
				Action: func(*cli.Context) error {
					progress := spinner.New(spinner.CharSets[SpinerIndex], Duration)
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
				Aliases: []string{"rn"},
				Usage:   "run <mode> -f config.json",
				Action: func(cCtx *cli.Context) error {
					mode := cCtx.Args().First()
					if mode == "debug" {
						progress := spinner.New(spinner.CharSets[SpinerIndex], Duration)
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
				Name:    "debug",
				Aliases: []string{"dbg"},
				Usage:   "debug",
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
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "version",
				Action: func(*cli.Context) error {
					wd, _ := os.Getwd()
					var path strings.Builder

					path.WriteString(wd)
					path.WriteString("/go.mod")

					version := "unknown"
					mod, err := os.ReadFile(path.String())
					if err != nil {
						fmt.Printf("Framework: %s\n", version)
						fmt.Printf("Cli: %s\n", Version)

						return nil
					}

					f, err := modfile.Parse(path.String(), mod, nil)
					if err != nil {
						fmt.Printf("Framework: %s\n", version)
						fmt.Printf("Cli: %s\n", Version)

						return nil
					}

					for _, v := range f.Require {
						if v.Mod.Path == "github.com/bimalabs/framework/v4" {
							version = v.Mod.Version

							break
						}
					}

					fmt.Printf("Framework: %s\n", version)
					fmt.Printf("Cli: %s\n", Version)

					return nil
				},
			},
			{
				Name:    "upgrade",
				Aliases: []string{"upg"},
				Usage:   "upgrade",
				Action: func(*cli.Context) error {
					return upgrade()
				},
			},
			{
				Name:    "makesure",
				Aliases: []string{"mks"},
				Usage:   "makesure",
				Action: func(ctx *cli.Context) error {
					progress := spinner.New(spinner.CharSets[SpinerIndex], Duration)
					progress.Suffix = " Checking toolchain installment... "
					progress.Start()

					if err := tool.Call("clean"); err != nil {
						progress.Stop()
						color.New(color.FgRed).Println("Error cleaning dependencies")

						return err
					}

					protocVersion := 0
					output, err := exec.Command("protoc", "--version").CombinedOutput()
					vSlice := strings.Split(string(output), " ")
					if len(vSlice) > 1 {
						vSlice = strings.Split(vSlice[1], ".")
						if len(vSlice) > 2 {
							major, _ := strconv.Atoi(vSlice[0])
							minor, _ := strconv.Atoi(vSlice[1])
							fix, _ := strconv.Atoi(vSlice[2])
							protocVersion = (10_000 * major) + (100 * minor) + fix
						}
					}

					protocGoVersion := 0
					output, err = exec.Command("protoc-gen-go", "--version").CombinedOutput()
					vSlice = strings.Split(string(output), " ")
					if len(vSlice) > 1 {
						vSlice[1] = strings.TrimPrefix(vSlice[1], "v")
						vSlice = strings.Split(vSlice[1], ".")
						if len(vSlice) > 2 {
							major, _ := strconv.Atoi(vSlice[0])
							minor, _ := strconv.Atoi(vSlice[1])
							fix, _ := strconv.Atoi(vSlice[2])
							protocGoVersion = (10_000 * major) + (100 * minor) + fix
						}
					}

					protocGRpcVersion := 0
					output, err = exec.Command("protoc-gen-go-grpc", "--version").CombinedOutput()
					vSlice = strings.Split(string(output), " ")
					if len(vSlice) > 1 {
						vSlice = strings.Split(vSlice[1], ".")
						if len(vSlice) > 2 {
							major, _ := strconv.Atoi(vSlice[0])
							minor, _ := strconv.Atoi(vSlice[1])
							fix, _ := strconv.Atoi(vSlice[2])
							protocGRpcVersion = (10_000 * major) + (100 * minor) + fix
						}
					}

					if protocVersion >= ProtocMinVersion && protocGoVersion >= ProtocGoMinVersion && protocGRpcVersion >= ProtocGRpcMinVersion {
						progress.Stop()
						color.New(color.FgGreen).Println("Toolchain is already installed")

						return nil
					}

					progress.Stop()

					progress = spinner.New(spinner.CharSets[SpinerIndex], Duration)
					progress.Suffix = " Try to install/update to latest toolchain... "
					progress.Start()
					err = tool.Call("toolchain")
					if err != nil {
						progress.Stop()
						color.New(color.FgRed).Println("Error install toolchain")

						return err
					}

					progress.Stop()
					color.New(color.FgGreen).Println("Toolchain installed")

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func upgrade() error {
	temp := strings.TrimSuffix(os.TempDir(), "/")
	os.RemoveAll(fmt.Sprintf("%s/bima", temp))

	progress := spinner.New(spinner.CharSets[SpinerIndex], Duration)
	progress.Suffix = " Checking new update... "
	progress.Start()

	wd := fmt.Sprintf("%s/bima", temp)
	output, err := exec.Command("git", "clone", "--depth", "1", "https://github.com/bimalabs/cli.git", wd).CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))

		return nil
	}

	cmd := exec.Command("git", "rev-list", "--tags", "--max-count=1")
	cmd.Dir = wd
	output, err = cmd.CombinedOutput()

	re := regexp.MustCompile(`\r?\n`)
	commitId := re.ReplaceAllString(string(output), "")

	cmd = exec.Command("git", "describe", "--tags", commitId)
	cmd.Dir = wd
	output, err = cmd.CombinedOutput()

	re = regexp.MustCompile(`\r?\n`)
	latest := re.ReplaceAllString(string(output), "")
	if latest == Version {
		progress.Stop()
		color.New(color.FgGreen).Println("Bima Cli is already up to date")

		return nil
	}

	progress.Stop()

	progress = spinner.New(spinner.CharSets[SpinerIndex], Duration)
	progress.Suffix = " Updating Bima Cli... "
	progress.Start()

	cmd = exec.Command("git", "fetch")
	cmd.Dir = wd
	err = cmd.Run()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))

		return nil
	}

	cmd = exec.Command("git", "checkout", latest)
	cmd.Dir = wd
	err = cmd.Run()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))

		return nil
	}

	cmd = exec.Command("go", "get")
	cmd.Dir = wd
	cmd.Run()

	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = wd
	cmd.Run()

	cmd = exec.Command("go", "run", "dumper/main.go")
	cmd.Dir = wd
	output, err = cmd.CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))

		return err
	}

	cmd = exec.Command("go", "get", "-u")
	cmd.Dir = wd
	output, err = cmd.CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))

		return err
	}

	cmd = exec.Command("go", "build", "-o", "bima")
	cmd.Dir = wd
	output, err = cmd.CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))

		return err
	}

	binPath := os.Getenv("GOBIN")
	if binPath == "" {
		binPath = os.Getenv("GOPATH")
	}

	if binPath == "" {
		output, err := exec.Command("which", "go").CombinedOutput()
		if err != nil {
			color.New(color.FgRed).Println(string(output))

			return err
		}

		binPath = filepath.Dir(string(output))
	}

	cmd = exec.Command("mv", "bima", fmt.Sprintf("%s/bin/bima", binPath))
	cmd.Dir = wd
	output, err = cmd.CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))

		return err
	}

	progress.Stop()
	color.New(color.FgGreen).Print("Bima Cli is upgraded to ")
	color.New(color.FgGreen, color.Bold).Println(latest)

	return nil
}
