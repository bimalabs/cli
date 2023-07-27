package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/bimalabs/cli/bima"
	"github.com/bimalabs/cli/tool"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"golang.org/x/mod/modfile"
)

func UpdateDependenciesCommand() *cli.Command {
	return &cli.Command{
		Name:        "update",
		Aliases:     []string{"upd"},
		Description: "update",
		Usage:       "Update project dependencies",
		Action: func(*cli.Context) error {
			progress := spinner.New(spinner.CharSets[bima.SpinerIndex], bima.Duration)
			progress.Suffix = " Updating dependencies... "
			progress.Start()
			if err := tool.Call("update"); err != nil {
				progress.Stop()
				color.New(color.FgRed).Println("Error updating dependencies")

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
	}
}

func CleanDependenciesCommand() *cli.Command {
	return &cli.Command{
		Name:        "clean",
		Aliases:     []string{"cln"},
		Description: "clean",
		Usage:       "Cleaning project dependencies",
		Action: func(*cli.Context) error {
			progress := spinner.New(spinner.CharSets[bima.SpinerIndex], bima.Duration)
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
	}
}

func GenerateProtobufCommand() *cli.Command {
	return &cli.Command{
		Name:        "generate",
		Aliases:     []string{"gen", "genproto"},
		Description: "generate",
		Usage:       "Generate code from protobuf file(s)",
		Action: func(*cli.Context) error {
			progress := spinner.New(spinner.CharSets[bima.SpinerIndex], bima.Duration)
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
	}
}

func CheckVersionCommand() *cli.Command {
	return &cli.Command{
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
				fmt.Printf("Cli: %s\n", bima.Version)

				return nil
			}

			f, err := modfile.Parse(path.String(), mod, nil)
			if err != nil {
				fmt.Printf("Framework: %s\n", framework)
				fmt.Printf("Cli: %s\n", bima.Version)

				return nil
			}

			for _, v := range f.Require {
				if v.Mod.Path == "github.com/bimalabs/framework/v4" {
					framework = v.Mod.Version

					break
				}
			}

			fmt.Printf("Framework: %s\n", framework)
			fmt.Printf("Cli: %s\n", bima.Version)
			fmt.Printf("SKeleton: %s\n", bima.SkeletonVersion)

			return nil
		},
	}
}

func MakesureToolchainInstalledCommand() *cli.Command {
	return &cli.Command{
		Name:        "makesure",
		Aliases:     []string{"mks"},
		Description: "makesure",
		Usage:       "Check and install toolchain when it possible",
		Action: func(ctx *cli.Context) error {
			return tool.Call("makesure", bima.ProtocMinVersion, bima.ProtocGoMinVersion, bima.ProtocGRpcMinVersion)
		},
	}
}

func UpgradeCliCommand() *cli.Command {
	return &cli.Command{
		Name:        "upgrade",
		Aliases:     []string{"upg"},
		Description: "upgrade",
		Usage:       "Upgrade cli to latest version",
		Action: func(*cli.Context) error {
			return tool.Call("upgrade", bima.Version)
		},
	}
}
