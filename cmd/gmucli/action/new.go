package action

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"

	color "github.com/logrusorgru/aurora"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xlab/treeprint"

	"github.com/jllopis/gmu/cmd/gmucli/conf"
)

var (
	newCmd = &cobra.Command{
		Use:   "new",
		Short: "create a new gRPC project",
		Run:   newCmdRun,
	}

	tmplFuncMap = template.FuncMap{
		"ToLower": strings.ToLower,
		"ToUpper": strings.ToUpper,
	}

	packageName   string
	path          string
	projectUrl    string
	projectEmail  string
	serviceName   string
	apiVersion    string
	protocVersion string
)

func init() {
	RootCmd.AddCommand(newCmd)

	newCmd.Flags().StringVarP(&path, "path", "p", "", "path of the project. Defaults to current working directory")
	newCmd.Flags().StringVarP(&projectUrl, "project-url", "", "", "URL of the project")
	newCmd.Flags().StringVarP(&projectEmail, "project-email", "", "", "contact email for the project")
	newCmd.Flags().StringVarP(&serviceName, "service-name", "", "", "service name, defaults to the project name if not specified")
	newCmd.Flags().StringVarP(&apiVersion, "api-version", "", "v1", "api major version")
	newCmd.Flags().StringVarP(&protocVersion, "protoc-version", "", "3.7.0", "protobuf compiler version")
}

