package tool

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bimalabs/cli/bima"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	env = `APP_DEBUG=true
APP_PORT=7777
GRPC_PORT=1717
APP_NAME=%s
APP_SECRET=%s
`

	adapter = `package adapters

import (
    "context"

    "github.com/bimalabs/framework/v4/paginations"
    "github.com/vcraescu/go-paginator/v2"
)

type %s struct {
}

func (a *%s) CreateAdapter(ctx context.Context, paginator paginations.Pagination) paginator.Adapter {
    // TODO

    return nil
}
`

	driver = `package drivers

import (
    "gorm.io/gorm"
)

type %s string

func (_ %s) Connect(host string, port int, user string, password string, dbname string, debug bool) *gorm.DB {
    // TODO

    return nil
}

func (m %s) Name() string {
    return string(m)
}
`

	route = `package routes

import (
    "net/http"

    "github.com/bimalabs/framework/v4/middlewares"
    "google.golang.org/grpc"
)

type %s struct {
}

func (r *%s) Path() string {
    return "/%s"
}

func (r *%s) Method() string {
    return http.MethodGet
}

func (r *%s) SetClient(client *grpc.ClientConn) {
    // TODO
}

func (r *%s) Middlewares() []middlewares.Middleware {
    // TODO

    return nil
}

func (r *%s) Handle(response http.ResponseWriter, request *http.Request, params map[string]string) {
    // TODO
}
`

	middleware = `package middlewares

import (
    "net/http"
)

type %s struct {
}

func (m *%s) Attach(request *http.Request, response http.ResponseWriter) bool {
    // TODO

    return false
}

