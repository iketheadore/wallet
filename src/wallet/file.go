package wallet

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"errors"

	"gopkg.in/sirupsen/logrus.v1"
)

// This holds the root directory.
var (
	rootDir string
	log     = logrus.New()
)

// SetRootDir sets the root directory.
func SetRootDir(r string) error {
	var err error
	if rootDir, err = filepath.Abs(r); err != nil {
		return err
	}
	if err = os.MkdirAll(rootDir, os.FileMode(0700)); err != nil {
		return err
	}
	return nil
}

func ExtractLabel(filePath string) string {
	base := path.Base(filePath)
	return strings.TrimSuffix(base, string(FileExt))
}

func LabelPath(label string) string {
	return filepath.Join(rootDir, fmt.Sprintf("%s%s", label, FileExt))
}

func ListLabels() ([]string, error) {
	list, e := ioutil.ReadDir(rootDir)
	if e != nil {
		return nil, e
	}
	var out []string
	for _, info := range list {
		if info.IsDir() {
			continue
		}
		name := info.Name()
		if strings.HasSuffix(name, string(FileExt)) == false {
			continue
		}
		out = append(out, strings.TrimSuffix(name, string(FileExt)))
	}
	return out, nil
}

type LabelAction func(data []byte, label, fPath string, prefix Prefix) error

func RangeLabels(action LabelAction) error {
	list, err := ioutil.ReadDir(rootDir)
	if err != nil {
		return err
	}
	for _, info := range list {
		if info.IsDir() {
			continue
		}
		name := info.Name()
		if strings.HasSuffix(name, string(FileExt)) == false {
			continue
		}
		label := strings.TrimSuffix(name, string(FileExt))
		fPath := LabelPath(label)

		data, err := OpenAndReadAll(fPath)
		if err != nil {
			return err
		}

		if len(data) < PrefixSize {
			return errors.New("wallet file has invalid size")
		}

		var prefix Prefix
		copy(prefix[:PrefixSize], data[:PrefixSize])

		if err := action(data, label, fPath, prefix); err != nil {
			return errors.New("action failed with error: " + err.Error())
		}
	}
	return nil
}

func OpenAndReadAll(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}
