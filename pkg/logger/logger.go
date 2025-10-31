package logger

import "C"
import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// Config 日志配置
type Config struct {
	ReportCaller bool         `mapstructure:"report_caller" validate:"omitempty"`
	PrettyPrint  bool         `mapstructure:"pretty_print" validate:"omitempty"`
	Level        logrus.Level `mapstructure:"level" validate:"required,min=0,max=6"`
	FilePath     string       `mapstructure:"file_path" validate:"required_if=EnableFile true"`
	EnableFile   bool         `mapstructure:"enable_file" validate:"required"`
}

// Init 初始化 logrus
func Init(c Config) {
	logrus.SetLevel(c.Level)
	logrus.SetReportCaller(c.ReportCaller)

	formatter := &logrus.JSONFormatter{
		TimestampFormat:   "2006-01-02T15:04:05.000Z07:00",
		DisableHTMLEscape: true,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "msg",
			logrus.FieldKeyFunc:  "caller",
		},
		CallerPrettyfier: func(frame *runtime.Frame) (function, file string) {
			return "", fmt.Sprintf("%s:%d", filepath.Base(frame.File), frame.Line)
		},
		PrettyPrint: c.PrettyPrint,
	}
	logrus.SetFormatter(formatter)

	// 输出到控制台 + 文件
	if c.EnableFile {
		dir := filepath.Dir(c.FilePath)
		fmt.Println(dir)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
				log.Fatalf("创建日志目录失败: %v", mkErr)
			}
		}
		logrus.SetOutput(io.MultiWriter(os.Stdout, &lumberjack.Logger{
			Filename:   c.FilePath,
			MaxSize:    500,
			MaxAge:     7,
			MaxBackups: 30,
			Compress:   true,
		}))
	} else {
		logrus.SetOutput(os.Stdout)
	}
}
