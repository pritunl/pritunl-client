package logger

import (
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"time"

	"github.com/pritunl/pritunl-client/service/colorize"
	"github.com/pritunl/pritunl-client/service/errortypes"
	"github.com/sirupsen/logrus"
)

var (
	blueArrow    = colorize.ColorString("▶", colorize.BlueBold, colorize.None)
	whiteDiamond = colorize.ColorString("◆", colorize.WhiteBold, colorize.None)
)

func entryFields(entry *logrus.Entry) (keys []string,
	fields map[string]interface{}, errStr string) {

	fields = make(map[string]interface{}, len(entry.Data)+2)

	var errData *errortypes.ErrorData
	for key, val := range entry.Data {
		switch key {
		case "error":
			if !isNil(val) {
				errStr = fmt.Sprintf("%s", val)
			}
		case "error_data":
			if data, ok := val.(*errortypes.ErrorData); ok && data != nil {
				errData = data
			}
		default:
			fields[key] = val
		}
	}

	if errData != nil {
		fields["error_key"] = errData.Error
		fields["error_msg"] = errData.Message
	}

	keys = make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return
}

func format(entry *logrus.Entry) (output []byte) {
	msg := fmt.Sprintf("%s%s %s %s",
		formatTime(entry.Time),
		formatLevel(entry.Level),
		blueArrow,
		entry.Message,
	)

	keys, fields, errStr := entryFields(entry)

	for _, key := range keys {
		msg += fmt.Sprintf(" %s %s=%v", whiteDiamond,
			colorize.ColorString(key, colorize.CyanBold, colorize.None),
			colorize.ColorString(fmt.Sprintf("%#v", fields[key]),
				colorize.GreenBold, colorize.None))
	}

	if errStr != "" {
		msg += "\n" + colorize.ColorString(errStr, colorize.Red, colorize.None)
	}

	if string(msg[len(msg)-1]) != "\n" {
		msg += "\n"
	}

	output = []byte(msg)

	return
}

func formatPlain(entry *logrus.Entry) (output []byte) {
	msg := fmt.Sprintf("%s%s ▶ %s",
		entry.Time.Format("[2006-01-02 15:04:05]"),
		formatLevelPlain(entry.Level),
		entry.Message,
	)

	keys, fields, errStr := entryFields(entry)

	for _, key := range keys {
		msg += fmt.Sprintf(" ◆ %s=%v", key,
			fmt.Sprintf("%#v", fields[key]))
	}

	if errStr != "" {
		msg += "\n" + errStr
	}

	if string(msg[len(msg)-1]) != "\n" {
		msg += "\n"
	}

	output = []byte(msg)

	return
}

func formatTime(timestamp time.Time) (str string) {
	return colorize.ColorString(
		timestamp.Format("[2006-01-02 15:04:05]"),
		colorize.Bold,
		colorize.None,
	)
}

func formatLevel(lvl logrus.Level) (str string) {
	var colorBg colorize.Color

	switch lvl {
	case logrus.InfoLevel:
		colorBg = colorize.CyanBg
		str = "[INFO]"
	case logrus.WarnLevel:
		colorBg = colorize.YellowBg
		str = "[WARN]"
	case logrus.ErrorLevel:
		colorBg = colorize.RedBg
		str = "[ERRO]"
	case logrus.FatalLevel:
		colorBg = colorize.RedBg
		str = "[FATL]"
	case logrus.PanicLevel:
		colorBg = colorize.RedBg
		str = "[PANC]"
	default:
		colorBg = colorize.BlackBg
	}

	str = colorize.ColorString(str, colorize.WhiteBold, colorBg)

	return
}

func formatLevelPlain(lvl logrus.Level) string {
	switch lvl {
	case logrus.InfoLevel:
		return "[INFO]"
	case logrus.WarnLevel:
		return "[WARN]"
	case logrus.ErrorLevel:
		return "[ERRO]"
	case logrus.FatalLevel:
		return "[FATL]"
	case logrus.PanicLevel:
		return "[PANC]"
	default:
	}

	return ""
}

func isNil(val interface{}) bool {
	if val == nil {
		return true
	}

	switch reflect.TypeOf(val).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface,
		reflect.Func, reflect.Chan:

		return reflect.ValueOf(val).IsNil()
	}

	return false
}

type formatter struct{}

func (f *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	if runtime.GOOS == "windows" {
		return formatPlain(entry), nil
	}
	return format(entry), nil
}
