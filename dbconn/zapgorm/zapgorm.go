// Package zapgorm is an integrator for using zap logger with gorm v2
package zapgorm

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/logger"
)

// Logger is the bridge type for integrating zap to gorm.
type Logger struct {
	level *zap.AtomicLevel
	zap   *zap.Logger
	trace bool
}

// New creates and returns a new loggger
func New(logger *zap.Logger, trace bool) Logger {
	return Logger{zap: logger, trace: trace}
}

// LogMode sets the log level to zap logger and returns the instance
func (l Logger) LogMode(level logger.LogLevel) logger.Interface {
	l.level.SetLevel(zapLevel(level))
	return l
}

// Info print info
func (l Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.zap.Info(msg, convertFields(data)...)
}

// Warn print warn messages
func (l Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.zap.Warn(msg, convertFields(data)...)
}

// Error print error messages
func (l Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.zap.Error(msg, convertFields(data)...)
}

// Trace print sql message
func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if !l.trace {
		return
	}
	query, rowsAffected := fc()
	end := time.Now()
	l.zap.Debug("Trace",
		zap.String("query", query),
		zap.Int64("rowsAffected", rowsAffected),
		// zap.Time("begin", begin),
		// zap.Time("end", end),
		zap.Duration("duration", end.Sub(begin)),
		zap.Error(err),
	)
}

func convertFields(values []interface{}) []zap.Field {
	if len(values) < 2 {
		return []zap.Field{}
	}
	switch values[0] {
	case "sql":
		return []zap.Field{
			zap.String("query", values[3].(string)),
			zap.Any("values", values[4]),
			zap.Duration("duration", values[2].(time.Duration)),
			zap.Int64("affected-rows", values[5].(int64)),
			zap.String("source", values[1].(string)), // if AddCallerSkip(6) is well defined, we can safely remove this field
		}
	default:
		return []zap.Field{
			zap.Any("values", values[2:]),
			zap.Any("source", values[1]), // if AddCallerSkip(6) is well defined, we can safely remove this field
		}
	}
}

func zapLevel(l logger.LogLevel) zapcore.Level {
	switch l {
	case logger.Silent:
		return zapcore.FatalLevel
	case logger.Info:
		return zapcore.InfoLevel
	case logger.Warn:
		return zapcore.WarnLevel
	case logger.Error:
		return zapcore.ErrorLevel
	default:
		return zapcore.FatalLevel
	}
}
func gormLevel(l zapcore.Level) logger.LogLevel {
	switch l {
	case zapcore.FatalLevel:
		return logger.Silent
	case zapcore.InfoLevel:
		return logger.Info
	case zapcore.WarnLevel:
		return logger.Warn
	case zapcore.ErrorLevel:
		return logger.Error
	default:
		return logger.Info
	}
}
