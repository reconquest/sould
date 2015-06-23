package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ExecutionError struct {
	Err     error
	Output  []byte
	Command string
}

type Mirror struct {
	Name string
	Dir  string
}

func CreateMirror(
	storageDir string, name string, origin string,
) (Mirror, error) {
	mirrorDir := filepath.Join(storageDir, name)

	_, err := os.Stat(mirrorDir)
	if err == nil {
		return Mirror{}, fmt.Errorf(
			"directory '%s' already exists",
			mirrorDir,
		)
	} else if !os.IsNotExist(err) {
		return Mirror{}, err
	}

	err = os.MkdirAll(mirrorDir, 0770)
	if err != nil {
		return Mirror{}, err
	}

	mirror := Mirror{
		Name: name,
		Dir:  mirrorDir,
	}

	err = mirror.Clone(origin)
	if err != nil {
		return Mirror{}, err
	}

	return mirror, err
}

func GetMirror(
	storageDir string, name string,
) (Mirror, error) {
	mirrorDir := filepath.Join(storageDir, name)

	_, err := os.Stat(mirrorDir)
	if err != nil {
		return Mirror{}, err
	}

	mirror := Mirror{
		Name: name,
		Dir:  mirrorDir,
	}

	return mirror, nil
}

func (mirror Mirror) GetArchive() ([]byte, error) {
	return mirror.execute(exec.Command("git", "archive", "HEAD"))
}

func (mirror Mirror) Pull() error {
	_, err := mirror.execute(exec.Command("git", "remote", "update"))

	return err
}

func (mirror Mirror) execute(command *exec.Cmd) ([]byte, error) {
	command.Dir = mirror.Dir

	output, err := command.CombinedOutput()
	if err != nil {
		execErr := ExecutionError{
			Err:     err,
			Output:  output,
			Command: strings.Join(command.Args, " "),
		}

		return output, execErr
	}

	return output, nil
}

func (mirror Mirror) Clone(url string) error {
	_, err := mirror.execute(
		exec.Command("git", "clone", "--recursive", "--mirror", url, "."),
	)

	return err
}

func (mirror Mirror) GetOrigin() (string, error) {
	output, err := mirror.execute(
		exec.Command("git", "config", "--get", "remote.origin.url"),
	)

	return string(output), err
}

func (mirror Mirror) GetModDate() (time.Time, error) {
	var dirs = []string{
		"refs/heads",
		"refs/tags",
	}

	var modDate time.Time

	for _, dir := range dirs {
		fileinfo, err := os.Stat(mirror.Dir + "/" + dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return modDate, err
		}

		newModDate := fileinfo.ModTime()
		if newModDate.Second() > modDate.Second() {
			modDate = newModDate
		}
	}

	return modDate, nil
}

func (execErr ExecutionError) Error() string {
	return fmt.Sprintf(
		"%s: %s\n%s",
		execErr.Command, execErr.Err.Error(), execErr.Output,
	)
}
