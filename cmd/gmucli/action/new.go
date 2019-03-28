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
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
)

var (
	newCmd = &cobra.Command{
		Use:   "new",
		Short: "create a new gRPC project",
		Run:   newCmdRun,
	}

	tmplFuncMap = template.FuncMap{
		"ToLower": strings.ToLower,
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
	if len(args) != 1 {
		fmt.Printf("[%s] You must supply a path for the service, e.g gmu new my-project\n", color.Red("ERR"))
		return
	}

	basePath, err := getBasePath(path)
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

	checkDirMustNotExist(projectRootPath + "/" + projectName)

	// rd := &Directory{Name: strcase.ToSnake(projectName), Path: filepath.Join(projectRootPath, strcase.ToSnake(projectName))}
	rd := &Directory{Name: strings.ToLower(projectName), Path: projectRootPath}

	apid := rd.addDirectory("api")
	apid.addDirectory("proto").
		addDirectory(apiVersion).
		addFile(strings.ToLower(projectName)+".proto", "proto.tmpl", false)
	apid.addDirectory("swagger").
		addDirectory(apiVersion)

	cmdd := rd.addDirectory("cmd")
	cmdd.addDirectory("server").
		addFile("main.go", "cmd-server-main.go.tmpl", false)
	cmdd.addDirectory("client-grpc").
		addFile("main.go", "cmd-client-grpc-main.go.tmpl", false)
	cmdd.addDirectory("client-rest").
		addFile("main.go", "cmd-client-rest-main.go.tmpl", false)

	pkgd := rd.addDirectory("pkg")
	pkgd.addDirectory("api").
		addDirectory(apiVersion)
	pkgd.addDirectory("cmd").addFile("server.go", "pkg-cmd-server.go.tmpl", false)
	pkgd.addDirectory("logger").addFile("logger.go", "logger.go.tmpl", false)

	svcd := pkgd.addDirectory("service").addDirectory(apiVersion)
	svcd.addFile(serviceName+".go", "pkg-service-sample.go.tmpl", false)
	svcd.addFile(serviceName+"_test.go", "pkg-service-sample-test.go.tmpl", false)

	protocold := pkgd.addDirectory("protocol")
	pgrpcd := protocold.addDirectory("grpc")
	pgrpcd.addFile("server.go", "pkg-protocol-grpc-server.go.tmpl", false)
	gmidd := pgrpcd.addDirectory("middleware")
	gmidd.addFile("logger.go", "pkg-protocol-grpc-middleware-logger.go.tmpl", false)
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
		ApiVersion:    apiVersion,
		ProjectUrl:    projectUrl,
		ProjectEmail:  projectEmail,
		ProtocVersion: protocVersion,
	}

	if serviceName == "" {
		serviceName = projectName
	}
	project.ServiceName = strcase.ToCamel(serviceName)

	fmt.Println("Project wil be created as:")
	fmt.Printf("\tPackage Name: %s\n\tProject Name: %s\n\tProject URL: %s\n\tProject Email: %s\n\tBase Path: %s\n\tAPI Version: %s\n\tProtoc Version: %s\n\tRoot Dir: %s\n\n",
		color.Bold(project.PackageName),
		color.Bold(project.ProjectName),
		color.Bold(project.ProjectUrl),
		color.Bold(project.ProjectEmail),
		color.Bold(project.BasePath),
		color.Bold(project.ApiVersion),
		color.Bold(project.ProtocVersion),
		project.RootDir,
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
	fmt.Println("Done!")
}

func getBasePath(path string) (string, error) {
	var projectPath, cwd string
	var err error
	if cwd, err = os.Getwd(); err != nil {
		return "", err
	}

	if &path == nil || path == "" || path == "." {
		projectPath = cwd
	} else {
		if !strings.HasPrefix(path, "/") {
			fmt.Printf("found relative path %v\n", path)
			projectPath = cwd + "/" + path
		}
	}
	return projectPath, nil
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
	if err := execute(p.BasePath, "make", "tools/protoc-gen-go"); err != nil {
		fmt.Printf("error executing make tools/protoc-gen-go, error: %s\n", err.Error())
	}
	// run make to get protoc-gen-grpc-gateway
	if err := execute(p.BasePath, "make", "tools/protoc-gen-grpc-gateway"); err != nil {
		fmt.Printf("error executing make tools/protoc-gen-grpc-gateway, error: %s\n", err.Error())
	}
	// run make to get grpc-gen-swagger
	if err := execute(p.BasePath, "make", "tools/protoc-gen-swagger"); err != nil {
		fmt.Printf("error executing make tools/protoc-gen-swagger, error: %s\n", err.Error())
	}
	// run third_party protoc-gen
	if err := execute(p.BasePath, "make", "proto"); err != nil {
		fmt.Printf("error executing third_party/protoc-gen.sh, error: %s\n", err.Error())
	}
	return nil
}

func execute(path, command string, params ...string) error {
	var cmd *exec.Cmd
	if len(params) > 0 {
		fmt.Printf("\t%s  %s %v\n", color.Cyan("run"), command, params)
		cmd = exec.Command(command, params...)
	} else {
		fmt.Printf("\t%s  %s\n", color.Cyan("run"), command)
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
