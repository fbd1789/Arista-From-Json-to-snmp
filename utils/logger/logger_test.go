package logger

import (
	"log/slog"
	"regexp"
	"strings"
	"testing"
)

type captureStream struct {
	lines [][]byte
}

func (cs *captureStream) Write(bytes []byte) (int, error) {
	cs.lines = append(cs.lines, bytes)
	return len(bytes), nil
}

func Test_WritesToProvidedStream(t *testing.T) {
	cs := &captureStream{}
	handler := NewSyslogHandler(WithWriter(cs))
	logger := slog.New(handler)

	logger.Info("testing logger", "test", "test_val", "test2", "test2_val", slog.Group("group", "ga1", "ga1_val", "ga2", "ga2_val"))
	if len(cs.lines) != 1 {
		t.Errorf("expected 1 lines logged, got: %d", len(cs.lines))
	}

	lineMatcher := regexp.MustCompile(`testing logger`)
	// fmt.Println(">>>" + string(cs.lines[0]) + "<<<")

	line := string(cs.lines[0])
	if lineMatcher.MatchString(line) == false {
		t.Errorf("expected `testing logger` but found `%s`", line)
	}
	if !strings.HasSuffix(line, "\n") {
		t.Errorf("exected line to be terminated with `\\n` but found `%s`", line[len(line)-1:])
	}
}

func Test_Formatting(t *testing.T) {
	cs := &captureStream{}
	handler := NewSyslogHandler(WithWriter(cs))
	l := slog.New(handler)

	tests := []struct {
		want string
		fn   func(msg string, args ...any)
		msg  string
		args []any
	}{
		{
			want: "test message\n",
			fn:   l.Info,
			msg:  "test message",
		},
		{
			want: "[badfsdfz [bing bong=bang bump=boop]] [bunk foo=bar bat=bin] test message\n",
			fn:   l.Info,
			msg:  "test message",
			args: []any{slog.Group("badfsdfz", slog.Group("bing", "bong", "bang", "bump", "boop")), slog.Group("bunk", "foo", "bar", "bat", "bin")},
		},
	}

	for _, tst := range tests {
		tst.fn(tst.msg, tst.args...)
		line := string(cs.lines[len(cs.lines)-1])
		if line != tst.want {
			t.Errorf("wanted '%s' but got '%s'", tst.want, line)
		}
	}
}
