package main

import (
	"errors"
	"fmt"
	"github.com/gofrs/flock"
	"github.com/hanzhihua/yajs/utils"
	"os"
	"path"
	"runtime"
	"syscall"
)

const (
	pidFile = "yajs.pid"
)

func createRuntimeDir() string {

	runtimeDir := "/run/yajs"
	if runtime.GOOS == "darwin" {
		runtimeDir = "/tmp/yajs"
	}

	if err := makeRuntimeDir(runtimeDir); err != nil {
		utils.Logger.Errorf("fail to create runtime dir")
		os.Exit(-1)
	}
	return runtimeDir
}

func makeRuntimeDir(dir string) error {
	perm := os.FileMode(0o755)
	err := os.MkdirAll(dir, perm)
	if err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp(dir, "tmp")
	if err != nil {
		return err
	}
	fileName := tmpFile.Name()
	tmpFile.Close()
	os.Remove(fileName)
	return nil
}

func CreatePidFile() (*flock.Flock, error) {

	runtimeDir := createRuntimeDir()
	filename := pidFile
	fileFullName := path.Join(runtimeDir, filename)

	fd, err := os.OpenFile(fileFullName, os.O_CREATE|os.O_RDWR, 0o664)
	if err != nil {
		return nil, err
	}

	defer fd.Close()

	fd.Truncate(0)
	_, err = fd.WriteString(fmt.Sprintf("%d", os.Getpid()))
	if err != nil {
		return nil, err
	}

	lock := flock.New(fileFullName)
	b, err := lock.TryLock()
	if err != nil {
		return nil, err
	}
	if !b {
		return nil, errors.New("fail to lock pid file")
	}

	return lock, nil
}

func getPid() (int, error) {
	pid := -1
	runtimeDir := createRuntimeDir()
	fd, err := os.OpenFile(path.Join(runtimeDir, pidFile), os.O_RDONLY, 0o664)
	if err != nil {
		return pid, err
	}
	defer fd.Close()
	if _, err = fmt.Fscanf(fd, "%d", &pid); err != nil {
		return pid, err
	}
	return pid, nil
}

func removePidFile(lock *flock.Flock) {
	filename := lock.Path()
	lock.Close()
	os.Remove(filename)
}

func Reload() error {
	pid, err := getPid()
	if err != nil {
		return err
	}

	if process, err := os.FindProcess(pid); err == nil {
		return process.Signal(syscall.SIGUSR1)
	}
	return err
}
