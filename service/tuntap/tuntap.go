package tuntap

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pritunl/pritunl-client/service/config"
	"github.com/pritunl/pritunl-client/service/platform"
	"github.com/pritunl/pritunl-client/service/utils"
)

var (
	curSize      = 0
	curTotalSize = 0
	taps         = []string{}
	tapsLock     = sync.Mutex{}
	renameError  = false
)

func Configure() (err error) {
	if runtime.GOOS != "windows" {
		return
	}

	metric := config.Config.InterfaceMetric

	metricStr := ""
	if metric == 0 {
		metricStr = "automatic"
	} else {
		metricStr = strconv.Itoa(metric)
	}

	size := Size()

	systemDir, err := platform.SystemDirectory()
	if err != nil {
		return
	}

	pth := filepath.Join(systemDir, "netsh.exe")

	for i := 0; i < size; i++ {
		_, _ = utils.ExecInputOutputCombindLogged(
			fmt.Sprintf(
				"interface ipv4 set interface \"Pritunl %d\" metric=%s",
				i+1,
				metricStr,
			),
			pth,
		)
		_, _ = utils.ExecInputOutputCombindLogged(
			fmt.Sprintf(
				"interface ipv6 set interface \"Pritunl %d\" metric=%s",
				i+1,
				metricStr,
			),
			pth,
		)
	}

	return
}

func get() (adpaters []string, count int, err error) {
	toolpath := utils.TapCtlPath()

	output, err := utils.ExecCombinedOutputLogged(
		nil,
		toolpath,
		"list",
	)
	if err != nil {
		return
	}

	adpaters = []string{}
	for _, line := range strings.Split(output, "\n") {
		lines := strings.Fields(line)
		if len(lines) < 2 {
			continue
		}

		count += 1

		if !renameError && !strings.Contains(
			strings.ToLower(line), "pritunl") {

			continue
		}

		adpaters = append(adpaters, lines[0])
	}

	return
}

func clean() (err error) {
	toolpath := utils.TapCtlPath()

	adapters, totalCount, err := get()
	if err != nil {
		return
	}

	for _, adapter := range adapters {
		_, _ = utils.ExecCombinedOutputLogged(
			[]string{
				"No devices",
			},
			toolpath,
			"delete",
			adapter,
		)
		totalCount -= 1
	}

	curSize = 0
	curTotalSize = totalCount
	taps = []string{}

	return
}

func Clean() (err error) {
	tapsLock.Lock()
	defer tapsLock.Unlock()

	err = clean()
	if err != nil {
		return
	}

	return
}

func Resize(size int) (err error) {
	tapsLock.Lock()
	defer tapsLock.Unlock()

	toolpath := utils.TapCtlPath()

	if size <= 3 {
		size = 3
	} else if size < 6 {
		size = 6
	} else {
		size = 9
	}

	add := size - curSize

	output := ""

	for i := 0; i < add; i++ {
		tapName := fmt.Sprintf("Pritunl %d", curSize+1)
		tapDcoName := fmt.Sprintf("Pritunl DCO %d", curSize+1)

		if renameError {
			output, err = utils.ExecCombinedOutputLogged(
				nil,
				toolpath,
				"create",
				"--hwid", "tap0901",
			)
		} else {
			output, err = utils.ExecCombinedOutputLogged(
				nil,
				toolpath,
				"create",
				"--name", tapName,
				"--hwid", "tap0901",
			)
		}
		if err != nil {
			err = nil

			if strings.Contains(strings.ToLower(output), "renaming") {
				if !renameError {
					renameError = true
					_ = clean()
				}
			} else {
				_, _ = utils.ExecCombinedOutputLogged(
					[]string{
						"No devices",
					},
					toolpath,
					"delete",
					tapName,
				)
			}

			if renameError {
				output, err = utils.ExecCombinedOutputLogged(
					nil,
					toolpath,
					"create",
					"--hwid", "tap0901",
				)
			} else {
				output, err = utils.ExecCombinedOutputLogged(
					nil,
					toolpath,
					"create",
					"--name", tapName,
					"--hwid", "tap0901",
				)
			}
			if err != nil {
				err = nil

				if size-curTotalSize > 0 {
					_, err = utils.ExecCombinedOutputLogged(
						nil,
						toolpath,
						"create",
						"--hwid", "tap0901",
					)
					if err != nil {
						_ = clean()
						return
					}

					curTotalSize += 1
				}

				time.Sleep(200 * time.Millisecond)
				continue
			}
		}

		output, err = utils.ExecCombinedOutputLogged(
			nil,
			toolpath,
			"create",
			"--name", tapDcoName,
			"--hwid", "ovpn-dco",
		)
		if err != nil {
			err = nil
		}

		curSize += 1
		curTotalSize += 1
		taps = append(taps, tapName)

		time.Sleep(200 * time.Millisecond)
	}

	sort.Strings(taps)

	return
}

func Size() int {
	tapsLock.Lock()
	defer tapsLock.Unlock()

	return curSize
}

func Acquire(dco bool) (tap string) {
	tapsLock.Lock()
	defer tapsLock.Unlock()

	if len(taps) == 0 {
		tap = "null"
		return
	}

	tap, taps = taps[0], taps[1:]

	if dco {
		tap = strings.Replace(tap, "Pritunl", "Pritunl DCO", 1)
	}

	return
}

func Release(tap string) {
	tapsLock.Lock()
	defer tapsLock.Unlock()

	if tap == "null" {
		return
	}

	tap = strings.Replace(tap, "DCO ", "", 1)

	taps = append(taps, tap)
	sort.Strings(taps)
}
