package log

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

// Init init log
func Init(logPath string, logFileName string, logLevel string) {
	createPathIfNotExist(logPath)

	baseLogPath := path.Join(logPath, logFileName)
	writer, err := rotatelogs.New(
		baseLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(baseLogPath), // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour), // 日志切割时间间隔
	)
	if err != nil {
		log.Errorf("config local file system logger error. %v", err)
	}

	lfHook := lfshook.NewHook(
		lfshook.WriterMap{
			log.TraceLevel: writer,
			log.DebugLevel: writer, // 为不同级别设置不同的输出目的
			log.InfoLevel:  writer,
			log.WarnLevel:  writer,
			log.ErrorLevel: writer,
			log.FatalLevel: writer,
			log.PanicLevel: writer,
		},
		&log.TextFormatter{
			DisableSorting: true,
		})
	log.AddHook(lfHook)
	logLevelValue, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Errorf("parse log level err: %s %v", logLevel, err)
		return
	}
	log.SetLevel(logLevelValue)

	fmter := new(log.TextFormatter)
	fmter.FullTimestamp = true                    // 显示完整时间
	fmter.TimestampFormat = "2006-01-02 15:04:05" // 时间格式
	fmter.DisableTimestamp = false                // 禁止显示时间
	fmter.DisableColors = false                   // 禁止颜色显示
	fmter.ForceColors = true

	log.SetFormatter(fmter)
}

func createPathIfNotExist(path string) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(path, os.ModePerm)
			if err != nil {
				fmt.Printf("create path %s err:%v", path, err)
				return
			}
		}
	}
}

func withOutFileOut() {
	f, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return
	}
	writer := bufio.NewWriter(f)
	log.SetOutput(writer)
}

// Trace logs a message at level Info on the standard logger.
func Trace(args ...interface{}) {
	log.Trace(args...)
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	log.Info(args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...interface{}) {
	log.Error(args...)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(args ...interface{}) {
	log.Panic(args...)
}

// Tracef logs a message at level Trace on the standard logger.
func Tracef(format string, args ...interface{}) {
	log.Tracef(format, args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func Panicf(format string, args ...interface{}) {
	log.Panicf(format, args...)
}