func newCmdRun(cmd *cobra.Command, args []string) {
	// config, err := conf.LoadConfig(cmd)
	err := conf.LoadConfig(cmd)
	if err != nil {
		fmt.Printf("gmu: failed to load configuation: %s\n", err.Error())
		os.Exit(1)
	}

	if len(args) != 1 {
		fmt.Printf("[%s] You must supply a path for the service, e.g gmu new my-project\n", color.Red("ERR"))
		return
	}

	basePath, err := getBasePath(viper.GetString("path"))
	if err != nil {
		fmt.Printf("can not set the base project path. Error: %s\n", err.Error())
		os.Exit(1)
	}

	// we need (modules) a packageName like {{repository}}/{{account}}/{{projectName}} like github.com/jllopis/gmu so we can
	// set go.mod and generated go files correctly.
	// If args[0] has no '/' separator, we assume the argument has not enough info so we exit alerting the user.
	if strings.Index(args[0], "/") == -1 {
		fmt.Printf("package name does not seem correct. You provided %s and a name in the form of 'github.com/jllopis/gmu' is needed\n\n", args[0])
		os.Exit(1)
	}
	projectName := strings.ToLower(args[0][strings.LastIndex(args[0], "/")+1:])
	packageName = args[0][:strings.LastIndex(args[0], "/")] + "/" + projectName
	projectRootPath := filepath.Join(basePath, projectName)
	if viper.GetString("service-name") == "" {
		viper.Set("service-name", strcase.ToCamel(projectName))
	}
	// fmt.Printf("projectName = %s\nserviceName = %s\n", projectName, serviceName)
	// os.Exit(0)

	checkDirMustNotExist(projectRootPath + "/" + projectName)

	// rd := &Directory{Name: strcase.ToSnake(projectName), Path: filepath.Join(projectRootPath, strcase.ToSnake(projectName))}
	rd := &Directory{Name: strings.ToLower(projectName), Path: projectRootPath}

	apid := rd.addDirectory("api")
	apid.addDirectory("proto").
		addDirectory(viper.GetString("api-version")).
		addFile(strings.ToLower(projectName)+".proto", "proto.tmpl", false)
	apid.addDirectory("swagger").
		addDirectory(viper.GetString("api-version"))

	cmdd := rd.addDirectory("cmd")
	cmdd.addDirectory("server").
		addFile("main.go", "cmd-server-main.go.tmpl", false)
	cmdd.addDirectory("client-grpc").
		addFile("main.go", "cmd-client-grpc-main.go.tmpl", false)
	cmdd.addDirectory("client-rest").
		addFile("main.go", "cmd-client-rest-main.go.tmpl", false)

	pkgd := rd.addDirectory("pkg")
	pkgd.addDirectory("api").
		addDirectory(viper.GetString("api-version"))
	pkgd.addDirectory("cmd").addFile("server.go", "pkg-cmd-server.go.tmpl", false)
	pkgd.addDirectory("logger").addFile("logger.go", "pkg-logger-logger.go.tmpl", false)
	pkgd.addDirectory("version").addFile("version.go", "pkg-version-version.go.tmpl", false)

	svcd := pkgd.addDirectory("service").addDirectory(viper.GetString("api-version"))
	svcd.addFile(strings.ToLower(viper.GetString("service-name"))+".go", "pkg-service-sample.go.tmpl", false)
	svcd.addFile(strings.ToLower(viper.GetString("service-name"))+"_test.go", "pkg-service-sample-test.go.tmpl", false)

	protocold := pkgd.addDirectory("protocol")
	pgrpcd := protocold.addDirectory("grpc")
	pgrpcd.addFile("server.go", "pkg-protocol-grpc-server.go.tmpl", false)
	pgrpcd.addFile("server_interceptors.go", "pkg-protocol-grpc-server-interceptors.go.tmpl", false)
	pgrpcd.addFile("server_options.go", "pkg-protocol-grpc-server-options.go.tmpl", false)

	gmidd := pgrpcd.addDirectory("middleware")
	gmidd.addFile("middleware.go", "pkg-protocol-grpc-middleware-middleware.go.tmpl", false)
	gmidd.addFile("logger.go", "pkg-protocol-grpc-middleware-logger.go.tmpl", false)
	gmidd.addFile("prometheus.go", "pkg-protocol-grpc-middleware-prometheus.go.tmpl", false)

	prestd := protocold.addDirectory("rest")
	prestd.addFile("server.go", "pkg-protocol-rest-server.go.tmpl", false)

	rmidd := prestd.addDirectory("middleware")
	rmidd.addFile("logger.go", "pkg-protocol-rest-middleware-logger.go.tmpl", false)
	rmidd.addFile("request-id.go", "pkg-protocol-rest-middleware-request-id.go.tmpl", false)

	scriptsd := rd.addDirectory("scripts")
	scriptsd.addFile("get-protoc", "getprotoc.tmpl", true)
	scriptsd.addFile("get-ext-protos", "getextprotos.tmpl", true)

	tpd := rd.addDirectory("third_party")
	tpd.addFile("protoc-gen.sh", "protocgen.tmpl", true)

	rd.addDirectory("tools")
	rd.addFile("go.mod", "go.mod.tmpl", false)
	rd.addFile(".gitignore", "gitignore.tmpl", false)
	rd.addFile("Makefile", "makefile.tmpl", false)
	rd.addFile("config.mk", "configmk.tmpl", false)
	rd.addFile("README.md", "readme.tmpl", false)
	rd.addFile("LICENCIA.md", "licencia.tmpl", false)
	rd.addFile("FAQ-Licencia.md", "faq-licencia.tmpl", false)

	project := &Project{
		PackageName:   packageName,
		ProjectName:   projectName,
		BasePath:      projectRootPath,
		RootDir:       rd,
		ServiceName:   viper.GetString("service-name"),
		ApiVersion:    viper.GetString("api-version"),
		ProjectUrl:    viper.GetString("project-url"), //projectUrl,
		ProjectEmail:  viper.GetString("project-email"),
		ProtocVersion: protocVersion,
	}

	fmt.Println("Project wil be created as:")
	fmt.Printf("\tPackage Name: %s\n\tService Name: %s\n\tProject Name: %s\n\tProject URL: %s\n\tProject Email: %s\n\tBase Path: %s\n\tAPI Version: %s\n\tProtoc Version: %s\n\n",
		color.Bold(project.PackageName),
		color.Bold(project.ServiceName),
		color.Bold(project.ProjectName),
		color.Bold(project.ProjectUrl),
		color.Bold(project.ProjectEmail),
		color.Bold(project.BasePath),
		color.Bold(project.ApiVersion),
		color.Bold(project.ProtocVersion),
	)
	fmt.Fprintf(os.Stdout, "Is this OK? [%s]es/[%s]o\n",
		color.Green("y"),
		color.Red("n"),
	)
	scan := bufio.NewScanner(os.Stdin)
	scan.Scan()
	if !strings.Contains(strings.ToLower(scan.Text()), "y") {
		os.Exit(0)
	}

	err = project.Flush()
	if err != nil {
		fmt.Printf("error creating project: %v\n", err)
		os.Exit(1)
	}

	// postActions will take the actions needed to finalize the setup like runnig
	// scripts to fill the directories, managing file permissions, or just showing some
	// indications to follow.
	project.doPostActions()

	// fmt.Println(rd)
	fmt.Printf("\n%s\n\n", color.Green("DONE!").Bold())
	fmt.Printf("Your project %s has been created at %s\n\n\n",
		color.Bold(project.ProjectName).Green(),
		color.Bold(project.BasePath).Green(),
	)
	fmt.Printf("Find some info about running the project in the %s file at the root of your project.\n\n", color.Bold("README.md"))
}

