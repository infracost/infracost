package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

var (
	//go:embed embed/*
	assets embed.FS

	allowedProviders = map[string]struct{}{
		"aws":    {},
		"azure":  {},
		"google": {},
	}

	embedPathExp  = regexp.MustCompile(`^embed/`)
	tmplSuffixExp = regexp.MustCompile(`\.tmpl$`)

	renderDir = "internal"
)

func main() {
	var c config
	flag.StringVar(&c.CloudProvider, "cloud-provider", "aws", "Cloud provider to create resource for, one of [aws, azure, google]")
	flag.StringVar(&c.Filename, "resource-name", "", "The resource name to generate, use underscores between names, e.g. autoscaling_group (required)")
	flag.BoolVar(&c.WithHelp, "with-help", false, "Generate your resources with doc blocks and examples to help you get started. Useful for understanding how to add a resource")
	flag.Parse()

	if flag.NFlag() == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if c.Filename == "" {
		exitWithErr(errors.New("Resource name cannot be blank"))
	}

	if _, ok := allowedProviders[c.CloudProvider]; !ok {
		exitWithErr(fmt.Errorf("[%s] is an invalid provider, please use one of [aws, azure, google]", c.CloudProvider))
	}

	c.Filename = strings.ToLower(c.Filename)
	c.ResourceName = toCamel(c.Filename)

	assetMap, err := initAssets()
	if err != nil {
		exitWithErr(fmt.Errorf("Error reading embed template dir:\n%w", err))
	}

	written, err := writeFiles(assetMap, c)
	if err != nil {
		exitWithErr(fmt.Errorf("Error generating files for resource:\n%w", err))
	}

	err = addResourceToRegistry(c)
	if err != nil {
		exitWithErr(fmt.Errorf("Error could not add resource to registry:\n%w", err))
	}

	writeOutput(c, written)
}

type config struct {
	CloudProvider string
	ResourceName  string
	Filename      string

	WithHelp bool
}

func (c config) FullResourceName() string {
	prefix := c.CloudProvider
	if c.CloudProvider == "azure" {
		prefix = "azurerm"
	}

	return fmt.Sprintf("%s_%s", prefix, c.Filename)
}

func (c config) RFuncName() string {
	return "new" + c.ResourceName
}

func (c config) RegistryFuncName() string {
	return "get" + c.ResourceName + "RegistryItem"
}

func (c config) ProviderDirectory() string {
	return path.Join("internal/providers/terraform/", c.CloudProvider)
}

func (c config) ResourceURL() string {
	switch c.CloudProvider {
	case "aws":
		return "https://aws.amazon.com"
	case "azure":
		return "https://azure.microsoft.com"
	case "google":
		return "https://cloud.google.com"
	}

	return "https://example.com"
}

func (c config) RunTestCommand() string {
	return strings.Join([]string{
		fmt.Sprintf("ARGS=\"--run Test%s -v -update\"", c.ResourceName),
		fmt.Sprintf("make test_%s", c.CloudProvider),
	}, " ")
}

func (c config) registryLocation() string {
	return path.Join(c.ProviderDirectory(), "registry.go")
}

func (c config) testDataLocation() string {
	return path.Join(c.ProviderDirectory(), "testdata", c.Filename+"_test")
}

type rep struct {
	exp   *regexp.Regexp
	value string
}

func (c config) toRegexpLookup() []rep {
	return []rep{
		{
			exp:   regexp.MustCompile(`\$filename\$`),
			value: strings.ToLower(c.Filename),
		},
		{
			exp:   regexp.MustCompile(`\$resource_name\$`),
			value: strings.ToLower(c.ResourceName),
		},
		{
			exp:   regexp.MustCompile(`\$cloud_provider\$`),
			value: strings.ToLower(c.CloudProvider),
		},
	}
}

func exitWithErr(err error) {
	fmt.Fprint(os.Stderr, err.Error()+"\n")
	os.Exit(1)
}

