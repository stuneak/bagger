package logger

import (
	"log"
	"runtime"
	"strings"
)

func NewLogger(prefix string) func(string, ...interface{}) {
	l := log.New(log.Writer(), "["+prefix+"] ", log.Flags())
	return func(format string, args ...interface{}) {
		pc, _, _, _ := runtime.Caller(1)
		fn := runtime.FuncForPC(pc).Name()
		if i := strings.LastIndex(fn, "."); i >= 0 {
			fn = fn[i+1:]
		}
		l.Printf(fn+": "+format, args...)
	}
}

func NewFatalLogger(prefix string) func(string, ...interface{}) {
	l := log.New(log.Writer(), "["+prefix+"] ", log.Flags())
	return func(format string, args ...interface{}) {
		pc, _, _, _ := runtime.Caller(1)
		fn := runtime.FuncForPC(pc).Name()
		if i := strings.LastIndex(fn, "."); i >= 0 {
			fn = fn[i+1:]
		}
		l.Fatalf(fn+": "+format, args...)
	}
}
