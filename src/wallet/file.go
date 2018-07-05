package wallet

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"errors"

	"gopkg.in/sirupsen/logrus.v1"
)

// This holds the root directory.
var (
	log = logrus.New()
)

// LabelPath obtains the path to the wallet file of the given label.
func LabelPath(rootDir, label string) string {
	return filepath.Join(rootDir, fmt.Sprintf("%s%s", label, FileExt))
}

type LabelAction func(data []byte, label, fPath string, prefix Prefix) error

func RangeLabels(rootDir string, action LabelAction) error {
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
		fPath := LabelPath(rootDir, label)

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
