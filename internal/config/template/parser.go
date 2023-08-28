package template

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	pathToRegexp "github.com/soongo/path-to-regexp"
	"gopkg.in/yaml.v2"
)

var (
	defaultInfracostTmplName = "infracost.yml.tmpl"
)

// Variables hold the global variables that are passed into any template that the Parser evaluates.
type Variables struct {
	Branch     string
	BaseBranch string
}

// Parser is the representation of an initialized Infracost template parser.
// It exposes custom template functions to the user which can act on Parser.repoDir
// or in isolation.
type Parser struct {
	repoDir   string
	template  *template.Template
	variables Variables
}

// NewParser returns a safely initialized Infracost template parser, this builds the underlying template with the
// Parser functions and sets the underlying default template name. Default variables can be passed to the parser which
// will be passed to the template on execution.
func NewParser(repoDir string, variables Variables) *Parser {
	absRepoDir, _ := filepath.Abs(repoDir)
	p := Parser{repoDir: absRepoDir, variables: variables}
	t := template.New(defaultInfracostTmplName).Funcs(template.FuncMap{
		"base":       p.base,
		"stem":       p.stem,
		"ext":        p.ext,
		"startsWith": p.startsWith,
		"endsWith":   p.endsWith,
		"contains":   p.contains,
		"pathExists": p.pathExists,
		"matchPaths": p.matchPaths,
		"list":       p.list,
		"relPath":    p.relPath,
		"isDir":      p.isDir,
		"readFile":   p.readFile,
		"parseJson":  p.parseJson,
		"parseYaml":  p.parseYaml,
	})
	p.template = t

	return &p
}

// CompileFromFile writes an Infracost config to io.Writer wr using the provided template path.
func (p *Parser) CompileFromFile(templatePath string, wr io.Writer) error {
	baseTemplate := p.template

	// if the template name is not `infracost.yml.tmpl` we need to change the template base
	// name to match the template name, otherwise executing the template will fail. This is
	// done by calling .New which copies over all the configuration from the base template
	// to a new one.
	base := filepath.Base(templatePath)
	if base != defaultInfracostTmplName {
		baseTemplate = baseTemplate.New(base)
	}

	t, err := baseTemplate.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("could not parse template path: %s err: %w", templatePath, err)
	}

	err = t.Execute(wr, p.variables)
	if err != nil {
		return fmt.Errorf("could not execute template: %s err: %w", templatePath, err)
	}

	return nil
}

// Compile writes an Infracost config to io.Writer wr using the provided template string.
func (p *Parser) Compile(template string, wr io.Writer) error {
	// rewrite escaped values to their literal values so that we get a nice output.
	template = strings.ReplaceAll(template, `\n`, "\n")
	template = strings.ReplaceAll(template, `\r`, "\r")
	template = strings.ReplaceAll(template, `\t`, "\t")

	t, err := p.template.Parse(template)
	if err != nil {
		return fmt.Errorf("could not parse template: %q err: %w", template, err)
	}

	err = t.Execute(wr, p.variables)
	if err != nil {
		return fmt.Errorf("could not execute template: %q err: %w", template, err)
	}

	return nil
}

// base returns the last element of path.
func (p *Parser) base(path string) string {
	return filepath.Base(path)
}

// base returns the last element of path with the extension removed.
func (p *Parser) stem(path string) string {
	return strings.TrimSuffix(p.base(path), p.ext(path))
}

// ext returns the file name extension used by path.
func (p *Parser) ext(path string) string {
	return filepath.Ext(path)
}

