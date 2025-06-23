package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger определяет методы логирования
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

// LogOutput определяет тип вывода логов
type LogOutput string

const (
	OutputConsole LogOutput = "console"
	OutputFile    LogOutput = "file"
	OutputBoth    LogOutput = "both"
)

type logger struct {
	level    string
	format   string
	output   LogOutput
	filePath string
	log      *log.Logger
	file     *os.File
}

// Config содержит настройки логгера
type Config struct {
	Level    string
	Format   string
	Output   LogOutput
	FilePath string
}

// New создает новый экземпляр логгера
func New(level, format string) Logger {
	return NewWithConfig(Config{
		Level:  level,
		Format: format,
		Output: OutputConsole,
	})
}

// NewWithConfig создает логгер с расширенной конфигурацией
func NewWithConfig(cfg Config) Logger {
	l := &logger{
		level:  cfg.Level,
		format: cfg.Format,
		output: cfg.Output,
	}

	var writers []io.Writer

	// Консольный вывод
	if cfg.Output == OutputConsole || cfg.Output == OutputBoth {
		writers = append(writers, os.Stdout)
	}

	// Файловый вывод
	if cfg.Output == OutputFile || cfg.Output == OutputBoth {
		if cfg.FilePath == "" {
			cfg.FilePath = filepath.Join("logs", fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02")))
		}

		// Создаем директорию если не существует
		dir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Failed to create log directory: %v", err)
		} else {
			file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				log.Printf("Failed to open log file: %v", err)
			} else {
				l.file = file
				writers = append(writers, file)
			}
		}
	}

	// Создаем multi-writer
	var output io.Writer
	if len(writers) == 1 {
		output = writers[0]
	} else if len(writers) > 1 {
		output = io.MultiWriter(writers...)
	} else {
		output = os.Stdout
	}

	l.log = log.New(output, "", log.LstdFlags|log.Lshortfile)
	return l
}

func (l *logger) shouldLog(level string) bool {
	levels := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
		"fatal": 4,
	}

	currentLevel, ok := levels[l.level]
	if !ok {
		currentLevel = 0
	}

	targetLevel, ok := levels[level]
	if !ok {
		return true
	}

	return targetLevel >= currentLevel
}

// logMessage форматирует и выводит сообщение в зависимости от формата
func (l *logger) logMessage(level string, message string) {
	if l.format == "json" {
		logEntry := map[string]interface{}{
			"time":  time.Now().Format(time.RFC3339),
			"level": level,
			"msg":   message,
		}
		if data, err := json.Marshal(logEntry); err == nil {
			l.log.Println(string(data))
		} else {
			l.log.Printf("[%s] %s", level, message)
		}
	} else {
		l.log.Printf("[%s] %s", level, message)
	}
}

func (l *logger) Debug(args ...interface{}) {
	if l.shouldLog("debug") {
		l.logMessage("DEBUG", fmt.Sprint(args...))
	}
}

func (l *logger) Debugf(format string, args ...interface{}) {
	if l.shouldLog("debug") {
		l.logMessage("DEBUG", fmt.Sprintf(format, args...))
	}
}

func (l *logger) Info(args ...interface{}) {
	if l.shouldLog("info") {
		l.logMessage("INFO", fmt.Sprint(args...))
	}
}

func (l *logger) Infof(format string, args ...interface{}) {
	if l.shouldLog("info") {
		l.logMessage("INFO", fmt.Sprintf(format, args...))
	}
}

func (l *logger) Warn(args ...interface{}) {
	if l.shouldLog("warn") {
		l.logMessage("WARN", fmt.Sprint(args...))
	}
}

func (l *logger) Warnf(format string, args ...interface{}) {
	if l.shouldLog("warn") {
		l.logMessage("WARN", fmt.Sprintf(format, args...))
	}
}

func (l *logger) Error(args ...interface{}) {
	if l.shouldLog("error") {
		l.logMessage("ERROR", fmt.Sprint(args...))
	}
}

func (l *logger) Errorf(format string, args ...interface{}) {
	if l.shouldLog("error") {
		l.logMessage("ERROR", fmt.Sprintf(format, args...))
	}
}

func (l *logger) Fatal(args ...interface{}) {
	l.logMessage("FATAL", fmt.Sprint(args...))
	os.Exit(1)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.logMessage("FATAL", fmt.Sprintf(format, args...))
	os.Exit(1)
}
 