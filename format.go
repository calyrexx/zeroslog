package zeroslog

import (
	"bytes"
	"log/slog"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/bytedance/sonic"
)

const (
	cReset  = "\x1b[0m"
	cRed    = "\x1b[91m"
	cGreen  = "\x1b[92m"
	cYellow = "\x1b[93m"
	cBlue   = "\x1b[96m"
)

var (
	lvlPlain = [...]string{"DEBU", "INFO", "WARN", "ERRO"}

	lvlColor = [...]string{
		cGreen + "DEBU" + cReset,
		cBlue + "INFO" + cReset,
		cYellow + "WARN" + cReset,
		cRed + "ERRO" + cReset,
	}

	staticSpaces = [64]byte{
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ',
		' ', ' ', ' ', ' ',
	}
)

func appendVal(b *bytes.Buffer, v any) {
	if v == nil {
		b.WriteString("nil")

		return
	}

	if handled := appendSimple(b, v); handled {
		return
	}

	if handled := appendNumeric(b, v); handled {
		return
	}

	if jsonBytes, err := sonic.Marshal(v); err == nil {
		b.Write(jsonBytes)
	} else {
		b.WriteString("<unsupported>")
	}
}

func appendSimple(b *bytes.Buffer, v any) bool {
	switch vv := v.(type) {
	case string:
		if stringsContainsSpace(vv) {
			b.WriteByte('"')
			b.WriteString(vv)
			b.WriteByte('"')
		} else {
			b.WriteString(vv)
		}

		return true
	case bool:
		b.WriteString(strconv.FormatBool(vv))

		return true
	case error:
		b.WriteString(vv.Error())

		return true
	case time.Duration:
		b.WriteString(vv.String())

		return true
	default:
		return false
	}
}

func appendNumeric(b *bytes.Buffer, v any) bool {
	switch vv := v.(type) {
	case int:
		var num [32]byte

		b.Write(strconv.AppendInt(num[:0], int64(vv), 10))

		return true
	case int64:
		var num [64]byte

		b.Write(strconv.AppendInt(num[:0], vv, 10))

		return true
	case float32:
		var num [32]byte

		b.Write(strconv.AppendFloat(num[:0], float64(vv), 'f', -1, 64))

		return true
	case float64:
		var num [64]byte

		b.Write(strconv.AppendFloat(num[:0], vv, 'f', -1, 64))

		return true
	default:
		return false
	}
}

func padRunes(buf *bytes.Buffer, msg string, width int) {
	msgLen := utf8.RuneCountInString(msg)

	switch {
	case msgLen < width:
		n := width - msgLen
		if n <= len(staticSpaces) {
			buf.Write(staticSpaces[:n])
		} else {
			for i := 0; i < n; i++ {
				buf.WriteByte(' ')
			}
		}
	case msgLen >= width:
		buf.WriteByte(' ')
	}
}

func chooseLevelStr(l slog.Level, color bool) string {
	switch l {
	case slog.LevelInfo:
		if color {
			return lvlColor[1]
		}

		return lvlPlain[1]
	case slog.LevelError:
		if color {
			return lvlColor[3]
		}

		return lvlPlain[3]
	case slog.LevelWarn:
		if color {
			return lvlColor[2]
		}

		return lvlPlain[2]
	default:
		if color {
			return lvlColor[0]
		}

		return lvlPlain[0]
	}
}

func levelColorCode(l slog.Level) string {
	switch l {
	case slog.LevelInfo:
		return cBlue
	case slog.LevelError:
		return cRed
	case slog.LevelWarn:
		return cYellow
	case slog.LevelDebug:
		return cGreen
	default:
		return ""
	}
}

func stringsContainsSpace(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' {
			return true
		}
	}

	return false
}

func writeQualifiedKey(buf *bytes.Buffer, groups []string, key string) {
	if len(groups) > 0 {
		for i, g := range groups {
			if i > 0 {
				buf.WriteByte('.')
			}

			buf.WriteString(g)
		}

		if key != "" {
			buf.WriteByte('.')
		}
	}

	buf.WriteString(key)
}
