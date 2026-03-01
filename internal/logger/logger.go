package logger

import "log"

type Logger struct {
	l *log.Logger
}

func New(l *log.Logger) *Logger {
	return &Logger{l: l}
}

func (l *Logger) logf(level, format string, args ...any) {
	if l.l != nil {
		l.l.Printf("["+level+"] "+format, args...)
	}
}

func (l *Logger) Infof(format string, args ...any)  { l.logf("INFO", format, args...) }
func (l *Logger) Warnf(format string, args ...any)  { l.logf("WARN", format, args...) }
func (l *Logger) Errorf(format string, args ...any) { l.logf("ERROR", format, args...) }
