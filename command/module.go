package command

import (
	"fmt"

	"github.com/bimalabs/cli/tool"
	"github.com/urfave/cli/v2"
)

func ModuleCommand(file string) *cli.Command {
	return &cli.Command{
		Name:        "module",
		Aliases:     []string{"mod"},
		Usage:       "Create or remove module",
		Description: "module <command>",
		Subcommands: []*cli.Command{moduleAdd(file), removeModule()},
	}
}

func moduleAdd(file string) *cli.Command {
	return &cli.Command{
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
		Description: "module add <name> [-c <config>]",
		Usage:       "Create new module <name> use <config> file",
		Action: func(ctx *cli.Context) error {
			name := ctx.Args().First()
			if name == "" {
				fmt.Println("Usage: bima module add <name> [-c <config>]")

				return nil
			}

			return tool.Module(name).Create(file)
		},
	}
}

func removeModule() *cli.Command {
	return &cli.Command{
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
	}
}
