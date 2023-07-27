package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bimalabs/cli/bima"
	"github.com/bimalabs/cli/tool"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"mvdan.cc/sh/interp"
	"mvdan.cc/sh/syntax"
)

func BuildAppCommand() *cli.Command {
	return &cli.Command{
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

			progress := spinner.New(spinner.CharSets[bima.SpinerIndex], bima.Duration)
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
	}
}

func DumpServiceContainerCommand() *cli.Command {
	return &cli.Command{
		Name:        "dump",
		Aliases:     []string{"dmp"},
		Description: "dump",
		Usage:       "Dump service container",
		Action: func(*cli.Context) error {
			progress := spinner.New(spinner.CharSets[bima.SpinerIndex], bima.Duration)
			progress.Suffix = " Dumping service container... "
			progress.Start()
			time.Sleep(1 * time.Second)

			err := tool.Call("dump")
			progress.Stop()

			return err
		},
	}
}

func RunAppCommand(file string) *cli.Command {
	return &cli.Command{
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
			if tool.Pid() != 0 {
				_ = tool.Call("kill")
			}

			mode := ctx.Args().First()
			if mode == "debug" {
				progress := spinner.New(spinner.CharSets[bima.SpinerIndex], bima.Duration)
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

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				go func() {
					_ = runner.Run(ctx, cmd)
				}()

				var pid = 0
				for {
					if pid != 0 {
						break
					}

					pid = tool.Pid()
					if pid == 0 {
						time.Sleep(100 * time.Millisecond)

						continue
					}
				}

				if pid == 0 {
					return errors.New("PID not exists")
				}

				return tool.Debug(ctx, pid)
			}

			progress := spinner.New(spinner.CharSets[bima.SpinerIndex], bima.Duration)
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
	}
}
