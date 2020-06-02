package logger

import (
	"os"
	"proctor/internal/app/service/infra/config"

	log "github.com/sirupsen/logrus"
)

var logger *log.Logger

func Setup() {
	if logger != nil {
		return
	}

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	logLevel, err := log.ParseLevel(config.Config().LogLevel)
	if err != nil {
		log.Panic(err)
	}
	log.SetLevel(logLevel)

	logger = &log.Logger{
		Out:       os.Stdout,
		Formatter: &log.JSONFormatter{},
		Hooks:     make(log.LevelHooks),
		Level:     logLevel,
	}

	return
}

func AddHook(hook log.Hook) {
	logger.Hooks.Add(hook)
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Warn(args ...interface{}) {
	log.Warn(args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func Panic(args ...interface{}) {
	log.Panic(args...)
}

func LogErrors(err error, action string, args ...interface{}) {
	if err != nil {
		Error("Failed to ", action, " with errors ", err, " and data ", args)
	} else {
		Debug("Success to ", action, " with data ", args)
	}
}
