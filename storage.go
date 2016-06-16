package main

import (
	"os"
	"path/filepath"
	"strings"
)

func getAllMirrors(rootDir string) ([]string, error) {
	var mirrors []string

	err := filepath.Walk(
		rootDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				return nil
			}

			mirror := strings.Trim(
				strings.TrimPrefix(path, rootDir),
				"/",
			)

			if mirror == "" {
				return nil
			}

			if isFileExists(filepath.Join(path, "HEAD")) {
				mirrors = append(mirrors, mirror)
				return filepath.SkipDir
			}

			return nil
		},
	)

	return mirrors, err
}

func isFileExists(path string) bool {
	stat, err := os.Stat(path)
	return !os.IsNotExist(err) && !stat.IsDir()
}
