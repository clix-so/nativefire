package logger

import (
	"fmt"
	"os"
)

type Logger struct {
	verbose bool
}

func New(verbose bool) *Logger {
	return &Logger{verbose: verbose}
}

func (l *Logger) Info(msg string) {
	fmt.Println(msg)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func (l *Logger) Debug(msg string) {
	if l.verbose {
		fmt.Printf("[DEBUG] %s\n", msg)
	}
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.verbose {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

func (l *Logger) Warn(msg string) {
	fmt.Printf("⚠️  %s\n", msg)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	fmt.Printf("⚠️  "+format+"\n", args...)
}

func (l *Logger) Error(msg string) {
	fmt.Fprintf(os.Stderr, "❌ %s\n", msg)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "❌ "+format+"\n", args...)
}

func (l *Logger) Success(msg string) {
	fmt.Printf("✅ %s\n", msg)
}

func (l *Logger) Successf(format string, args ...interface{}) {
	fmt.Printf("✅ "+format+"\n", args...)
}
