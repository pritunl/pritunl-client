package utils

import (
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/errortypes"
)

var (
	writeLock = sync.Mutex{}
)

func Chmod(pth string, mode os.FileMode) (err error) {
	err = os.Chmod(pth, mode)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrapf(err, "utils: Failed to chmod %s", pth),
		}
		return
	}

	return
}

func Exists(pth string) (exists bool, err error) {
	_, err = os.Stat(pth)
	if err == nil {
		exists = true
		return
	}

	if os.IsNotExist(err) {
		err = nil
		return
	}

	err = &errortypes.ReadError{
		errors.Wrapf(err, "utils: Failed to stat %s", pth),
	}
	return
}

func ExistsDir(pth string) (exists bool, err error) {
	stat, err := os.Stat(pth)
	if err == nil {
		exists = stat.IsDir()
		return
	}

	if os.IsNotExist(err) {
		err = nil
		return
	}

	err = &errortypes.ReadError{
		errors.Wrapf(err, "utils: Failed to stat %s", pth),
	}
	return
}

func ExistsFile(pth string) (exists bool, err error) {
	stat, err := os.Stat(pth)
	if err == nil {
		exists = !stat.IsDir()
		return
	}

	if os.IsNotExist(err) {
		err = nil
		return
	}

	err = &errortypes.ReadError{
		errors.Wrapf(err, "utils: Failed to stat %s", pth),
	}
	return
}

func ExistsMkdir(pth string, perm os.FileMode) (err error) {
	exists, err := ExistsDir(pth)
	if err != nil {
		return
	}

	if !exists {
		err = os.MkdirAll(pth, perm)
		if err != nil {
			err = &errortypes.WriteError{
				errors.Wrapf(err, "utils: Failed to mkdir %s", pth),
			}
			return
		}
	}

	return
}

func ExistsRemove(pth string) (err error) {
	exists, err := Exists(pth)
	if err != nil {
		return
	}

	if exists {
		err = os.RemoveAll(pth)
		if err != nil {
			err = &errortypes.WriteError{
				errors.Wrapf(err, "utils: Failed to rm %s", pth),
			}
			return
		}
	}

	return
}

func Remove(path string) (err error) {
	err = os.Remove(path)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrapf(err, "utils: Failed to remove '%s'", path),
		}
		return
	}

	return
}

func RemoveAll(path string) (err error) {
	err = os.RemoveAll(path)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrapf(err, "utils: Failed to remove '%s'", path),
		}
		return
	}

	return
}

func ContainsDir(pth string) (hasDir bool, err error) {
	exists, err := ExistsDir(pth)
	if !exists {
		return
	}

	entries, err := ioutil.ReadDir(pth)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrapf(err, "queue: Failed to read dir %s", pth),
		}
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			hasDir = true
			return
		}
	}

	return
}

func Create(path string, perm os.FileMode) (file *os.File, err error) {
	file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrapf(err, "utils: Failed to create '%s'", path),
		}
		return
	}

	return
}

func CreateWrite(path string, data string, perm os.FileMode) (err error) {
	file, err := Create(path, perm)
	if err != nil {
		return
	}
	defer func() {
		err = file.Close()
		if err != nil {
			err = &errortypes.WriteError{
				errors.Wrapf(err, "utils: Failed to write '%s'", path),
			}
			return
		}
	}()

	_, err = file.WriteString(data)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrapf(err, "utils: Failed to write to file '%s'", path),
		}
		return
	}

	return
}

func CreateWriteLock(path string, data string, perm os.FileMode) (err error) {
	writeLock.Lock()
	defer writeLock.Unlock()

	file, err := Create(path, perm)
	if err != nil {
		return
	}
	defer func() {
		err = file.Close()
		if err != nil {
			err = &errortypes.WriteError{
				errors.Wrapf(err, "utils: Failed to write '%s'", path),
			}
			return
		}
	}()

	_, err = file.WriteString(data)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrapf(err, "utils: Failed to write to file '%s'", path),
		}
		return
	}

	return
}

func Copy(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrapf(err, "utils: Failed to read file '%s'", src),
		}
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrapf(err, "utils: Failed to create file '%s'", dst),
		}
		return
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrapf(err,
				"utils: Failed to copy file '%s' to '%s'", src, dst),
		}
		return
	}

	err = out.Close()
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrapf(err, "utils: Failed to write file '%s'", dst),
		}
		return
	}

	return
}