func (m *%s) Priority() int {
    return 0
}
`
)

type (
	App        string
	Middleware string
	Driver     string
	Adapter    string
	Route      string
)

func (a App) Create() error {
	wd, _ := os.Getwd()
	if _, err := os.Stat(fmt.Sprintf("%s/%s", wd, string(a))); !os.IsNotExist(err) {
		return errors.New("Project already exits")
	}

	err := createApp(string(a))
	if err == nil {
		fmt.Printf("%s application created\n", color.New(color.FgGreen).Sprint(cases.Title(language.English).String(string(a))))

		util := color.New(color.Bold)

		fmt.Print("Move to ")
		util.Print(string(a))
		fmt.Print(" folder and type ")
		util.Println("bima run")
	}

	return err
}

func (m Middleware) Create() error {
	progress := spinner.New(spinner.CharSets[bima.SpinerIndex], bima.Duration)
	progress.Suffix = " Creating middleware... "
	progress.Start()
	time.Sleep(1 * time.Second)

	wd, err := os.Getwd()
	if err != nil {
		progress.Stop()

		return err
	}

	err = os.MkdirAll(fmt.Sprintf("%s/middlewares", wd), 0755)
	if err != nil {
		progress.Stop()

		return err
	}

	f, err := os.Create(fmt.Sprintf("%s/middlewares/%s.go", wd, strings.ToLower(string(m))))
	if err != nil {
		progress.Stop()

		return err
	}

	name := cases.Title(language.English).String(string(m))
	_, err = f.WriteString(fmt.Sprintf(middleware, name, name, name))
	if err != nil {
		progress.Stop()

		return err
	}

	_ = f.Sync()
	_ = f.Close()

	if err := Call("clean"); err != nil {
		progress.Stop()
		color.New(color.FgRed).Println("Error cleaning dependencies")

		return err
	}

	progress.Stop()
	fmt.Printf("Middleware %s created\n", color.New(color.FgGreen).Sprint(name))

	return nil
}

func (d Driver) Create() error {
	progress := spinner.New(spinner.CharSets[bima.SpinerIndex], bima.Duration)
	progress.Suffix = " Creating database driver... "
	progress.Start()
	time.Sleep(1 * time.Second)

	wd, err := os.Getwd()
	if err != nil {
		progress.Stop()

		return err
	}

	err = os.MkdirAll(fmt.Sprintf("%s/drivers", wd), 0755)
	if err != nil {
		progress.Stop()

		return err
	}

	f, err := os.Create(fmt.Sprintf("%s/drivers/%s.go", wd, strings.ToLower(string(d))))
	if err != nil {
		progress.Stop()

		return err
	}

	name := cases.Title(language.English).String(string(d))
	_, err = f.WriteString(fmt.Sprintf(driver, name, name, name))
	if err != nil {
		progress.Stop()

		return err
	}

	_ = f.Sync()
	_ = f.Close()

	if err := Call("clean"); err != nil {
		progress.Stop()
		color.New(color.FgRed).Println("Error cleaning dependencies")

		return err
	}

	progress.Stop()
	fmt.Printf("Driver %s created\n", color.New(color.FgGreen).Sprint(name))

	return nil
}

func (a Adapter) Create() error {
	progress := spinner.New(spinner.CharSets[bima.SpinerIndex], bima.Duration)
	progress.Suffix = " Creating pagination adapter... "
	progress.Start()
	time.Sleep(1 * time.Second)

	wd, err := os.Getwd()
	if err != nil {
		progress.Stop()

		return err
	}

	err = os.MkdirAll(fmt.Sprintf("%s/adapters", wd), 0755)
	if err != nil {
		progress.Stop()

		return err
	}

	f, err := os.Create(fmt.Sprintf("%s/adapters/%s.go", wd, strings.ToLower(string(a))))
	if err != nil {
		progress.Stop()

		return err
	}

	name := cases.Title(language.English).String(string(a))
	_, err = f.WriteString(fmt.Sprintf(adapter, name, name))
	if err != nil {
		progress.Stop()

		return err
	}

	_ = f.Sync()
	_ = f.Close()

	if err := Call("clean"); err != nil {
		progress.Stop()

		color.New(color.FgRed).Println("Error cleaning dependencies")

		return err
	}

	progress.Stop()
	fmt.Printf("Adapter %s created\n", color.New(color.FgGreen).Sprint(name))

	return nil
}

func (r Route) Create() error {
	progress := spinner.New(spinner.CharSets[bima.SpinerIndex], bima.Duration)
	progress.Suffix = " Creating route placeholder... "
	progress.Start()
	time.Sleep(1 * time.Second)

	wd, err := os.Getwd()
	if err != nil {
		progress.Stop()

		return err
	}

	err = os.MkdirAll(fmt.Sprintf("%s/routes", wd), 0755)
	if err != nil {
		progress.Stop()

		return err
	}

	lName := strings.ToLower(string(r))
	f, err := os.Create(fmt.Sprintf("%s/routes/%s.go", wd, lName))
	if err != nil {
		progress.Stop()

		return err
	}

	name := cases.Title(language.English).String(string(r))
	_, err = f.WriteString(fmt.Sprintf(route, name, name, lName, name, name, name, name))
	if err != nil {
		progress.Stop()

		return err
	}

	_ = f.Sync()
	_ = f.Close()

	if err := Call("clean"); err != nil {
		progress.Stop()

		color.New(color.FgRed).Println("Error cleaning dependencies")

		return err
	}

	progress.Stop()
	fmt.Printf("Route %s created\n", color.New(color.FgGreen).Sprint(name))

	return nil
}

func createApp(name string) error {
	progress := spinner.New(spinner.CharSets[bima.SpinerIndex], bima.Duration)
	progress.Suffix = fmt.Sprintf(" Creating %s project... ", color.New(color.FgGreen).Sprint(name))
	progress.Start()

	output, err := exec.Command("git", "clone", "--depth", "1", "https://github.com/bimalabs/skeleton.git", name).CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))

		return err
	}

	wd, _ := os.Getwd()
	dir := fmt.Sprintf("%s/%s", wd, name)
	cmd := exec.Command("git", "fetch", "origin", fmt.Sprintf("refs/tags/%s", bima.SkeletonVersion))
	cmd.Dir = dir
	output, err = cmd.CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))
		os.RemoveAll(dir)

		return err
	}

	cmd = exec.Command("git", "checkout", bima.SkeletonVersion)
	cmd.Dir = dir
	output, err = cmd.CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))
		os.RemoveAll(dir)

		return err
	}

	output, err = exec.Command("rm", "-rf", fmt.Sprintf("%s/.git", name)).CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))
		os.RemoveAll(dir)

		return err
	}

	f, err := os.Create(fmt.Sprintf("%s/.env", name))
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))
		os.RemoveAll(dir)

		return err
	}

	hasher := sha256.New()
	hasher.Write([]byte(time.Now().Format(time.RFC3339)))

	_, err = f.WriteString(fmt.Sprintf(env, name, base64.URLEncoding.EncodeToString(hasher.Sum(nil))))
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))
		os.RemoveAll(dir)

		return err
	}

	_ = f.Sync()
	_ = f.Close()

	progress.Stop()

	progress = spinner.New(spinner.CharSets[bima.SpinerIndex], bima.Duration)
	progress.Suffix = " Download dependencies... "
	progress.Start()

	cmd = exec.Command("go", "mod", "download")
	cmd.Dir = fmt.Sprintf("%s/%s", wd, name)
	_ = cmd.Run()

	cmd = exec.Command("go", "run", "dumper/main.go")
	cmd.Dir = dir
	output, err = cmd.CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))
		os.RemoveAll(dir)

		return err
	}

	progress = spinner.New(spinner.CharSets[bima.SpinerIndex], bima.Duration)
	progress.Suffix = " Cleaning project... "
	progress.Start()

	cmd = exec.Command("go", "get")
	cmd.Dir = dir
	output, err = cmd.CombinedOutput()
	if err != nil {
		progress.Stop()
		color.New(color.FgRed).Println(string(output))
		os.RemoveAll(dir)

		return err
	}

	progress.Stop()

	return nil
}
