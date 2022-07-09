# Bimalabs Cli

## Install

- Download latest release from `https://github.com/bimalabs/cli/tags`

- Extract and build using `go build -o bima`

- Move to your bin folder `sudo mv bima /usr/local/bin/bima`

## Command List

- `bima create app <name>` to create new application

- `bima create middleware <name>` to create middleware under `middlewares` folder

- `bima create route <name>` to create route under `routes` folder

- `bima create driver <name>` to create database driver under `drivers` folder

- `bima create adapter <name>` to create pagination adapter under `adapters` folder

- `bima module add <name>` to add new module

- `bima module remove <name>` to remove module

- `bima dump` to generate service container codes

- `bima update` to update framework and dependencies

- `bima clean` to clean dependencies

- `bima generate` to generate code from protobuff

- `bima run` to run application

- `bima debug` to debug application (only support VS Code using [Debug Launcher Extension](https://marketplace.visualstudio.com/items?itemName=fabiospampinato.vscode-debug-launcher) and [Delve](https://github.com/go-delve/delve/tree/master/Documentation/installation))

- `bima build` to build application

- `bima version` to show framework and cli version

- `bima upgrade` to upgrade cli version
