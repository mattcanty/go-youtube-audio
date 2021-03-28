package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/denisbrodbeck/machineid"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

type Field struct {
	Key   string
	Value interface{}
}

const (
	contextLogTag string = "context"
	errorLogTag   string = "error"
	deviceLogTag  string = "deviceID"
)

var logEntry *logrus.Entry

func init() {
	level, err := logrus.ParseLevel("DEBUG")
	if err != nil {
		log.Fatalf(err.Error())
	}

	logger = &logrus.Logger{
		Out:   os.Stdout,
		Level: level,
	}
	logger.Formatter = &logrus.TextFormatter{}

	id, err := machineid.ID()
	if err != nil {
		log.Fatalf(err.Error())
	}

	logEntry = logger.WithFields(logrus.Fields{
		deviceLogTag: id,
	})
}

func Error(errMessage string, err error, fields ...*Field) {
	withFields(logEntry, fields...)
	logEntry.
		WithField(contextLogTag, getCallerInfo()).
		WithField(errorLogTag, err).
		Error(errMessage)
}

func Fatal(errMessage string, err error, fields ...*Field) {
	withFields(logEntry, fields...)
	logEntry.
		WithField(contextLogTag, getCallerInfo()).
		WithField(errorLogTag, err).
		Fatal(errMessage)
}

func Info(msg string, fields ...*Field) {
	withFields(logEntry, fields...)
	logEntry.WithField(contextLogTag, getCallerInfo()).Info(msg)
}

func Debug(msg string, fields ...*Field) {
	withFields(logEntry, fields...)
	logEntry.WithField(contextLogTag, getCallerInfo()).Debug(msg)
}

func Warn(msg string, fields ...*Field) {
	withFields(logEntry, fields...)
	logEntry.Warn(msg)
}

func getCallerInfo() string {
	_, filePath, lineNo, isOk := runtime.Caller(2)
	if isOk {
		pathArray := strings.Split(filePath, "/")
		fileName := pathArray[len(pathArray)-1]
		return fmt.Sprintf("%s#%d", fileName, lineNo)
	} else {
		return ""
	}
}

func withFields(logEntry *logrus.Entry, fields ...*Field) {
	if fields != nil {
		for _, field := range fields {
			logEntry = logEntry.WithField(field.Key, field.Value)
		}
	}
}
