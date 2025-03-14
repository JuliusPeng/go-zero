package pathx

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/tal-tech/go-zero/tools/goctl/internal/version"
)

// NL defines a new line
const (
	NL       = "\n"
	goctlDir = ".goctl"
	gitDir   = ".git"
)

var goctlHome string

// RegisterGoctlHome register goctl home path
func RegisterGoctlHome(home string) {
	goctlHome = home
}

// CreateIfNotExist creates a file if it is not exists
func CreateIfNotExist(file string) (*os.File, error) {
	_, err := os.Stat(file)
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("%s already exist", file)
	}

	return os.Create(file)
}

// RemoveIfExist deletes the specified file if it is exists
func RemoveIfExist(filename string) error {
	if !FileExists(filename) {
		return nil
	}

	return os.Remove(filename)
}

// RemoveOrQuit deletes the specified file if read a permit command from stdin
func RemoveOrQuit(filename string) error {
	if !FileExists(filename) {
		return nil
	}

	fmt.Printf("%s exists, overwrite it?\nEnter to overwrite or Ctrl-C to cancel...",
		aurora.BgRed(aurora.Bold(filename)))
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	return os.Remove(filename)
}

// FileExists returns true if the specified file is exists
func FileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

// FileNameWithoutExt returns a file name without suffix
func FileNameWithoutExt(file string) string {
	return strings.TrimSuffix(file, filepath.Ext(file))
}

// GetGoctlHome returns the path value of the goctl home where Join $HOME with .goctl
func GetGoctlHome() (string, error) {
	if len(goctlHome) != 0 {
		return goctlHome, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, goctlDir), nil
}

// GetGitHome returns the git home of goctl.
func GetGitHome() (string, error) {
	goctlH, err := GetGoctlHome()
	if err != nil {
		return "", err
	}

	return filepath.Join(goctlH, gitDir), nil
}

// GetTemplateDir returns the category path value in GoctlHome where could get it by GetGoctlHome
func GetTemplateDir(category string) (string, error) {
	home, err := GetGoctlHome()
	if err != nil {
		return "", err
	}
	if home == goctlHome {
		// backward compatible, it will be removed in the feature
		// backward compatible start
		beforeTemplateDir := filepath.Join(home, version.GetGoctlVersion(), category)
		fs, _ := ioutil.ReadDir(beforeTemplateDir)
		var hasContent bool
		for _, e := range fs {
			if e.Size() > 0 {
				hasContent = true
			}
		}
		if hasContent {
			return beforeTemplateDir, nil
		}
		// backward compatible end

		return filepath.Join(home, category), nil
	}

	return filepath.Join(home, version.GetGoctlVersion(), category), nil
}

// InitTemplates creates template files GoctlHome where could get it by GetGoctlHome
func InitTemplates(category string, templates map[string]string) error {
	dir, err := GetTemplateDir(category)
	if err != nil {
		return err
	}

	if err := MkdirIfNotExist(dir); err != nil {
		return err
	}

	for k, v := range templates {
		if err := createTemplate(filepath.Join(dir, k), v, false); err != nil {
			return err
		}
	}

	return nil
}

// CreateTemplate writes template into file even it is exists
func CreateTemplate(category, name, content string) error {
	dir, err := GetTemplateDir(category)
	if err != nil {
		return err
	}
	return createTemplate(filepath.Join(dir, name), content, true)
}

// Clean deletes all templates and removes the parent directory
func Clean(category string) error {
	dir, err := GetTemplateDir(category)
	if err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

// LoadTemplate gets template content by the specified file
func LoadTemplate(category, file, builtin string) (string, error) {
	dir, err := GetTemplateDir(category)
	if err != nil {
		return "", err
	}

	file = filepath.Join(dir, file)
	if !FileExists(file) {
		return builtin, nil
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// SameFile compares the between path if the same path,
// it maybe the same path in case case-ignore, such as:
// /Users/go_zero and /Users/Go_zero, as far as we know,
// this case maybe appear on macOS and Windows.
func SameFile(path1, path2 string) (bool, error) {
	stat1, err := os.Stat(path1)
	if err != nil {
		return false, err
	}

	stat2, err := os.Stat(path2)
	if err != nil {
		return false, err
	}

	return os.SameFile(stat1, stat2), nil
}

func createTemplate(file, content string, force bool) error {
	if FileExists(file) && !force {
		return nil
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	return err
}

// MustTempDir creates a temporary directory
func MustTempDir() string {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatalln(err)
	}

	return dir
}
