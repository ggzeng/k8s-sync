package logger

import (
	"os"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zapgrpc"
)

var log *zap.SugaredLogger

var levelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

func GetGrpcLogger() *zapgrpc.Logger {
	return zapgrpc.NewLogger(log.Desugar())
}

// Initialize init after viper init
func Initialize() {
	var syncWriters []zapcore.WriteSyncer
	var encoder zapcore.Encoder
	level := getLoggerLevel(viper.GetString("log.level"))
	fileConfig := &lumberjack.Logger{
		Filename:   viper.GetString("log.path"),    // 日志文件名
		MaxSize:    viper.GetInt("log.maxsize"),    // 日志文件大小
		MaxAge:     viper.GetInt("log.maxAge"),     // 最长保存天数
		MaxBackups: viper.GetInt("log.maxBackups"), // 最多备份几个
		LocalTime:  viper.GetBool("log.localtime"), // 日志时间戳
		Compress:   viper.GetBool("log.compress"),  // 是否压缩文件，使用gzip
	}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000000"))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	if viper.GetBool("log.jsonFormat") {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}
	if viper.GetBool("log.consoleStdout") {
		syncWriters = append(syncWriters, zapcore.AddSync(os.Stdout))
	}
	if viper.GetBool("log.fileStdout") {
		syncWriters = append(syncWriters, zapcore.AddSync(fileConfig))
	}
	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(syncWriters...),
		zap.NewAtomicLevelAt(level))
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	log = logger.Sugar()
}

func getLoggerLevel(lvl string) zapcore.Level {
	if level, ok := levelMap[lvl]; ok {
		return level
	}
	return zapcore.InfoLevel
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Warn(args ...interface{}) {
	log.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func DPanic(args ...interface{}) {
	log.DPanic(args...)
}

func DPanicf(format string, args ...interface{}) {
	log.DPanicf(format, args...)
}

func Panic(args ...interface{}) {
	log.Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	log.Panicf(format, args...)
}

func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}
