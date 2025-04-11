package funcs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar"
	homedir "github.com/mitchellh/go-homedir"
	componentsFuncs "github.com/turbot/terraform-components/lang/funcs"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"

	"github.com/infracost/infracost/internal/logging"
)

func MakeFileFunc(baseDir string, encBase64 bool) function.Function {
	ff := componentsFuncs.MakeFileFunc(baseDir, encBase64)

	return function.New(&function.Spec{
		Params: ff.Params(),
		Type:   function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			path := args[0].AsString()
			if err := isFullPathWithinRepo(baseDir, path); err != nil {
				logging.Logger.Debug().Msgf("isFullPathWithinRepo error: %s returning a blank string for filesytem func", err)

				return cty.StringVal(""), err
			}

			c, err := ff.Call(args)

			// if we get an error calling the underlying file function this is likely because the path
			// argument has been transformed at some point because of infracost mocking/state fallbacks
			// so we return a blank string instead of failing the evaluation. This is safer than returning
			// an error as we in complex expression cases we can actually return a partial value instead
			// of a unknown value which will cause subsequent evaluations to fail.
			if err != nil {
				logging.Logger.Debug().Msgf("error calling file func: %s returning a blank string for filesytem func", err)
				return cty.StringVal(""), nil
			}

			return c, nil
		},
	})
}

// MakeTemplateFileFunc constructs a function that takes a file path and
// an arbitrary object of named values and attempts to render the referenced
// file as a template using HCL template syntax.
//
// The template itself may recursively call other functions so a callback
// must be provided to get access to those functions. The template cannot,
// however, access any variables defined in the scope: it is restricted only to
// those variables provided in the second function argument, to ensure that all
// dependencies on other graph nodes can be seen before executing this function.
//
// As a special exception, a referenced template file may not recursively call
// the templatefile function, since that would risk the same file being
// included into itself indefinitely.
func MakeTemplateFileFunc(baseDir string, funcsCb func() map[string]function.Function) function.Function {
	ff := componentsFuncs.MakeTemplateFileFunc(baseDir, funcsCb)
	return function.New(&function.Spec{
		Params: ff.Params(),
		Type: func(args []cty.Value) (cty.Type, error) {
			if !(args[0].IsKnown() && args[1].IsKnown()) {
				return cty.DynamicPseudoType, nil
			}

			path := args[0].AsString()
			if err := isFullPathWithinRepo(baseDir, path); err != nil {
				return cty.String, nil
			}

			return ff.ReturnTypeForValues(args)
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			path := args[0].AsString()
			if err := isFullPathWithinRepo(baseDir, path); err != nil {
				logging.Logger.Debug().Msgf("isFullPathWithinRepo error: %s returning a blank string for templatefile func", err)

				return cty.DynamicVal, nil
			}

			return ff.Call(args)
		},
	})

}

// MakeFileExistsFunc constructs a function that takes a path
// and determines whether a file exists at that path
func MakeFileExistsFunc(baseDir string) function.Function {
	ff := componentsFuncs.MakeFileExistsFunc(baseDir)
	return function.New(&function.Spec{
		Params: ff.Params(),
		Type:   function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			path := args[0].AsString()
			if err := isFullPathWithinRepo(baseDir, path); err != nil {
				logging.Logger.Debug().Msgf("isPathInRepo error: %s returning false for filesytem func", err)

				return cty.False, nil
			}

			return ff.Call(args)
		},
	})
}

// MakeFileSetFunc constructs a function that takes a glob pattern
// and enumerates a file set from that pattern
func MakeFileSetFunc(baseDir string) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "path",
				Type: cty.String,
			},
			{
				Name: "pattern",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.Set(cty.String)),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			path := args[0].AsString()
			pattern := args[1].AsString()

			if !filepath.IsAbs(path) {
				path = filepath.Join(baseDir, path)
			}

			pattern = filepath.Join(path, pattern)
			matches, err := doublestar.Glob(pattern)
			if err != nil {
				return cty.UnknownVal(cty.Set(cty.String)), fmt.Errorf("failed to glob pattern (%s): %s", pattern, err)
			}

			var matchVals []cty.Value
			for _, match := range matches {
				if err := isPathInRepo(match); err != nil {
					logging.Logger.Debug().Msgf("isPathInRepo error: %s skipping match for filesytem func", err)
					continue
				}

				fi, err := os.Stat(match)

				if err != nil {
					return cty.UnknownVal(cty.Set(cty.String)), fmt.Errorf("failed to stat (%s): %s", match, err)
				}

				if !fi.Mode().IsRegular() {
					continue
				}

				match, err = filepath.Rel(path, match)
				if err != nil {
					return cty.UnknownVal(cty.Set(cty.String)), fmt.Errorf("failed to trim path of match (%s): %s", match, err)
				}

				match = filepath.ToSlash(match)
				matchVals = append(matchVals, cty.StringVal(match))
			}

			if len(matchVals) == 0 {
				return cty.SetValEmpty(cty.String), nil
			}

			return cty.SetVal(matchVals), nil
		},
	})
}

