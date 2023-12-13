package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// CreateDirectoryStructure reads a tree command output file and creates the
// directory structure in the specified temp directory.
func CreateDirectoryStructure(t *testing.T, treeOutputLocation string, tmpDir string) {
	file, err := os.Open(treeOutputLocation)
	require.NoError(t, err)

	defer file.Close()

	val, _ := os.ReadFile(treeOutputLocation)
	lines := strings.Split(string(val), "\n")
	fs := buildFileSystemTree(lines, 0, 1)
	writeFileSystem(t, fs, tmpDir)
}

var indentation = "    "

func stripTreeFormatting(line string) string {
	line = strings.ReplaceAll(line, "├── ", indentation)
	line = strings.ReplaceAll(line, "└── ", indentation)
	line = strings.ReplaceAll(line, "│   ", indentation)

	return line
}

func getDepth(line string) int {
	return strings.Count(line, indentation)
}

func createFileWithContents(filePath string) error {
	var content string
	filename := filepath.Base(filePath)
	filename = stripTreeFormatting(strings.TrimSpace(filename))

	switch filename {
	case "main.tf":
		content = `provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}`
	case "backend.tf":
		content = `terraform {
  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "example_corp"

    workspaces {
      name = "example_corp/web-app-prod"
    }
  }
}`
	default:
		content = ""
	}

	return os.WriteFile(filePath, []byte(content), 0600)
}

// FindDirectoriesWithTreeFile finds and returns a slice of directory paths within the given base directory
// that contain a "tree.txt" file.
func FindDirectoriesWithTreeFile(t *testing.T, baseDir string) []string {
	var directoriesWithTreeFile []string

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		treeFilePath := filepath.Join(path, "tree.txt")
		if _, err := os.Stat(treeFilePath); err == nil {
			directoriesWithTreeFile = append(directoriesWithTreeFile, path)
		}

		return nil
	})
	require.NoError(t, err)

	return directoriesWithTreeFile
}

type fileSystemNode struct {
	name     string
	isDir    bool
	children []*fileSystemNode
}

func writeFileSystem(t *testing.T, node *fileSystemNode, path string) {
	if node == nil {
		return
	}

	fullPath := filepath.Join(path, node.name)
	if node.isDir {
		err := os.MkdirAll(fullPath, 0755)
		require.NoError(t, err)

		for _, child := range node.children {
			writeFileSystem(t, child, fullPath)
		}

		return
	}

	err := createFileWithContents(fullPath)
	require.NoError(t, err)
}

func buildFileSystemTree(lines []string, currentLine int, currentIndent int) *fileSystemNode {
	if currentLine >= len(lines) {
		return nil
	}

	line := lines[currentLine]
	formattedLine := stripTreeFormatting(line)
	indent := getDepth(formattedLine)

	if indent != currentIndent {
		return nil
	}

	node := &fileSystemNode{
		name:  strings.TrimSpace(formattedLine),
		isDir: filepath.Ext(line) == "",
	}

	nextIndent := indent + 1
	for nextLine := currentLine + 1; nextLine < len(lines); nextLine++ {
		child := buildFileSystemTree(lines, nextLine, nextIndent)
		if child == nil {
			break
		}

		node.children = append(node.children, child)
		nextLine += len(child.children)
	}

	return node
}
