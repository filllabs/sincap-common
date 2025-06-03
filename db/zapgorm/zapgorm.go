// Package zapgorm provides zap logger integration for sqlx database operations
package zapgorm

import (
	"context"
	"database/sql/driver"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a zap-based logger for sqlx operations
type Logger struct {
	zap   *zap.Logger
	trace bool
}

// New creates and returns a new logger for sqlx operations
func New(logger *zap.Logger, trace bool) *Logger {
	return &Logger{zap: logger, trace: trace}
}

// LogQuery logs SQL query execution
func (l *Logger) LogQuery(ctx context.Context, query string, args []interface{}, duration time.Duration, err error) {
	if !l.trace {
		return
	}

	fields := []zap.Field{
		zap.String("query", query),
		zap.Any("args", args),
		zap.Duration("duration", duration),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.zap.Error("SQL query failed", fields...)
	} else {
		l.zap.Debug("SQL query executed", fields...)
	}
}

// LogExec logs SQL execution (INSERT, UPDATE, DELETE)
func (l *Logger) LogExec(ctx context.Context, query string, args []interface{}, result driver.Result, duration time.Duration, err error) {
	if !l.trace {
		return
	}

	fields := []zap.Field{
		zap.String("query", query),
		zap.Any("args", args),
		zap.Duration("duration", duration),
	}

	if result != nil {
		if rowsAffected, raErr := result.RowsAffected(); raErr == nil {
			fields = append(fields, zap.Int64("rows_affected", rowsAffected))
		}
		if lastInsertId, liErr := result.LastInsertId(); liErr == nil {
			fields = append(fields, zap.Int64("last_insert_id", lastInsertId))
		}
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.zap.Error("SQL execution failed", fields...)
	} else {
		l.zap.Debug("SQL execution completed", fields...)
	}
}

// Info logs info messages
func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

// Warn logs warning messages
func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// Error logs error messages
func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

// Debug logs debug messages
func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level zapcore.Level) {
	// Note: This is a simplified version. In a real implementation,
	// you might want to use zap.AtomicLevel for dynamic level changes
}

// WithTrace enables or disables SQL tracing
func (l *Logger) WithTrace(trace bool) *Logger {
	return &Logger{
		zap:   l.zap,
		trace: trace,
	}
}
