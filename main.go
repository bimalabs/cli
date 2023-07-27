package main

import (
	"log"
	"os"

	"github.com/bimalabs/cli/command"
	"github.com/urfave/cli/v2"
)

func main() {
	file := ""
	app := &cli.App{
		Name:                 "bima",
		Usage:                "Bima Framework Toolkit",
		Description:          "bima version",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			command.CreateCommand(),
			command.ModuleCommand(file),
			command.BuildAppCommand(),
			command.RunAppCommand(file),
			command.DumpServiceContainerCommand(),
			command.UpdateDependenciesCommand(),
			command.CleanDependenciesCommand(),
			command.GenerateProtobufCommand(),
			command.MakesureToolchainInstalledCommand(),
			command.CheckVersionCommand(),
			command.UpgradeCliCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
