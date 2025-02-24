package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
)

// Level чтоб в конфигурации и далее был тем же типом, что у zerolog
type (
	Level = zerolog.Level
)

// CustomLogger обёртка над zerolog.Logger
type CustomLogger struct {
	logger zerolog.Logger
}

const (
	TraceLevel = zerolog.TraceLevel
	DebugLevel = zerolog.DebugLevel
	InfoLevel  = zerolog.InfoLevel
	WarnLevel  = zerolog.WarnLevel
	ErrorLevel = zerolog.ErrorLevel
	FatalLevel = zerolog.FatalLevel
	PanicLevel = zerolog.PanicLevel
	NoLevel    = zerolog.NoLevel
)

// Эти переменные/функции могут пригодиться, чтобы настраивать глобальный уровень логгирования.
// Например, если нужно переопределить уровень логирования во всей программе.
var (
	SetGlobalLevel = zerolog.SetGlobalLevel

	// ConsoleWriter удобный writer для красивого вывода в консоль.
	ConsoleWriter = zerolog.ConsoleWriter{Out: os.Stderr}
)

// NewTextHandler Возвращает "текстовый" (ConsoleWriter) логгер, но можно и напрямую zerolog.New(ConsoleWriter).
func NewTextHandler(w io.Writer) zerolog.Logger {
	return zerolog.New(w).With().Timestamp().Logger()
}

// NewJSONHandler Возвращает "JSON" логгер
func NewJSONHandler(w io.Writer) zerolog.Logger {
	if w == nil {
		w = os.Stderr
	}
	return zerolog.New(w).With().Timestamp().Logger()

}

// NewFileWriter Настраивает `lumberjack.Logger` для ротации файлов
func NewFileWriter(config *LoggerOptions) io.Writer {
	return &lumberjack.Logger{
		Filename:   config.LogFilePath,
		MaxSize:    config.LogFileMaxSizeMB,
		MaxAge:     config.LogFileMaxAgeDays,
		MaxBackups: config.LogFileMaxBackups,
		Compress:   config.LogFileCompress,
	}
}

// Debug(message interface{}, args ...interface{})
func (l *CustomLogger) Debug(message interface{}, args ...interface{}) {
	if len(args) == 0 {
		l.logger.Debug().Msg(fmt.Sprintf("%v", message))
	} else {
		l.logger.Debug().Msgf(fmt.Sprintf("%v", message), args...)
	}
}

// Info(message string, args ...interface{})
func (l *CustomLogger) Info(message string, args ...interface{}) {
	if len(args) == 0 {
		l.logger.Info().Msg(message)
	} else {
		l.logger.Info().Msgf(message, args...)
	}
}

// Warn(message string, args ...interface{})
func (l *CustomLogger) Warn(message string, args ...interface{}) {
	if len(args) == 0 {
		l.logger.Warn().Msg(message)
	} else {
		l.logger.Warn().Msgf(message, args...)
	}
}

// Error(message string, args ...interface{})
func (l *CustomLogger) Error(message string, args ...interface{}) {
	if len(args) == 0 {
		l.logger.Error().Msg(message)
	} else {
		l.logger.Error().Msgf(message, args...)
	}
}

// Fatal(message string, args ...interface{})
func (l *CustomLogger) Fatal(message string, args ...interface{}) {
	if len(args) == 0 {
		l.logger.Fatal().Msg(message)
	} else {
		l.logger.Fatal().Msgf(message, args...)
	}
}
