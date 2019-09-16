package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

func init() {
	zapLog := zap.NewProductionConfig()
	zapLog.Level.SetLevel(zap.DebugLevel)
	zapLog.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapLog.DisableStacktrace = true
	zapLog.DisableCaller = false
	logger, err := zapLog.Build(zap.AddCallerSkip(1))

	if err != nil {
		panic(err)
	}
	globalLogger = logger

}

func Options(opts ...zap.Option) {
	globalLogger = globalLogger.WithOptions(opts...)
}

func Debug(msg string, fields ...zapcore.Field) {
	globalLogger.Debug(msg, fields...)
}
func Info(msg string, fields ...zapcore.Field) {
	globalLogger.Info(msg, fields...)
}
func Warn(msg string, fields ...zapcore.Field) {
	globalLogger.Warn(msg, fields...)
}
func Error(msg string, fields ...zapcore.Field) {
	globalLogger.Error(msg, fields...)
}
func Fatal(msg string, fields ...zapcore.Field) {
	globalLogger.Panic(msg, fields...)
}

func Debugf(msg string, args ...interface{}) {
	globalLogger.Debug(fmt.Sprintf(msg, args...))
}
func Infof(msg string, args ...interface{}) {
	globalLogger.Info(fmt.Sprintf(msg, args...))
}
func Warnf(msg string, args ...interface{}) {
	globalLogger.Warn(fmt.Sprintf(msg, args...))
}
func Errorf(msg string, args ...interface{}) {
	globalLogger.Error(fmt.Sprintf(msg, args...))
}
func Fatalf(msg string, args ...interface{}) {
	globalLogger.Panic(fmt.Sprintf(msg, args...))
}

var (
	Err        = zap.Error
	Time       = zap.Time
	Any        = zap.Any
	Binary     = zap.Binary
	Bool       = zap.Bool
	ByteString = zap.ByteString
	Duration   = zap.Duration
	Stack      = zap.Stack
	String     = zap.String
	Strings    = zap.Strings
	Stringer   = zap.Stringer
	Uint8      = zap.Uint8
	Uint16     = zap.Uint16
	Uint32     = zap.Uint32
	Uint64     = zap.Uint64
	Uint       = zap.Uint
	Int8       = zap.Int8
	Int16      = zap.Int16
	Int32      = zap.Int32
	Int64      = zap.Int64
	Int        = zap.Int
	Float32    = zap.Float32
	Float64    = zap.Float64
)
