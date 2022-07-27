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
	"strconv"
	"strings"
	"time"

	"github.com/bimalabs/framework/v4/configs"
	"github.com/bimalabs/generators"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/gertd/go-pluralize"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/joho/godotenv"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
	"mvdan.cc/sh/interp"
	"mvdan.cc/sh/syntax"
)

type (
	command string
	util    string
)

func Pid() int {
	content, err := os.ReadFile(".pid")
	if err != nil {
		return 0
	}

	pid, err := strconv.Atoi(string(content))
	if err != nil {
		return 0
	}

	return pid
}

func Debug(ctx context.Context, pid int) error {
	cmd, _ := syntax.NewParser().Parse(strings.NewReader(fmt.Sprintf("dlv attach %d --listen=:16517 --headless --api-version=2 --log", pid)), "")
	runner, _ := interp.New(interp.Env(nil), interp.StdIO(nil, os.Stdout, os.Stdout))

	return runner.Run(ctx, cmd)
}

func Call(name string, args ...interface{}) error {
	in := make([]reflect.Value, len(args))
	for k, v := range args {
		in[k] = reflect.ValueOf(v)
	}

	c := util(name)
	returns := reflect.ValueOf(c).MethodByName(cases.Title(language.English).String(string(c))).Call(in)
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
		output, err := exec.Command("go", "install", "github.com/go-delve/delve/cmd/dlv@latest").CombinedOutput()
		if err != nil {
			progress.Stop()
			color.New(color.FgRed).Println("Error install go debugger: ", output)

			return err
		}
	}

	protocVersion := 0
	output, err := exec.Command("protoc", "--version").CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Printf("%s is not installed\n", color.New(color.FgRed, color.Bold).Sprint("protoc"))

		return err
	}

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
	output, _ = exec.Command("protoc-gen-go", "--version").CombinedOutput()
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
	wd := fmt.Sprintf("%s/bima", temp)
	os.RemoveAll(wd)

	progress := spinner.New(spinner.CharSets[spinerIndex], duration)
	progress.Suffix = " Checking new update... "
	progress.Start()

	repository, err := git.PlainClone(wd, false, &git.CloneOptions{
		URL:   "https://github.com/bimalabs/cli.git",
		Depth: 1,
	})
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(err)

		return nil
	}

	var (
		latest string
		when   = time.Now().AddDate(-3, 0, 0)
	)

	tags, err := repository.TagObjects()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(err)

		return nil
	}

	_ = tags.ForEach(func(t *object.Tag) error {
		if when.Before(t.Tagger.When) {
			when = t.Tagger.When
			latest = t.Name
		}

		return nil
	})

	if latest == version {
		progress.Stop()
		color.New(color.FgGreen).Println("Bima cli is already up to date")

		return nil
	}

	progress.Stop()

	progress = spinner.New(spinner.CharSets[spinerIndex], duration)
	progress.Suffix = " Updating Bima cli... "
	progress.Start()

	cmd := exec.Command("git", "fetch", "origin", fmt.Sprintf("refs/tags/%s", latest))
	cmd.Dir = wd
	err = cmd.Run()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Printf("Error fetch tag %s\n", latest)

		return nil
	}

	cmd = exec.Command("git", "checkout", latest)
	cmd.Dir = wd
	err = cmd.Run()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println("Error checkout to latest tag")

		return nil
	}

	cmd = exec.Command("go", "get")
	cmd.Dir = wd
	_ = cmd.Run()

	cmd = exec.Command("go", "build", "-o", "bima")
	cmd.Dir = wd
	output, err := cmd.CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))

		return err
	}

	binPath := os.Getenv("GOBIN")
	if binPath == "" {
		binPath = fmt.Sprintf("%s/bin", os.Getenv("GOPATH"))
	}

	if binPath == "" {
		output, err := exec.Command("which", "go").CombinedOutput()
		if err != nil {
			color.New(color.FgRed).Println(string(output))

			return err
		}

		binPath = strings.TrimSuffix(filepath.Dir(string(output)), "/")
	}

	cmd = exec.Command("mv", "bima", fmt.Sprintf("%s/bima", binPath))
	cmd.Dir = wd
	output, err = cmd.CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))

		return err
	}

	progress.Stop()
	color.New(color.FgGreen).Print("Bima cli is upgraded to ")
	color.New(color.FgGreen, color.Bold).Println(latest)

	os.RemoveAll(wd)

	return nil
}

func (u util) Build(name string, debug bool) error {
	if debug {
		return command("go build -race -gcflags \"all=-N -l\" -o %s cmd/main.go").run(name)
	}

	return command("go build -o %s cmd/main.go").run(name)
}

func (u util) Dump() error {
	return command("go run dumper/main.go").run()
}

func (u util) Kill() error {
	pid := Pid()
	if pid == 0 {
		return nil
	}

	err := exec.Command("kill", "-9", strconv.Itoa(pid)).Run()
	if err == nil {
		os.Remove(".pid")
	}

	return nil
}

func (u util) Clean() error {
	return command("go mod tidy").run()
}

func (u util) Update() error {
	return command("go get -u").run()
}

func (u util) Run(file string) error {
	defer u.Kill()

	return command("go run -race cmd/main.go run %s").run(file)
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
		_ = godotenv.Load()
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

func NewGenerator(driver string, apiVersion string) *generators.Factory {
	return &generators.Factory{
		Driver:     driver,
		ApiVersion: apiVersion,
		Pluralizer: pluralize.NewClient(),
		Template:   &generators.Template{},
		Generators: []generators.Generator{
			&generators.Dic{},
			&generators.Model{},
			&generators.Module{},
			&generators.Proto{},
			&generators.Provider{},
			&generators.Server{},
			&generators.Swagger{},
		},
	}
}
