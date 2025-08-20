package zeroslog

import (
	"bytes"
	"context"
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
	mu       sync.Mutex
	attrs    []slog.Attr
	groups   []string
}

type Option func(*Handler)

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

	// ----- previous attrs -----
	for _, a := range h.attrs {
		h.writeAttr(buf, r.Level, h.groups, a)
	}

	// ----- attrs -----
	r.Attrs(func(a slog.Attr) bool {
		h.writeAttr(buf, r.Level, h.groups, a)
		return true
	})

	buf.WriteByte('\n')

	h.mu.Lock()
	_, _ = h.out.Write(buf.Bytes())
	h.mu.Unlock()

	bufPool.Put(buf)
	return nil
}

func (h *Handler) writeAttr(buf *bytes.Buffer, level slog.Level, groups []string, a slog.Attr) {
	switch a.Value.Kind() {
	case slog.KindGroup:
		groups = append(groups, a.Key)
		for _, ga := range a.Value.Group() {
			h.writeAttr(buf, level, groups, ga)
		}
		groups = groups[:len(groups)-1]
	default:
		if h.color {
			buf.WriteString(levelColorCode(level))
			writeQualifiedKey(buf, groups, a.Key)
			buf.WriteString(cReset)
		} else {
			writeQualifiedKey(buf, groups, a.Key)
		}
		buf.WriteByte('=')
		appendVal(buf, a.Value.Any())
		buf.WriteByte(' ')
	}
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
