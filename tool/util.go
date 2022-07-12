package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/bimalabs/framework/v4/configs"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
	"mvdan.cc/sh/interp"
	"mvdan.cc/sh/syntax"
)

type (
	command string
	util    string
)

func Call(name string, args ...interface{}) error {

	in := make([]reflect.Value, len(args))
	for k, v := range args {
		in[k] = reflect.ValueOf(v)
	}

	if len(in) == 0 {

	}

	c := util(name)
	returns := reflect.ValueOf(c).MethodByName(strings.Title(string(c))).Call(in)
	if len(returns) > 1 {
		return nil
	}

	v := returns[0]
	err, ok := v.Interface().(error)
	if !ok {
		return nil
	}

	return err
}

func (c command) run(args ...interface{}) error {
	var f string
	if len(args) == 0 {
		f = fmt.Sprint(string(c))
	} else {
		f = fmt.Sprintf(string(c), args...)
	}
	cmd, _ := syntax.NewParser().Parse(strings.NewReader(f), "")
	runner, _ := interp.New(interp.Env(nil), interp.StdIO(nil, os.Stdout, os.Stdout))

	return runner.Run(context.TODO(), cmd)
}

func (u util) Debug(pid int) error {
	return command("dlv attach %d --listen=:16517 --headless --api-version=2 --log").run(pid)
}

func (u util) Build(name string, debug bool) error {
	if debug {
		return command("go build -gcflags \"all=-N -l\" -o %s cmd/main.go").run(name)
	}

	return command("go build -o %s cmd/main.go").run(name)
}

func (u util) Dump() error {
	return command("go run dumper/main.go").run()
}

func (u util) Clean() error {
	return command("go mod tidy").run()
}

func (u util) toolchain() error {
	return command(`go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc
    `).run()
}

func (u util) Update() error {
	return command("go get -u").run()
}

func (u util) Run(file string) error {
	return command("go run cmd/main.go run %s").run(file)
}

func (u util) Genproto() error {
	return command("sh proto_gen.sh").run()
}

func config(config *configs.Env, filePath string, ext string) {
	switch ext {
	case ".env":
		godotenv.Load()
		parse(config)
	case ".yaml":
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalln(err.Error())
		}

		err = yaml.Unmarshal(content, config)
		if err != nil {
			log.Fatalln(err.Error())
		}
	case ".json":
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalln(err.Error())
		}

		err = json.Unmarshal(content, config)
		if err != nil {
			log.Fatalln(err.Error())
		}
	}
}

func parse(config *configs.Env) {
	config.Secret = os.Getenv("APP_SECRET")
	config.Debug, _ = strconv.ParseBool(os.Getenv("APP_DEBUG"))
	config.HttpPort, _ = strconv.Atoi(os.Getenv("APP_PORT"))
	config.RpcPort, _ = strconv.Atoi(os.Getenv("GRPC_PORT"))
	config.Service = os.Getenv("APP_NAME")

	dbPort, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	config.Db = configs.Db{
		Host:     os.Getenv("DB_HOST"),
		Port:     dbPort,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		Driver:   os.Getenv("DB_DRIVER"),
	}

	config.CacheLifetime, _ = strconv.Atoi(os.Getenv("CACHE_LIFETIME"))
}
