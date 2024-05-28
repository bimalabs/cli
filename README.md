# Bimalabs Cli

Bimalabs Cli is command line to create, run and build [bimalabs framework](https://github.com/bimalabs/framework), the all in one solution for developing backend app in few minutes.

## Requirements

- Go 1.16 or above

- Protoc 3.19.0 or above

- Protoc Gen Go 1.28.0 or above

- Protoc Gen gRpc 1.2.0 or above

- [gRPC Gateway Toolchain](https://github.com/grpc-ecosystem/grpc-gateway)

- [Delve](https://github.com/go-delve/delve/tree/master/Documentation/installation) for debug

## Install

- Checking `$GOPATH` in your environment variable using `echo $GOPATH`

- Download latest release from `https://github.com/bimalabs/cli/tags`

- Update dependencies using `go mod tidy`

- Extract and build using `go build -o bima-cli`

- Move to your bin folder `mv bima-cli $GOPATH/bin/bima`

- Checking toolchain installment `bima makesure`

## Command List

- `bima create app <name>` to create new application

- `bima create middleware <name>` to create middleware under `middlewares` folder

- `bima create route <name>` to create route under `routes` folder

- `bima create driver <name>` to create database driver under `drivers` folder

- `bima create adapter <name>` to create pagination adapter under `adapters` folder

- `bima module add <name> [<version> -c <config>]` to add new module with `version` using `config` file

- `bima module remove <name>` to remove module

- `bima dump` to generate service container codes

- `bima update` to update framework and dependencies

- `bima clean` to clean dependencies

- `bima generate` to generate code from protobuff

- `bima run <mode> [-c <config>]` to run application on `mode` mode using `config` file

- `bima build` to build application

- `bima version` to show framework and cli version

- `bima upgrade` to upgrade cli version

- `bima makesure` to install toolchain

## Enable autocomplete terminal

To enable autocomplete feature, refer to [Urfave Cli](https://cli.urfave.org/v2/examples/bash-completions)
