package logger

type NullLogger struct {
}

func (l *NullLogger) Error(msg string, args ...any) {
	return
}

func (l *NullLogger) Warn(msg string, args ...any) {
	return
}

func (l *NullLogger) Info(msg string, args ...any) {
	return
}