func getBasePath(path string) (string, error) {
	var cwd string
	var err error
	if cwd, err = os.Getwd(); err != nil {
		return "", err
	}

	if &path == nil || path == "" || path == "." {
		return cwd, nil
	}
	if strings.HasPrefix(path, "~") {
		// this it $HOME, so, expand
		path, err = homedir.Expand(path)
		if err != nil {
			return path, err
		}
	}
	if !strings.HasPrefix(path, "/") {
		fmt.Printf("found relative path %v\n", path)
		return cwd + "/" + path, nil
	}

	return path, nil
}

func checkDirMustNotExist(path string) {
	if _, err := os.Stat(path); err == nil {
		// dir exists
		fmt.Printf("Error: directory %s already exists. Can not override", path)
		os.Exit(1)
	}
}

type Project struct {
	PackageName   string
	ProjectName   string
	ProjectUrl    string
	ProjectEmail  string
	BasePath      string
	RootDir       *Directory
	ServiceName   string
	ApiVersion    string
	ProtocVersion string
}

type Directory struct {
	Name  string
	Path  string
	files []*File
	dirs  []*Directory
}

type File struct {
	Name     string
	Path     string
	Template string
	Exec     bool
}

func (p *Project) Flush() error {
	fmt.Printf("\t%s  %s\n", color.Green("create"), p.ProjectName)
	err := Mkdir(filepath.Join(p.BasePath))
	if err != nil {
		return err
	}
	return p.RootDir.flush(p)
}

func (p *Project) doPostActions() error {
	// run scripts/get-protoc
	if err := execute(p.BasePath, "scripts/get-protoc", "tools/protoc"); err != nil {
		fmt.Printf("error executing scripts/get-protoc, error: %s\n", err.Error())
	}
	// run scripts/get-ext-protos
	if err := execute(p.BasePath, "scripts/get-ext-protos"); err != nil {
		fmt.Printf("error executing scripts/get-ext-protos, error: %s\n", err.Error())
	}
	// run make to get protoc-gen-go
	// if err := execute(p.BasePath, "make", "tools/protoc-gen-go"); err != nil {
	if err := execute(p.BasePath, "go", "build", "-o", "tools/protoc-gen-go", "github.com/golang/protobuf/protoc-gen-go"); err != nil {
		fmt.Printf("error executing make tools/protoc-gen-go, error: %s\n", err.Error())
	}
	// run make to get protoc-gen-grpc-gateway
	if err := execute(p.BasePath, "go", "build", "-o", "tools/protoc-gen-grpc-gateway", "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway"); err != nil {
		fmt.Printf("error executing make tools/protoc-gen-grpc-gateway, error: %s\n", err.Error())
	}
	// run make to get grpc-gen-swagger
	if err := execute(p.BasePath, "go", "build", "-o", "tools/protoc-gen-swagger", "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger"); err != nil {
		fmt.Printf("error executing make tools/protoc-gen-swagger, error: %s\n", err.Error())
	}
	if err := execute(p.BasePath, "bash", "-c", "PATH="+p.BasePath+"/tools:$PATH protoc --proto_path=api/proto/"+p.ApiVersion+" --proto_path=third_party --go_out=plugins=grpc:pkg/api/"+p.ApiVersion+" "+p.ProjectName+".proto"); err != nil {
		fmt.Printf("error executing tools/protoc, error: %s\n", err.Error())
	}
	if err := execute(p.BasePath, "bash", "-c", "PATH="+p.BasePath+"/tools:$PATH protoc --proto_path=api/proto/"+p.ApiVersion+" --proto_path=third_party --grpc-gateway_out=logtostderr=true:pkg/api/"+p.ApiVersion+" "+p.ProjectName+".proto"); err != nil {
		fmt.Printf("error executing tools/protoc, error: %s\n", err.Error())
	}
	if err := execute(p.BasePath, "bash", "-c", "PATH="+p.BasePath+"/tools:$PATH protoc --proto_path=api/proto/"+p.ApiVersion+" --proto_path=third_party --swagger_out=logtostderr=true:api/swagger/"+p.ApiVersion+" "+p.ProjectName+".proto"); err != nil {
		fmt.Printf("error executing tools/protoc, error: %s\n", err.Error())
	}
	// create a git repo
	if err := execute(p.BasePath, "git", "init"); err != nil {
		fmt.Printf("error executing 'git init', error: %s\n", err.Error())
	}
	// and commit changes
	if err := execute(p.BasePath, "git", "add", "."); err != nil {
		fmt.Printf("error executing 'git add .', error: %s\n", err.Error())
	}
	if err := execute(p.BasePath, "git", "commit", "-qam", "'Initial import'"); err != nil {
		fmt.Printf("error executing git commit -qam 'Initial import', error: %s\n", err.Error())
	}
	return nil
}

