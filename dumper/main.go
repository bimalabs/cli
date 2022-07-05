package main

import (
	"fmt"
	"os"

	"github.com/bimalabs/cli/provider"
	"github.com/sarulabs/dingo/v4"
)

func main() {
	err := dingo.GenerateContainerWithCustomPkgName((*provider.Generator)(nil), "generated", "generator")
	if err != nil {
		fmt.Println("Error dumping container: ", err.Error())
		os.Exit(1)
	}
}
