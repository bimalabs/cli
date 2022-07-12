package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/bimalabs/framework/v4/configs"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
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

func (u util) Makesure(protoc int, protocGo int, protocGRpc int) error {
	progress := spinner.New(spinner.CharSets[spinerIndex], duration)
	progress.Suffix = " Checking toolchain installment... "
	progress.Start()

	if err := u.Clean(); err != nil {
		progress.Stop()
		color.New(color.FgRed).Println("Error cleaning dependencies")

		return err
	}

	_, err := exec.LookPath("dlv")
	if err != nil {
		output, err := exec.Command("go install github.com/go-delve/delve/cmd/dlv@latest").CombinedOutput()
		if err != nil {
			progress.Stop()
			color.New(color.FgRed).Println("Error install go debugger: ", output)

			return err
		}
	}

	protocVersion := 0
	output, err := exec.Command("protoc", "--version").CombinedOutput()
	vSlice := strings.Split(string(output), " ")
	if len(vSlice) > 1 {
		vSlice = strings.Split(vSlice[1], ".")
		if len(vSlice) > 2 {
			major, _ := strconv.Atoi(vSlice[0])
			minor, _ := strconv.Atoi(vSlice[1])
			fix, _ := strconv.Atoi(vSlice[2])
			protocVersion = (10_000 * major) + (100 * minor) + fix
		}
	}

	protocGoVersion := 0
	output, err = exec.Command("protoc-gen-go", "--version").CombinedOutput()
	vSlice = strings.Split(string(output), " ")
	if len(vSlice) > 1 {
		vSlice[1] = strings.TrimPrefix(vSlice[1], "v")
		vSlice = strings.Split(vSlice[1], ".")
		if len(vSlice) > 2 {
			major, _ := strconv.Atoi(vSlice[0])
			minor, _ := strconv.Atoi(vSlice[1])
			fix, _ := strconv.Atoi(vSlice[2])
			protocGoVersion = (10_000 * major) + (100 * minor) + fix
		}
	}

	protocGRpcVersion := 0
	output, err = exec.Command("protoc-gen-go-grpc", "--version").CombinedOutput()
	vSlice = strings.Split(string(output), " ")
	if len(vSlice) > 1 {
		vSlice = strings.Split(vSlice[1], ".")
		if len(vSlice) > 2 {
			major, _ := strconv.Atoi(vSlice[0])
			minor, _ := strconv.Atoi(vSlice[1])
			fix, _ := strconv.Atoi(vSlice[2])
			protocGRpcVersion = (10_000 * major) + (100 * minor) + fix
		}
	}

	if protocVersion >= protoc && protocGoVersion >= protocGo && protocGRpcVersion >= protocGRpc {
		progress.Stop()
		color.New(color.FgGreen).Println("Toolchain is already installed")

		return nil
	}

	progress.Stop()

	progress = spinner.New(spinner.CharSets[spinerIndex], duration)
	progress.Suffix = " Try to install/update to latest toolchain... "
	progress.Start()
	err = u.toolchain()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println("Error installing toolchain")

		return err
	}

	progress.Stop()
	color.New(color.FgGreen).Println("Toolchain installed")

	return nil
}

func (u util) Upgrade(version string) error {
	temp := strings.TrimSuffix(os.TempDir(), "/")
	os.RemoveAll(fmt.Sprintf("%s/bima", temp))

	progress := spinner.New(spinner.CharSets[spinerIndex], duration)
	progress.Suffix = " Checking new update... "
	progress.Start()

	wd := fmt.Sprintf("%s/bima", temp)
	output, err := exec.Command("git", "clone", "--depth", "1", "https://github.com/bimalabs/cli.git", wd).CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))

		return nil
	}

	cmd := exec.Command("git", "rev-list", "--tags", "--max-count=1")
	cmd.Dir = wd
	output, err = cmd.CombinedOutput()

	re := regexp.MustCompile(`\r?\n`)
	commitId := re.ReplaceAllString(string(output), "")

	cmd = exec.Command("git", "describe", "--tags", commitId)
	cmd.Dir = wd
	output, err = cmd.CombinedOutput()

	re = regexp.MustCompile(`\r?\n`)
	latest := re.ReplaceAllString(string(output), "")
	if latest == version {
		progress.Stop()
		color.New(color.FgGreen).Println("Bima Cli is already up to date")

		return nil
	}

	progress.Stop()

	progress = spinner.New(spinner.CharSets[spinerIndex], duration)
	progress.Suffix = " Updating Bima Cli... "
	progress.Start()

	cmd = exec.Command("git", "fetch")
	cmd.Dir = wd
	err = cmd.Run()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))

		return nil
	}

	cmd = exec.Command("git", "checkout", latest)
	cmd.Dir = wd
	err = cmd.Run()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))

		return nil
	}

	cmd = exec.Command("go", "get")
	cmd.Dir = wd
	cmd.Run()

	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = wd
	cmd.Run()

	cmd = exec.Command("go", "run", "dumper/main.go")
	cmd.Dir = wd
	output, err = cmd.CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))

		return err
	}

	cmd = exec.Command("go", "get", "-u")
	cmd.Dir = wd
	output, err = cmd.CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))

		return err
	}

	cmd = exec.Command("go", "build", "-o", "bima")
	cmd.Dir = wd
	output, err = cmd.CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))

		return err
	}

	binPath := os.Getenv("GOBIN")
	if binPath == "" {
		binPath = os.Getenv("GOPATH")
	}

	if binPath == "" {
		output, err := exec.Command("which", "go").CombinedOutput()
		if err != nil {
			color.New(color.FgRed).Println(string(output))

			return err
		}

		binPath = filepath.Dir(string(output))
	}

	cmd = exec.Command("mv", "bima", fmt.Sprintf("%s/bin/bima", binPath))
	cmd.Dir = wd
	output, err = cmd.CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))

		return err
	}

	progress.Stop()
	color.New(color.FgGreen).Print("Bima Cli is upgraded to ")
	color.New(color.FgGreen, color.Bold).Println(latest)

	return nil
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

func (u util) Update() error {
	return command("go get -u").run()
}

func (u util) Run(file string) error {
	return command("go run cmd/main.go run %s").run(file)
}

func (u util) Genproto() error {
	return command("sh proto_gen.sh").run()
}

func (u util) toolchain() error {
	return command(`go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc
    `).run()
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