func execute(path, command string, params ...string) error {
	var cmd *exec.Cmd
	if len(params) > 0 {
		fmt.Printf("\t%s  %s %v\n", color.Green("run"), command, params)
		cmd = exec.Command(command, params...)
	} else {
		fmt.Printf("\t%s  %s\n", color.Green("run"), command)
		cmd = exec.Command(command)
	}
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (d *Directory) flush(p *Project) error {
	for _, f := range d.files {
		fmt.Printf("\t%s  %s\n", color.Green("create"), strings.TrimPrefix(f.Path, p.BasePath+"/"))

		src, err := globalOptions.Box.FindString(f.Template)
		if err != nil {
			return err
		}
		t, err := template.New(f.Template).Funcs(tmplFuncMap).Parse(src)
		if err != nil {
			return err
		}

		file, err := os.Create(f.Path)
		if err != nil {
			return err
		}
		defer file.Close()
		err = t.Execute(file, p)
		if err != nil {
			return err
		}
		if f.Exec {
			if err := os.Chmod(f.Path, 0755); err != nil {
				fmt.Printf("Error: can not chmod file %s You must make it executable. Error: %s", f.Path, err.Error())
			}
		}
	}

	for _, g := range d.dirs {
		fmt.Printf("\t%s  %s\n", color.Green("create"), strings.TrimPrefix(g.Path, p.BasePath+"/"))
		err := Mkdir(g.Path)
		if err != nil {
			return err
		}

		err = g.flush(p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Directory) addDirectory(name string) *Directory {
	newD := &Directory{
		Name: name,
		Path: filepath.Join(d.Path, name),
	}
	d.dirs = append(d.dirs, newD)
	return newD
}

func (d *Directory) addFile(name, template string, exec bool) {
	d.files = append(d.files, &File{
		Name:     name,
		Template: template,
		Path:     filepath.Join(d.Path, name),
		Exec:     exec,
	})
}

func (d *Directory) String() string {
	t := d.tree(true, treeprint.New())
	return t.String()
}

func (f *Directory) tree(root bool, tree treeprint.Tree) treeprint.Tree {
	if !root {
		tree = tree.AddBranch(f.Name)
	}

	for _, v := range f.dirs {
		v.tree(false, tree)
	}

	for _, v := range f.files {
		tree.AddNode(v.Name)
	}

	return tree
}

func FileExists(dir string) bool {
	_, err := os.Stat(dir)
	return err == nil
}

func Mkdir(dir string) error {
	return os.MkdirAll(dir, os.ModePerm)
}
