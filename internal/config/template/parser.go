package template

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	pathToRegexp "github.com/soongo/path-to-regexp"
)

var (
	defaultInfracostTmplName = "infracost.yml.tmpl"
)

// Parser is the representation of an initialized Infracost template parser.
// It exposes custom template functions to the user which can act on Parser.projectDir
// or in isolation.
type Parser struct {
	projectDir string
	template   *template.Template
}

// NewParser returns a safely initialized Infracost template parser, this builds the underlying template with the
// Parser functions and sets the underlying default template name.
func NewParser(projectDir string) *Parser {
	p := Parser{projectDir: projectDir}
	t := template.New(defaultInfracostTmplName).Funcs(template.FuncMap{
		"base":       p.base,
		"stem":       p.stem,
		"ext":        p.ext,
		"startsWith": p.startsWith,
		"endsWith":   p.endsWith,
		"contains":   p.contains,
		"pathExists": p.pathExists,
		"matchPaths": p.matchPaths,
	})
	p.template = t

	return &p
}

// Compile writes an Infracost config to io.Writer wr using the provided template path.
func (p *Parser) Compile(templatePath string, wr io.Writer) error {
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

	err = t.Execute(wr, nil)
	if err != nil {
		return fmt.Errorf("could not execute template: %s err: %w", templatePath, err)
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
	if base == "." {
		base = p.projectDir
	}

	if !filepath.IsAbs(base) {
		base = filepath.Join(p.projectDir, base)
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
	_ = filepath.WalkDir(p.projectDir, func(path string, d fs.DirEntry, err error) error {
		rel, _ := filepath.Rel(p.projectDir, path)
		res, _ := match(rel)
		if res != nil {
			params := make(map[interface{}]interface{}, len(res.Params)+2)
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
