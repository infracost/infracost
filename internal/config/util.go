package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mitchellh/go-homedir"
)

func IsTruthy(s string) bool {
	return s == "1" || strings.EqualFold(s, "true")
}

func RootDir() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "../..")
}

func userConfigDir() string {
	dir, _ := homedir.Expand("~/.config/infracost")
	return dir
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}
