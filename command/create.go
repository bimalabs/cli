package command

import (
	"fmt"

	"github.com/bimalabs/cli/tool"
	"github.com/urfave/cli/v2"
)

func CreateCommand() *cli.Command {
	return &cli.Command{
		Name:        "create",
		Aliases:     []string{"new"},
		Usage:       "Create something with bima",
		Description: "bima create <command>",
		Subcommands: []*cli.Command{createPackage(), createMiddleware(), createRoute(), createAdapter(), createDriver()},
	}
}

func createPackage() *cli.Command {
	return &cli.Command{
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
	}
}

func createMiddleware() *cli.Command {
	return &cli.Command{
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
	}
}

func createDriver() *cli.Command {
	return &cli.Command{
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
	}
}

func createAdapter() *cli.Command {
	return &cli.Command{
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
	}
}

func createRoute() *cli.Command {
	return &cli.Command{
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
	}
}
