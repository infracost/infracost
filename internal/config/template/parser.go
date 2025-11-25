package template

import (
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	jsoniter "github.com/json-iterator/go"
	pathToRegexp "github.com/soongo/path-to-regexp"
	"gopkg.in/yaml.v2"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
)

var (
	defaultInfracostTmplName = "infracost.yml.tmpl"
)

type DetectedProject struct {
	Name              string
	Path              string
	TerraformVarFiles []string
	DependencyPaths   []string
	Env               string
}

type DetectedRooModule struct {
	Path     string
	Projects []DetectedProject
}

// Variables hold the global variables that are passed into any template that the Parser evaluates.
type Variables struct {
	RepoName            string
	Branch              string
	BaseBranch          string
	DetectedProjects    []DetectedProject
	DetectedRootModules []DetectedRooModule
}

// Parser is the representation of an initialized Infracost template parser.
// It exposes custom template functions to the user which can act on Parser.repoDir
// or in isolation.
type Parser struct {
	repoDir   string
	template  *template.Template
	variables Variables
	config    *config.Config
}

// NewParser returns a safely initialized Infracost template parser, this builds the underlying template with the
// Parser functions and sets the underlying default template name. Default variables can be passed to the parser which
// will be passed to the template on execution.
func NewParser(repoDir string, variables Variables, config *config.Config) *Parser {
	absRepoDir, _ := filepath.Abs(repoDir)
	p := Parser{repoDir: absRepoDir, variables: variables, config: config}
	t := template.New(defaultInfracostTmplName).Funcs(template.FuncMap{
		"base":         p.base,
		"stem":         p.stem,
		"ext":          p.ext,
		"lower":        p.lower,
		"startsWith":   p.startsWith,
		"endsWith":     p.endsWith,
		"contains":     p.contains,
		"splitList":    p.splitList,
		"join":         p.join,
		"trimPrefix":   p.trimPrefix,
		"trimSuffix":   p.trimSuffix,
		"replace":      p.replace,
		"quote":        p.quote,
		"squote":       p.squote,
		"pathExists":   p.pathExists,
		"matchPaths":   p.matchPaths,
		"list":         p.list,
		"relPath":      p.relPath,
		"isDir":        p.isDir,
		"readFile":     p.readFile,
		"parseJson":    p.parseJson,
		"parseYaml":    p.parseYaml,
		"isProduction": p.isProduction,
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

// lower returns a copy of the string s with all Unicode letters mapped to their lower case.
func (p *Parser) lower(s string) string {
	return strings.ToLower(s)
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

// splitList splits the string s into a slice of substrings separated by sep.
func (p *Parser) splitList(sep, s string) []string {
	return strings.Split(s, sep)
}

// join joins the list of strings l into a single string separated by sep.
func (p *Parser) join(sep string, l []string) string {
	return strings.Join(l, sep)
}

// trimPrefix returns s without the provided prefix string.
func (p *Parser) trimPrefix(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}

// trimSuffix returns s without the provided suffix string.
func (p *Parser) trimSuffix(s, suffix string) string {
	return strings.TrimSuffix(s, suffix)
}

// replace returns s with all instances of old replaced by new.
func (p *Parser) replace(old, new, s string) string {
	return strings.ReplaceAll(s, old, new)
}

// quote wraps the provided strings in double quotes.
// Taken from https://github.com/Masterminds/sprig/blob/581758eb7d96ae4d113649668fa96acc74d46e7f/strings.go#L83
func (p *Parser) quote(str ...any) string {
	out := make([]string, 0, len(str))
	for _, s := range str {
		if s != nil {
			out = append(out, fmt.Sprintf("%q", strval(s)))
		}
	}
	return strings.Join(out, " ")
}

// squote wraps the provided strings in single quotes.
// Taken from https://github.com/Masterminds/sprig/blob/581758eb7d96ae4d113649668fa96acc74d46e7f/strings.go#L93C1-L101C2
func (p *Parser) squote(str ...any) string {
	out := make([]string, 0, len(str))
	for _, s := range str {
		if s != nil {
			out = append(out, fmt.Sprintf("'%v'", s))
		}
	}
	return strings.Join(out, " ")
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

	targetPath := filepath.Join(base, path)
	if _, err := os.Stat(targetPath); err == nil {
		return true
	}

	return false
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
func (p *Parser) matchPaths(pattern string) []map[any]any {
	match := pathToRegexp.MustMatch(pattern, nil)
	var parserVariables map[string]any
	b, _ := jsoniter.Marshal(p.variables)
	_ = jsoniter.Unmarshal(b, &parserVariables)

	var matches []map[any]any
	_ = filepath.WalkDir(p.repoDir, func(path string, d fs.DirEntry, err error) error {
		rel, _ := filepath.Rel(p.repoDir, path)
		res, _ := match(rel)
		if res != nil {
			// create a map of only the size of the parser variables and the special params
			// from the match (_path and _dir).
			params := make(map[any]any, len(parserVariables)+2)

			for k, v := range parserVariables {
				params[k] = v
			}
			maps.Copy(params, res.Params)

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
func (p *Parser) list(v ...any) []any {
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
// This returns an empty map if the yaml is invalid.
func (p *Parser) parseYaml(contents string) map[string]any {
	var out map[string]any
	err := yaml.Unmarshal([]byte(contents), &out)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("failed to unmarshal yaml when calling parseYaml in the config template")
	}

	return out
}

// parseJson decodes the provided json contents and assigns decoded values into a
// generic out value. This can be used as a simple object in the templates.
// This returns an empty map if the json is invalid.
func (p *Parser) parseJson(contents string) map[string]any {
	var out map[string]any
	err := jsoniter.Unmarshal([]byte(contents), &out)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("failed to unmarshal json when calling parseJson in the config template")
	}

	return out
}

// isProduction returns true if the project is production.
func (p *Parser) isProduction(value string) bool {
	return p.config.IsProduction(value)
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

// strval returns the string representation of v.
// Taken from https://github.com/Masterminds/sprig/blob/581758eb7d96ae4d113649668fa96acc74d46e7f/strings.go#L174
func strval(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
