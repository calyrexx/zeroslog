package zeroslog

import (
	"bytes"
	"context"
	"github.com/bytedance/sonic"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"
)

type Handler struct {
	out      io.Writer
	timeFmt  string
	minLevel slog.Level
	color    bool
	json     bool
	mu       sync.Mutex
	attrs    []slog.Attr
	groups   []string
}

type Option func(*Handler)

func WithJSON() Option {
	return func(h *Handler) {
		h.json = true
	}
}

func WithTimeFormat(format string) Option {
	return func(h *Handler) {
		h.timeFmt = format
	}
}

func WithOutput(w io.Writer) Option {
	return func(h *Handler) {
		h.out = w
	}
}

func WithMinLevel(level slog.Level) Option {
	return func(h *Handler) {
		h.minLevel = level
	}
}

func WithColors() Option {
	return func(h *Handler) {
		h.color = true
	}
}

func New(opts ...Option) *Handler {
	h := &Handler{
		out:      os.Stderr,
		timeFmt:  time.RFC3339,
		minLevel: slog.LevelInfo,
		color:    false,
		json:     false,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *Handler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.minLevel
}

var bg = context.Background()

func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	if !h.Enabled(bg, r.Level) {
		return nil
	}

	if h.json {
		return h.handleJSON(r)
	}
	return h.handleText(r)
}

func (h *Handler) handleText(r slog.Record) error {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()

	levelStr := chooseLevelStr(r.Level, h.color)
	buf.WriteString(levelStr)

	// ----- timestamp -----
	buf.WriteByte('[')
	tsPtr := tsPool.Get().(*[]byte)
	tsBuf := (*tsPtr)[:0]
	tsBuf = r.Time.AppendFormat(tsBuf, h.timeFmt)
	buf.Write(tsBuf)
	*tsPtr = tsBuf
	tsPool.Put(tsPtr)
	buf.Write([]byte{']', ' '})

	// ----- message -----
	buf.WriteString(r.Message)
	padRunes(buf, r.Message, 50)

	// ----- groups -----
	if len(h.groups) > 0 {
		buf.WriteString("[")
		for i, g := range h.groups {
			if i > 0 {
				buf.WriteString(".")
			}
			buf.WriteString(g)
		}
		buf.Write([]byte{']', ' '})
	}

	// ----- previous attrs -----
	for _, a := range h.attrs {
		if h.color {
			buf.WriteString(levelColorCode(r.Level))
			buf.WriteString(a.Key)
			buf.WriteString(cReset)
		} else {
			buf.WriteString(a.Key)
		}
		buf.WriteByte('=')
		appendVal(buf, a.Value.Any())
		buf.WriteByte(' ')
	}

	// ----- attrs -----
	r.Attrs(func(a slog.Attr) bool {
		if h.color {
			buf.WriteString(levelColorCode(r.Level))
			buf.WriteString(a.Key)
			buf.WriteString(cReset)
		} else {
			buf.WriteString(a.Key)
		}
		buf.WriteByte('=')
		appendVal(buf, a.Value.Any())
		buf.WriteByte(' ')
		return true
	})

	buf.WriteByte('\n')

	h.mu.Lock()
	_, _ = h.out.Write(buf.Bytes())
	h.mu.Unlock()

	bufPool.Put(buf)
	return nil
}

func (h *Handler) handleJSON(r slog.Record) error {
	logMap := make(map[string]any)

	logMap["time"] = r.Time.Format(h.timeFmt)
	logMap["level"] = chooseLevelStr(r.Level, false)
	logMap["message"] = r.Message

	if len(h.groups) > 0 {
		logMap["context"] = h.groups
	}

	for _, a := range h.attrs {
		logMap[a.Key] = a.Value.Any()
	}

	r.Attrs(func(a slog.Attr) bool {
		logMap[a.Key] = a.Value.Any()
		return true
	})

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	encoder := sonic.ConfigFastest.NewEncoder(buf)
	if err := encoder.Encode(logMap); err != nil {
		return err
	}

	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.out.Write(buf.Bytes())
	return err
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	newAttrs = append(newAttrs, h.attrs...)
	newAttrs = append(newAttrs, attrs...)
	return &Handler{
		out:      h.out,
		timeFmt:  h.timeFmt,
		minLevel: h.minLevel,
		color:    h.color,
		attrs:    newAttrs,
		groups:   h.groups,
	}
}

func (h *Handler) WithGroup(group string) slog.Handler {
	newGroups := append([]string{}, h.groups...)
	if group != "" {
		newGroups = append(newGroups, group)
	}
	return &Handler{
		out:      h.out,
		timeFmt:  h.timeFmt,
		minLevel: h.minLevel,
		color:    h.color,
		attrs:    h.attrs,
		groups:   newGroups,
	}
}
