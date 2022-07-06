# Bimalabs Cli

## Install

- `git clone github.com/bimalabs/cli.git`

- `cd cli && go build -o bima`

- `sudo mv bima /usr/local/bin/bima`

## Usage

- `bima create app <name>` to create new application

- `bima module add <name>` to add new module

- `bima module remove <name>` to remove module

- `bima dump` to generate dic codes

- `bima update` to update framework and dependencies

- `bima clean` to cleaning dependencies

- `bima generate` to generating code from protobuff

- `bima run` to running application

- `bima version` to show framework and cli version

## TODO

- Add command to generate `middleware`, `route`, `pagination adapter` and `driver`

- Direct output from skeleton (show)

- Upgrade `bima` version
