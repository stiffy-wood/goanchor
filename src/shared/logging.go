package shared

import (
	"fmt"
	"io"
	"os"
)

const (
	debugLevel   = "DEBUG"
	infoLevel    = "INFO"
	warningLevel = "WARNING"
	errorLevel   = "ERROR"
)

type Logger struct {
	out       io.Writer
	context   string
	variables []string
}

func NewLogger(context string) *Logger {
	return NewLoggerWithVariables(context)
}

func NewLoggerWithVariables(context string, vars ...string) *Logger {
	return &Logger{
		out:       os.Stdout,
		context:   context,
		variables: vars,
	}
}

func (l *Logger) varsToString(vars []string) string {
	str := ""
	for _, v := range vars {
		str += fmt.Sprintf("[%s]", v)
	}
	return str
}

func (l *Logger) writeLog(fullMsg string) {
	_, err := l.out.Write([]byte(fullMsg))
	if err != nil {
		panic(err)
	}
}

func (l *Logger) formatLog(level, msg, contextVars, logVars string) string {
	return fmt.Sprintf("[%s] %s %s - %s %s\n", level, l.context, contextVars, msg, logVars)
}

func (l *Logger) log(level, msg string, vars []string) {
	l.writeLog(
		l.formatLog(
			level,
			msg,
			l.varsToString(l.variables),
			l.varsToString(vars),
		),
	)
}

func (l *Logger) Debug(msg string, vars ...string) {
	l.log(infoLevel, msg, vars)
}

func (l *Logger) Info(msg string, vars ...string) {
	l.log(debugLevel, msg, vars)
}

func (l *Logger) Warn(msg string, vars ...string) {
	l.log(warningLevel, msg, vars)
}

func (l *Logger) Error(msg string, vars ...string) {
	l.log(errorLevel, msg, vars)
}