// startsWith tests whether the string s begins with prefix.
func (p *Parser) startsWith(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// endsWith tests whether the string s ends with suffix.
func (p *Parser) endsWith(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// contains reports whether substr is within s.
func (p *Parser) contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// pathExists reports whether path is a subpath within base.
func (p *Parser) pathExists(base, path string) bool {
	if !filepath.IsAbs(base) {
		base = filepath.Join(p.repoDir, base)
	}

	// Ensure the base path is within the repo directory
	baseAbs, _ := filepath.Abs(base)
	repoDirAbs, _ := filepath.Abs(p.repoDir)
	// Add a file separator at the end to ensure we don't match a directory that starts with the same prefix
	// e.g. `/path/to/infracost` shouldn't match `/path/to/infra`.
	if !strings.HasPrefix(fmt.Sprintf("%s%s", baseAbs, string(filepath.Separator)), fmt.Sprintf("%s%s", repoDirAbs, string(filepath.Separator))) {
		return false
	}

	var fileExists bool
	_ = filepath.WalkDir(base, func(subpath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(base, subpath)
		if rel == path {
			fileExists = true

			// exit the WalkDir tree evaluation early as we've found the file we're looking for
			return errors.New("file found")
		}

		return nil
	})

	return fileExists
}

// matchPaths returns a list of matches that in the project directory tree that match the pattern.
// Results are returned using a map as keys are variable.
// Each result returned also contains two special keys:
//
//		_path - the full path of that the pattern matched on
//		_dir  - the base directory that the pattern matched on
//
//	With an example tree of:
//		├── environment
//		│     ├── dev
//		│     │     └── terraform.tfvars
//		│     └── prod
//		│         └── terraform.tfvars
//		├── infracost.yml.tmpl
//		└── main.tf
//
//	Using a pattern of:
//
//		"environment/:env/terraform.tfvars"
//
//	Returned results would be:
//
//		- { _path: environment/dev/terraform.tfvars, _dir: environment/dev, env: dev }
//		- { _path: environment/prod/terraform.tfvars, _dir: environment/prod, env: prod }
func (p *Parser) matchPaths(pattern string) []map[interface{}]interface{} {
	match := pathToRegexp.MustMatch(pattern, nil)

	var matches []map[interface{}]interface{}
	_ = filepath.WalkDir(p.repoDir, func(path string, d fs.DirEntry, err error) error {
		rel, _ := filepath.Rel(p.repoDir, path)
		res, _ := match(rel)
		if res != nil {
			var out map[string]interface{}
			params := make(map[interface{}]interface{})

			b, _ := jsoniter.Marshal(p.variables)
			_ = jsoniter.Unmarshal(b, &out)

			for k, v := range out {
				params[k] = v
			}
			for k, v := range res.Params {
				params[k] = v
			}

			params["_path"] = rel
			dir := filepath.Dir(rel)
			params["_dir"] = dir

			matches = append(matches, params)
		}

		return nil
	})

	return matches
}

// list is a useful function for creating an arbitrary array of values which can be
// looped over in a template. For example:
//
//	$my_list = list "foo" "bar"
//	{{- range $my_list }}
//		{{ . }}
//	{{- end }}
func (p *Parser) list(v ...interface{}) []interface{} {
	return v
}

// relPath returns a relative path that is lexically equivalent to targpath when
// joined to basepath with an intervening separator. If there is an error returning the
// relative path we panic so that the error is show when executing the template.
func (p *Parser) relPath(basepath string, tarpath string) string {
	rel, err := filepath.Rel(basepath, tarpath)
	if err != nil {
		panic(err)
	}

	return rel
}

// isDir returns is path points to a directory.
func (p *Parser) isDir(path string) bool {
	fullPath := filepath.Join(p.repoDir, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// readFile reads the named file and returns the contents.
func (p *Parser) readFile(path string) string {
	if !isSubdirectory(p.repoDir, path) {
		panic(fmt.Sprintf("%q must be within the repository root %q", path, filepath.Base(p.repoDir)))
	}

	fullPath := filepath.Join(p.repoDir, path)
	b, err := os.ReadFile(fullPath)
	if err != nil {
		panic(err)
	}

	return string(b)
}

// parseYaml decodes provided yaml contents and assigns decoded values into a
// generic out value. This can be used as a simple object in the templates.
func (p *Parser) parseYaml(contents string) map[string]interface{} {
	var out map[string]interface{}
	err := yaml.Unmarshal([]byte(contents), &out)
	if err != nil {
		panic(err)
	}

	return out
}

// parseJson decodes the provided json contents and assigns decoded values into a
// generic out value. This can be used as a simple object in the templates.
func (p *Parser) parseJson(contents string) map[string]interface{} {
	var out map[string]interface{}
	err := jsoniter.Unmarshal([]byte(contents), &out)
	if err != nil {
		panic(err)
	}

	return out
}

func isSubdirectory(base, target string) bool {
	full := filepath.Join(base, target)
	fileInfo, err := os.Lstat(full)
	if err != nil {
		return false
	}

	absBasePath, err := filepath.Abs(base)
	if err != nil {
		return false
	}

	absTargetPath, err := filepath.Abs(full)
	if err != nil {
		return false
	}

	relPath, err := filepath.Rel(absBasePath, absTargetPath)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(relPath, "..") && fileInfo.Mode()&os.ModeSymlink == 0
}
