package logger

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestInit_Console(t *testing.T) {
	config := Config{
		ReportCaller: false,
		PrettyPrint:  false,
		Level:        logrus.InfoLevel,
		EnableFile:   false,
	}

	Init(config)

	// Verify logger level
	assert.Equal(t, logrus.InfoLevel, logrus.GetLevel())

	// Capture log output AFTER Init
	var buf bytes.Buffer
	logrus.SetOutput(&buf)

	// Test log output
	logrus.Info("test message")
	assert.Contains(t, buf.String(), "test message")
}

func TestInit_WithFile(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "logs", "test.log")

	config := Config{
		ReportCaller: false,
		PrettyPrint:  false,
		Level:        logrus.DebugLevel,
		FilePath:     logFile,
		EnableFile:   true,
	}

	Init(config)

	// Verify logger level
	assert.Equal(t, logrus.DebugLevel, logrus.GetLevel())

	// Write log
	logrus.Info("file test message")

	// Verify file exists
	_, err := os.Stat(logFile)
	assert.NoError(t, err, "log file should be created")
}

func TestInit_WithReportCaller(t *testing.T) {
	config := Config{
		ReportCaller: true,
		PrettyPrint:  false,
		Level:        logrus.WarnLevel,
		EnableFile:   false,
	}

	Init(config)

	// Capture output AFTER Init
	var buf bytes.Buffer
	logrus.SetOutput(&buf)

	// Test that caller is reported
	logrus.Warn("test warning")
	output := buf.String()

	assert.Contains(t, output, "test warning")
	// When ReportCaller is true, the log should contain file info
	assert.Contains(t, output, "logger_test.go")
}

func TestInit_WithPrettyPrint(t *testing.T) {
	config := Config{
		ReportCaller: false,
		PrettyPrint:  true,
		Level:        logrus.InfoLevel,
		EnableFile:   false,
	}

	Init(config)

	var buf bytes.Buffer
	logrus.SetOutput(&buf)

	logrus.Info("pretty print test")
	output := buf.String()

	assert.Contains(t, output, "pretty print test")
}

func TestInit_DifferentLevels(t *testing.T) {
	levels := []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	}

	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			config := Config{
				ReportCaller: false,
				PrettyPrint:  false,
				Level:        level,
				EnableFile:   false,
			}

			Init(config)

			assert.Equal(t, level, logrus.GetLevel())
		})
	}
}

func TestInit_CreateDirectory(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "deep", "nested", "dir", "test.log")

	config := Config{
		ReportCaller: false,
		PrettyPrint:  false,
		Level:        logrus.InfoLevel,
		FilePath:     logFile,
		EnableFile:   true,
	}

	Init(config)

	// Verify directory was created
	dir := filepath.Dir(logFile)
	info, err := os.Stat(dir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestInit_JSONFormatter(t *testing.T) {
	config := Config{
		ReportCaller: false,
		PrettyPrint:  false,
		Level:        logrus.InfoLevel,
		EnableFile:   false,
	}

	Init(config)

	var buf bytes.Buffer
	logrus.SetOutput(&buf)

	logrus.WithFields(logrus.Fields{
		"key1": "value1",
		"key2": 123,
	}).Info("json test")

	output := buf.String()

	// Should contain JSON formatted output
	assert.Contains(t, output, "\"timestamp\":")
	assert.Contains(t, output, "\"level\":")
	assert.Contains(t, output, "\"msg\":")
	assert.Contains(t, output, "json test")
}
