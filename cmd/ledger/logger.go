package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/newrelic/go-agent/v3/integrations/logcontext/nrlogrusplugin"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func buildLogger() {

	now := time.Now().Unix()

	level, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		log.Panicf(fmt.Sprintf("failed to configure log level: %s", err))
	}

	logger = logrus.New()

	logger.SetOutput(ioutil.Discard)

	logger.AddHook(&writerHook{
		Writer:    os.Stdout,
		LogLevels: logrus.AllLevels,
	})

	logger.AddHook(&writerHook{
		Writer: &lumberjack.Logger{
			Filename: fmt.Sprintf("logs/%d-error.log", now),
			MaxSize:  10,
			Compress: false,
		},
		LogLevels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		},
	})

	logger.AddHook(&writerHook{
		Writer: &lumberjack.Logger{
			Filename:   fmt.Sprintf("logs/%d-info.log", now),
			MaxBackups: 3,
			MaxSize:    10,
			Compress:   false,
		},
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
		},
	})

	logger.SetLevel(level)
	// logger.SetFormatter(&logrus.TextFormatter{
	// 	FullTimestamp: true,
	// })
	logrus.SetFormatter(nrlogrusplugin.ContextFormatter{})

	// if cfg.Env == "production" {
	// }

}

type writerHook struct {
	Writer    io.Writer
	LogLevels []logrus.Level
}

func (w *writerHook) Fire(entry *logrus.Entry) error {

	if txn := newrelic.FromContext(entry.Context); txn != nil && entry.Level < logrus.InfoLevel {

		data := entry.Data

		msg := entry.Message

		nre := newrelic.Error{}
		nre.Stack = newrelic.NewStackTrace()

		if err, ok := data["error"]; ok {
			msg = fmt.Sprintf("%s: %v", msg, err)
			delete(data, "error")
		}

		if service, ok := data["service"]; ok {
			nre.Class = service.(string)
		}

		nre.Message = msg
		nre.Attributes = data

		txn.NoticeError(nre)

	}

	line, err := entry.String()
	if err != nil {
		return err
	}

	_, err = w.Writer.Write([]byte(line))
	return err

}

func (w *writerHook) Levels() []logrus.Level {
	return w.LogLevels
}
