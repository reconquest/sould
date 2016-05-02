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

			if path == "config" && !info.IsDir() {
				return filepath.SkipDir
			}

			if !info.IsDir() {
				return nil
			}

			mirror := strings.Trim(
				strings.TrimPrefix(path, rootDir),
				"/",
			)

			if mirror != "" {
				mirrors = append(mirrors, mirror)
			}

			return nil
		},
	)

	return mirrors, err
}
