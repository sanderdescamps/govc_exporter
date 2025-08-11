package logger

type NullLogger struct {
}

func (l *NullLogger) Error(msg string, args ...any) {
}

func (l *NullLogger) Warn(msg string, args ...any) {
}

func (l *NullLogger) Info(msg string, args ...any) {
}

func (l *NullLogger) Debug(msg string, args ...any) {
}
