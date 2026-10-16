package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/flytam/filenamify"
)

var downloadDirectoryWindows = `Downloads`

var downloadDirectoryLinux = `Downloads`

var downloadDirectory string

func init() {
	var candidate string

	switch runtime.GOOS {
	case "windows":
		candidate = downloadDirectoryWindows
	case "linux":
		candidate = downloadDirectoryLinux
	default:
		panic("Unsupported platform")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		panic("Failed to get user home/profile directory")
	}

	var target string

	target = filepath.Join(home, candidate)
	_, err = os.ReadDir(target)
	if err != nil {
		if os.IsNotExist(err) {
			panic("Directory not found")
		}
		panic(err)
	}

	downloadDirectory = target
}

type JsonStore struct {
	root string
}

func GetJsonStore() (*JsonStore, error) {
	storeDirectory := filepath.Join(downloadDirectory, "EDCITR")

	return &JsonStore{root: storeDirectory}, nil
}

func (c *JsonStore) Root() string {
	return c.root
}

func (c *JsonStore) NewEventFile(eventName string) (io.WriteCloser, error) {
	return c.newFile("events", eventName)
}

func (c *JsonStore) NewNPCMessageFile(identifier string) (io.WriteCloser, error) {
	return c.newFile("npc-messages", identifier)
}

func (c *JsonStore) NewNotificationFile(identifier string) (io.WriteCloser, error) {
	return c.newFile("notifications", identifier)
}

func (c *JsonStore) newFile(subdir, filename string) (io.WriteCloser, error) {
	_, err := os.Stat(filepath.Join(c.root, subdir))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(filepath.Join(c.root, subdir), 0755)
			if err != nil {
				return nil, fmt.Errorf("failed to create event directory: %w", err)
			}
		} else {
			return nil, fmt.Errorf("cannot read event directory: %w", err)
		}
	}

	sanitized, err := filenamify.Filenamify(filename, filenamify.Options{Replacement: "_"})
	if err != nil {
		return nil, fmt.Errorf("failed to sanitize identifier: %w", err)
	}
	filenameWithExt := fmt.Sprintf("%s.json", sanitized)
	fd, err := os.OpenFile(filepath.Join(c.root, subdir, filenameWithExt), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to create event file: %w", err)
	}
	return fd, nil
}
