# Bimalabs Cli

## Requirements

- Go 1.16 or above

- Protoc 3.19.0 or above

- Protoc Gen Go 1.28.0 or above

- Protoc Gen gRpc 1.2.0 or above

- [gRPC Gateway Toolchain](https://github.com/grpc-ecosystem/grpc-gateway)

- [Delve](https://github.com/go-delve/delve/tree/master/Documentation/installation) for debug

## Install

- Download latest release from `https://github.com/bimalabs/cli/tags`

- Update dependencies using `go mod tidy`

- Extract and build using `go build -o bima`

- Move to your bin folder `mv bima $GOPATH/bin/bima`

- Checking toolchain installment `bima makesure`

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

- `bima run <mode> [-f <config>]` to run application

- `bima debug <pid>` to debug application

- `bima build` to build application

- `bima version` to show framework and cli version

- `bima upgrade` to upgrade cli version

- `bima makesure` to install toolchain

## Enable autocomplete terminal

To enable autocomplete feature, refer to [Urfave Cli](https://cli.urfave.org/v2/#enabling)