// File reads the contents of the file at the given path.
//
// The file must contain valid UTF-8 bytes, or this function will return an error.
//
// The underlying function implementation works relative to a particular base
// directory, so this wrapper takes a base directory string and uses it to
// construct the underlying function before calling it.
func File(baseDir string, path cty.Value) (cty.Value, error) {
	fn := MakeFileFunc(baseDir, false)
	return fn.Call([]cty.Value{path})
}

// FileExists determines whether a file exists at the given path.
//
// The underlying function implementation works relative to a particular base
// directory, so this wrapper takes a base directory string and uses it to
// construct the underlying function before calling it.
func FileExists(baseDir string, path cty.Value) (cty.Value, error) {
	fn := MakeFileExistsFunc(baseDir)
	return fn.Call([]cty.Value{path})
}

// FileSet enumerates a set of files given a glob pattern
//
// The underlying function implementation works relative to a particular base
// directory, so this wrapper takes a base directory string and uses it to
// construct the underlying function before calling it.
func FileSet(baseDir string, path, pattern cty.Value) (cty.Value, error) {
	fn := MakeFileSetFunc(baseDir)
	return fn.Call([]cty.Value{path, pattern})
}

// FileBase64 reads the contents of the file at the given path.
//
// The bytes from the file are encoded as base64 before returning.
//
// The underlying function implementation works relative to a particular base
// directory, so this wrapper takes a base directory string and uses it to
// construct the underlying function before calling it.
func FileBase64(baseDir string, path cty.Value) (cty.Value, error) {
	fn := MakeFileFunc(baseDir, true)
	return fn.Call([]cty.Value{path})
}

// isFullPathWithinRepo joins path to the baseDir if needed and checks if
// it is within the repository directory.
func isFullPathWithinRepo(baseDir, path string) error {
	// isFullPathWithinRepo is a no-op when not running in github/gitlab app env.
	ciPlatform := os.Getenv("INFRACOST_CI_PLATFORM")
	if ciPlatform != "github_app" && ciPlatform != "gitlab_app" {
		return nil
	}

	fullPath, err := homedir.Expand(path)
	if err != nil {
		return fmt.Errorf("failed to expand ~: %s", err)
	}

	if !filepath.IsAbs(fullPath) {
		fullPath = filepath.Join(baseDir, fullPath)
	}

	return isPathInRepo(filepath.Clean(fullPath))
}

func isPathInRepo(path string) error {
	// isPathInRepo is a no-op when not running in github/gitlab app env.
	ciPlatform := os.Getenv("INFRACOST_CI_PLATFORM")
	if ciPlatform != "github_app" && ciPlatform != "gitlab_app" {
		return nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(wd, path)
	}

	// ensure the path resolves to the real symlink path
	path = symlinkPath(path)

	clean := filepath.Clean(wd)
	if wd != "" && !strings.HasPrefix(path, clean) {
		return fmt.Errorf("file %s is not within the repository directory %s", path, wd)
	}

	return nil
}

// symlinkPath checks the given file path and returns the real path if it is a
// symlink.
func symlinkPath(filepathStr string) string {
	fileInfo, err := os.Lstat(filepathStr)
	if err != nil {
		return filepathStr
	}

	if fileInfo.Mode()&os.ModeSymlink != 0 {
		realPath, err := filepath.EvalSymlinks(filepathStr)
		if err != nil {
			return filepathStr
		}

		return realPath
	}

	return filepathStr
}
