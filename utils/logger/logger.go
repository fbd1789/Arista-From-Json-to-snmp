package logger

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"log/syslog"
	"os"
	"path/filepath"
	"sync"
)

var LevelSeverityMap map[slog.Leveler]syslog.Priority = map[slog.Leveler]syslog.Priority{
	slog.LevelError: syslog.LOG_ERR,
	slog.LevelWarn:  syslog.LOG_WARNING,
	slog.LevelInfo:  syslog.LOG_INFO,
	slog.LevelDebug: syslog.LOG_DEBUG,
}

func EnableSyslogger(prio syslog.Priority, lvl slog.Leveler) error {
	l := slog.New(NewSyslogHandler(WithPriority(prio), WithLevel(lvl)))
	slog.SetDefault(l)
	return nil
}

// func EnableConsoleDebugLogger) {
// 	EnableConsoleLogging(slog.LevelDebug, true)
// }

func EnableConsoleLogger(level slog.Level, source bool) {
	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: source,
	}))
	slog.SetDefault(l)
}

type Option func(h *SyslogHandler)

func WithPriority(p syslog.Priority) Option {
	return func(h *SyslogHandler) {
		h.prio = p
	}
}

func WithLevel(lvl slog.Leveler) Option {
	return func(h *SyslogHandler) {
		h.lvl = lvl
	}
}

func WithNetwork(n string) Option {
	return func(h *SyslogHandler) {
		h.network = n
	}
}

func WithAddr(a string) Option {
	return func(h *SyslogHandler) {
		h.addr = a
	}
}

func WithTag(t string) Option {
	return func(h *SyslogHandler) {
		h.tag = t
	}
}

func WithWriter(w io.Writer) Option {
	return func(h *SyslogHandler) {
		h.w = w
	}
}

type SyslogHandler struct {
	m       *sync.Mutex
	w       io.Writer
	network string
	addr    string
	tag     string
	lvl     slog.Leveler
	prio    syslog.Priority
}

func NewSyslogHandler(opts ...Option) *SyslogHandler {
	h := &SyslogHandler{
		m:    &sync.Mutex{},
		tag:  filepath.Base(os.Args[0]),
		lvl:  slog.LevelInfo,
		prio: syslog.LOG_USER,
	}

	for _, fn := range opts {
		fn(h)
	}

	// only dial if not writer is set
	if h.w == nil {
		var err error
		h.w, err = syslog.Dial(h.network, h.addr, h.prio, h.tag)
		if err != nil {
			slog.Error("failed dial syslog server")
			os.Exit(1)
		}
	}

	return h
}

func (h *SyslogHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return lvl >= h.lvl.Level()
}

func (h *SyslogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *SyslogHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *SyslogHandler) appendAttr(b []byte, a slog.Attr) []byte {
	a.Value = a.Value.Resolve()
	if a.Equal(slog.Attr{}) {
		return b
	}

	switch a.Value.Kind() {
	case slog.KindGroup:
		g := a.Value.Group()

		if len(g) == 0 {
			return b
		}
		b = append(b, "["...)
		b = append(b, a.Key...)
		b = append(b, " "...)
		for i, ga := range g {
			b = h.appendAttr(b, ga)
			if i < len(g)-1 {
				b = append(b, " "...)
			}
		}
		b = append(b, "]"...)
	default:
		b = append(b, a.Key...)
		b = append(b, "="...)
		b = append(b, a.Value.String()...)
	}
	return b
}

func (h *SyslogHandler) writeSyslogLevel(w *syslog.Writer, buf []byte) error {
	var err error
	sev := LevelSeverityMap[h.lvl]
	m := string(buf)
	switch sev {
	case syslog.LOG_EMERG:
		err = w.Emerg(m)
	case syslog.LOG_ALERT:
		err = w.Alert(m)
	case syslog.LOG_CRIT:
		err = w.Crit(m)
	case syslog.LOG_ERR:
		err = w.Err(m)
	case syslog.LOG_WARNING:
		err = w.Warning(m)
	case syslog.LOG_NOTICE:
		err = w.Notice(m)
	case syslog.LOG_INFO:
		err = w.Info(m)
	case syslog.LOG_DEBUG:
		err = w.Debug(m)
	default:
		_, err = w.Write(buf)
	}
	return err
}

func (h *SyslogHandler) Handle(ctx context.Context, r slog.Record) error {
	var err error

	buf := make([]byte, 0, 1024)

	if r.NumAttrs() == 0 && r.Message == "" {
		return errors.New("refusing to log with neither message nor attrs")
	}

	// prival := int(h.prio)*8 + sev
	// buf = append(buf, "<"...)
	// buf = append(buf, strconv.Itoa(prival)...)
	// buf = append(buf, ">"...)
	// buf = append(buf, "1"...)
	// buf = append(buf, " "...)
	// buf = append(buf, time.Now().Format(time.RFC3339)...)
	// buf = append(buf, " "...)
	// buf = append(buf, h.tag...)
	// buf = append(buf, " "...)
	// si := int(LevelSeverityMap[h.lvl])
	// tag := fmt.Sprintf("PP-%d-%s", si)

	i := 0
	//if r.NumAttrs() > 0 {
	r.Attrs(func(a slog.Attr) bool {
		buf = h.appendAttr(buf, a)
		if i < r.NumAttrs()-1 {
			buf = append(buf, " "...)
		}
		i++
		return true
	})
	//}

	if r.Message != "" {
		if len(buf) > 0 {
			buf = append(buf, " "...)
		}

		buf = append(buf, r.Message...)
	}

	buf = append(buf, "\n"...)

	switch w := h.w.(type) {
	case *syslog.Writer:
		err = h.writeSyslogLevel(w, buf)
	default:
		_, err = w.Write(buf)
	}

	return err
}