func writeOutput(c config, written []string) {
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("\nSuccessfully generated resource %s, files written:\n\n", c.FullResourceName()))
	for i, s := range written {
		if i == len(written)-1 {
			b.WriteString("\t" + s + "\n\n")
			break
		}

		b.WriteString("\t" + s + "\n")
	}

	b.WriteString(fmt.Sprintf("Added function %s to resource registry:\n\n", c.RegistryFuncName()))
	b.WriteString(fmt.Sprintf("\t%s\n\n", c.registryLocation()))

	b.WriteString(strings.Join(
		[]string{
			"Start by adding an example resource to the Terraform test file:",
			fmt.Sprintf("\t%s", path.Join(c.testDataLocation(), c.Filename+"_test.tf")),
			"and running the following command to generate initial Infracost output:",
			fmt.Sprintf("\t%s", c.RunTestCommand()),
			"Check out 'Adding new resources' section in our CONTRIBUTING.md guide for next steps!",
			"Happy hacking!!\n",
		}, "\n\n"),
	)

	fmt.Println(b.String())
}

func addResourceToRegistry(c config) error {
	f, err := decorator.ParseFile(token.NewFileSet(), c.registryLocation(), nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("Could not parse registry file %s for decorating %w", c.registryLocation(), err)
	}

	for _, decl := range f.Decls {
		if v, ok := decl.(*dst.GenDecl); ok {
			for _, spec := range v.Specs {
				if vs, ok := spec.(*dst.ValueSpec); ok && vs.Names[0].Name == "ResourceRegistry" {
					cl := vs.Values[0].(*dst.CompositeLit)
					e := &dst.CallExpr{
						Fun: &dst.Ident{
							Name: c.RegistryFuncName(),
						},
					}

					e.Decorations().Before = dst.NewLine
					e.Decorations().After = dst.NewLine
					cl.Elts = append(cl.Elts, e)
				}
			}
		}
	}

	ff, err := os.OpenFile(c.registryLocation(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Could not open registry file %s for writing %w", c.registryLocation(), err)
	}

	err = decorator.Fprint(ff, f)
	if err != nil {
		return fmt.Errorf("Could not write function %s to registry file %s %w", c.RegistryFuncName(), c.registryLocation(), err)
	}

	return nil
}

func writeFiles(assetMap map[string]*template.Template, c config) ([]string, error) {
	replacements := c.toRegexpLookup()
	made := make([]string, 0, len(assetMap))

	for p, tmpl := range assetMap {
		p = embedPathExp.ReplaceAllString(p, "")
		fileLoc := path.Join(renderDir, p)
		for _, repl := range replacements {
			fileLoc = repl.exp.ReplaceAllString(fileLoc, repl.value)
		}

		pieces := strings.Split(fileLoc, "/")
		dirLoc := strings.Join(pieces[:len(pieces)-1], "/")

		err := os.MkdirAll(dirLoc, os.ModePerm)
		if err != nil {
			return made, fmt.Errorf("Could not create directory %s %w", dirLoc, err)
		}

		sanitised := tmplSuffixExp.ReplaceAllString(fileLoc, "")
		file, err := os.Create(sanitised)
		if err != nil {
			return made, fmt.Errorf("Could not create file %s %w", fileLoc, err)
		}
		made = append(made, sanitised)

		err = tmpl.Execute(file, c)
		if err != nil {
			return made, fmt.Errorf("Could not execute template for file %s %w", fileLoc, err)
		}

		file.Close()
	}

	sort.Strings(made)
	return made, nil
}

func toCamel(snakeCase string) string {
	var camelCase string
	var isToUpper bool

	for k, v := range snakeCase {
		if k == 0 {
			camelCase = strings.ToUpper(string(snakeCase[0]))
			continue
		}

		if isToUpper {
			camelCase += strings.ToUpper(string(v))
			isToUpper = false
			continue
		}

		if v == '_' {
			isToUpper = true
			continue
		}

		camelCase += string(v)
	}

	return camelCase
}

func initAssets() (map[string]*template.Template, error) {
	templates := make(map[string]*template.Template)

	err := fs.WalkDir(assets, ".", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("Error walking resource generator embedded assets %w", err)
		}

		if !info.IsDir() && strings.HasSuffix(path, ".tmpl") {
			t, err := template.ParseFS(assets, path)
			if err != nil {
				return fmt.Errorf("Error parsing template for file %s %w", path, err)
			}

			templates[path] = t
		}

		return nil
	})

	return templates, err
}
