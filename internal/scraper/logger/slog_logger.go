package logger

import (
	"log/slog"
)

type SLogLogger struct {
	name   string
	kind   string
	logger *slog.Logger
}

func NewSLogLogger(logger *slog.Logger, options ...func(*SLogLogger)) *SLogLogger {
	svr := &SLogLogger{
		logger: logger,
	}
	for _, o := range options {
		o(svr)
	}
	return svr
}

func WithName(name string) func(*SLogLogger) {
	return func(s *SLogLogger) {
		s.name = name
	}
}

func WithKind(kind string) func(*SLogLogger) {
	return func(s *SLogLogger) {
		s.kind = kind
	}
}

func (l *SLogLogger) defaultArgs() []any {
	args := []any{}
	if l.name != "" {
		args = append(args, "sensor_name", l.name)
	}
	if l.kind != "" {
		args = append(args, "sensor_kind", l.kind)
	}
	return args
}

func (l *SLogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, append(l.defaultArgs(), args...)...)
}

func (l *SLogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, append(l.defaultArgs(), args...)...)
}

func (l *SLogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, append(l.defaultArgs(), args...)...)
}

func (l *SLogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, append(l.defaultArgs(), args...)...)
}
