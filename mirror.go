package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kovetskiy/executil"
)

type Mirror struct {
	Name string
	URL  string
	Dir  string
}

func (mirror *Mirror) String() string {
	return mirror.Name + " (" + mirror.URL + ")"
}

func CreateMirror(
	storageDir string, name string, url string,
) (Mirror, error) {
	mirrorDir := filepath.Join(storageDir, name)

	_, err := os.Stat(mirrorDir)
	if err == nil {
		return Mirror{}, fmt.Errorf(
			"directory %s already exists",
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
		URL:  url,
		Dir:  mirrorDir,
	}

	err = mirror.Clone(url)
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

	url, err := mirror.GetURL()
	if err != nil {
		return mirror, err
	}

	mirror.URL = url

	return mirror, nil
}

func (mirror *Mirror) Archive(
	stdout io.Writer, treeish string,
) error {
	cmd := exec.Command("git", "archive", treeish)
	cmd.Dir = mirror.Dir
	cmd.Stdout = stdout

	_, _, err := executil.Run(cmd, executil.IgnoreStdout)
	return err
}

func (mirror *Mirror) Fetch() error {
	cmd := exec.Command("git", "remote", "update")
	cmd.Dir = mirror.Dir
	_, _, err := executil.Run(cmd)
	return err
}

func (mirror *Mirror) SpoofChangeset(branch, tag string) error {
	cmd := exec.Command("git", "branch", "--force", branch, tag)
	cmd.Dir = mirror.Dir
	_, _, err := executil.Run(cmd)
	if err != nil {
		return err
	}

	return mirror.removeTag(tag)
}

func (mirror *Mirror) removeTag(tag string) error {
	cmd := exec.Command("git", "tag", "-d", tag)
	cmd.Dir = mirror.Dir
	_, _, err := executil.Run(cmd)
	return err
}

func (mirror *Mirror) Clone(url string) error {
	cmd := exec.Command("git", "clone", "--recursive", "--mirror", url, ".")
	cmd.Dir = mirror.Dir
	_, _, err := executil.Run(cmd)
	return err
}

func (mirror *Mirror) GetURL() (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = mirror.Dir
	stdout, _, err := executil.Run(cmd)
	return strings.TrimSpace(string(stdout)), err
}

func (mirror *Mirror) GetModifyDate() (time.Time, error) {
	dirs := []string{
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
		if newModDate.Unix() > modDate.Unix() {
			modDate = newModDate
		}
	}

	return modDate, nil
}
