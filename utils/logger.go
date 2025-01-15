package utils

import (
	"fmt"
	"strings"
	"time"
)

type loggerColors struct {
	Reset   string
	Red     string
	Green   string
	Yellow  string
	Blue    string
	Magenta string
	Cyan    string
	Gray    string
	White   string
}

var Colors = loggerColors{
	Reset:   "\033[0m",
	Red:     "\033[31m",
	Green:   "\033[32m",
	Yellow:  "\033[33m",
	Blue:    "\033[34m",
	Magenta: "\033[35m",
	Cyan:    "\033[36m",
	Gray:    "\033[37m",
	White:   "\033[97m",
}

type Logger struct {
	messages []string
}

var Log *Logger = NewLogger()

func NewLogger() *Logger {
	return &Logger{
		messages: make([]string, 0),
	}
}

func (l *Logger) GetMessages() []string {
	return l.messages
}

func (l *Logger) FormatMessage(prefix string, args ...interface{}) string {
	// Center the prefix and make it exactly 10 characters wide
	padding := (10 - len(prefix)) / 2
	prefix = fmt.Sprintf("%s%s%s", strings.Repeat(" ", padding), prefix, strings.Repeat(" ", 10-len(prefix)-padding))

	msg := fmt.Sprintf("[%s] - %s | ", prefix, time.Now().Format("2006-01-02 15:04:05.000"))
	for _, arg := range args {
		msg += fmt.Sprintf(" %v", arg)
	}

	l.messages = append(l.messages, msg)

	return msg
}

func (l *Logger) Info(args ...interface{}) {
	fmt.Println(Colors.Blue + l.FormatMessage("INFO", args...) + Colors.Reset)
}

func (l *Logger) Error(args ...interface{}) {
	fmt.Println(Colors.Red + l.FormatMessage("ERROR", args...) + Colors.Reset)
}

func (l *Logger) Warn(args ...interface{}) {
	fmt.Println(Colors.Yellow + l.FormatMessage("WARN", args...) + Colors.Reset)
}

func (l *Logger) Log(args ...interface{}) {
	fmt.Println(l.FormatMessage("LOG", args...))
}

func (l *Logger) Success(args ...interface{}) {
	fmt.Println(Colors.Green + l.FormatMessage("SUCCESS", args...) + Colors.Reset)
}

func (l *Logger) Debug(args ...interface{}) {
	fmt.Println(Colors.Gray + l.FormatMessage("DEBUG", args...) + Colors.Reset)
}

func (l *Logger) Custom(color string, prefix string, args ...interface{}) {
	fmt.Println(color + l.FormatMessage(prefix, args...) + Colors.Reset)
}
