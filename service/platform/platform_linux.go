package platform

import (
	"os"
	"syscall"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-client/service/errortypes"
)

var (
	instanceLockFile *os.File
)

func SystemDirectory() (pth string, err error) {
	return
}

func MkdirLinkedSecure(pth string) (err error) {
	info, err := os.Lstat(pth)
	if !os.IsNotExist(err) {
		if err != nil {
			err = &errortypes.ReadError{
				errors.Wrap(err, "utils: Failed to check path"),
			}
			return
		}

		if info.Mode()&os.ModeSymlink != 0 {
			_, err = os.Stat(pth)
			if err != nil {
				err = nil
				return
			}
		}

		err = os.Chown(pth, os.Getuid(), os.Getuid())
		if err != nil {
			err = &errortypes.ReadError{
				errors.Wrap(err, "utils: Failed to chown directory"),
			}
			return
		}

		err = os.Chmod(pth, 0700)
		if err != nil {
			err = &errortypes.ReadError{
				errors.Wrap(err, "utils: Failed to chmod directory"),
			}
			return
		}
	} else {
		err = os.MkdirAll(pth, 0700)
		if err != nil {
			err = &errortypes.ReadError{
				errors.Wrap(err, "utils: Failed to create directory"),
			}
			return
		}
	}

	return
}

func MkdirSecure(pth string) (err error) {
	if _, err = os.Stat(pth); !os.IsNotExist(err) {
		err = os.Chown(pth, os.Getuid(), os.Getuid())
		if err != nil {
			err = &errortypes.ReadError{
				errors.Wrap(err, "utils: Failed to chown directory"),
			}
			return
		}

		err = os.Chmod(pth, 0700)
		if err != nil {
			err = &errortypes.ReadError{
				errors.Wrap(err, "utils: Failed to chmod directory"),
			}
			return
		}
	} else {
		err = os.MkdirAll(pth, 0700)
		if err != nil {
			err = &errortypes.ReadError{
				errors.Wrap(err, "utils: Failed to create directory"),
			}
			return
		}
	}

	return
}

func MkdirReadSecure(pth string) (err error) {
	if _, err = os.Stat(pth); !os.IsNotExist(err) {
		err = os.Chown(pth, os.Getuid(), os.Getuid())
		if err != nil {
			err = &errortypes.ReadError{
				errors.Wrap(err, "utils: Failed to chown directory"),
			}
			return
		}

		err = os.Chmod(pth, 0755)
		if err != nil {
			err = &errortypes.ReadError{
				errors.Wrap(err, "utils: Failed to chmod directory"),
			}
			return
		}
	} else {
		err = os.MkdirAll(pth, 0755)
		if err != nil {
			err = &errortypes.ReadError{
				errors.Wrap(err, "utils: Failed to create directory"),
			}
			return
		}
	}

	return
}

func InstanceLock(pth string) (locked bool, err error) {
	file, err := os.OpenFile(pth, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		err = &errortypes.WriteError{
			errors.Wrapf(err, "platform: Failed to open lock %s", pth),
		}
		return
	}

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		_ = file.Close()
		if err == syscall.EWOULDBLOCK {
			err = nil
			return
		}

		err = &errortypes.WriteError{
			errors.Wrapf(err, "platform: Failed to lock %s", pth),
		}
		return
	}

	instanceLockFile = file
	locked = true

	return
}
