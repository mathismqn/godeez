package logger

import "log"

type Logger struct {
	l *log.Logger
}

func New(l *log.Logger) *Logger {
	return &Logger{l: l}
}

func (l *Logger) Infof(format string, args ...any) {
	if l.l != nil {
		l.l.Printf("[INFO] "+format, args...)
	}
}

func (l *Logger) Warnf(format string, args ...any) {
	if l.l != nil {
		l.l.Printf("[WARN] "+format, args...)
	}
}

func (l *Logger) Errorf(format string, args ...any) {
	if l.l != nil {
		l.l.Printf("[ERROR] "+format, args...)
	}
}
