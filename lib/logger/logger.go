package logger

import (
	"context"

	"github.com/leonardovalentini/crypto-coins/lib/contextKey"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:generate mockgen -destination=../tests/mocks/logger/mockLogger.go -package=logger github.com/leonardovalentini/crypto-coins/lib/logger LoggerI
type LoggerI interface {
	SetCtx(ctx context.Context)
	Info(message string, fields ...zap.Field)
	Fatal(message string, fields ...zap.Field)
	Debug(message string, fields ...zap.Field)
	Error(message string, fields ...zap.Field)
}

type Logger struct {
	ctx context.Context
}

var Log *zap.Logger

func init() {
	var err error

	config := zap.NewProductionConfig()

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.StacktraceKey = ""
	config.EncoderConfig = encoderConfig

	Log, err = config.Build(zap.AddCallerSkip(1))

	if err != nil {
		panic(err)
	}
}

func NewLogger(ctx context.Context) *Logger {
	return &Logger{ctx: ctx}
}

func (l *Logger) SetCtx(ctx context.Context) {
	l.ctx = ctx
}

func (l *Logger) Info(message string, fields ...zap.Field) {
	fields = append(fields, GetZapField(l.ctx)...)
	Log.Info(message, fields...)
}

func Info(message string, fields ...zap.Field) {
	Log.Info(message, fields...)
}

func (l *Logger) Fatal(message string, fields ...zap.Field) {
	fields = append(fields, GetZapField(l.ctx)...)
	Log.Fatal(message, fields...)
}

func Fatal(message string, fields ...zap.Field) {
	Log.Fatal(message, fields...)
}

func (l *Logger) Debug(message string, fields ...zap.Field) {
	fields = append(fields, GetZapField(l.ctx)...)
	Log.Debug(message, fields...)
}

func Debug(message string, fields ...zap.Field) {
	Log.Debug(message, fields...)
}

func (l *Logger) Error(message string, fields ...zap.Field) {
	fields = append(fields, GetZapField(l.ctx)...)
	Log.Error(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	Log.Error(message, fields...)
}

func GetZapField(ctx context.Context) []zap.Field {
	fields := []zap.Field{}
	if id, ok := contextKey.GetRequestId(ctx); ok {
		fields = append(fields, zap.String("request_id", id))
	}

	if jobId, ok := contextKey.GetJobId(ctx); ok {
		fields = append(fields, zap.String("job_id", jobId))
	}

	return fields
}
