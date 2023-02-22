package log

import (
	"io"
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zapLog

type zapLog struct {
	log *zap.Logger
}

type LogConfig struct {
	Level        string // 日志级别 debug, info, warn, error
	Path         string // 日志输出目录，为空则打印在控制台
	MaxAge       int32  // 历史日志文件保留天数
	rotationTime int32  // 日志切割时间间隔，单位天
	Prod         bool   // 是否生产环境
}

type Option func(option *LogConfig)

// InitLogger return LogConfig
func InitLogger(options ...Option) *zap.Logger {
	cnf := &LogConfig{
		Level:        "info",
		Path:         "./logs/log_out.log",
		MaxAge:       7,
		rotationTime: 1,
		Prod:         false,
	}

	for _, fn := range options {
		fn(cnf)
	}

	var zapLevel zapcore.Level
	switch cnf.Level {
	case "debug":
		zapLevel = zap.DebugLevel
	case "info":
		zapLevel = zap.InfoLevel
	case "error":
		zapLevel = zap.ErrorLevel
	default:
		zapLevel = zap.WarnLevel
	}

	writer := getWriter(cnf.Path, cnf.MaxAge, cnf.rotationTime)

	// 可按日志等级分割日志
	//core := zapcore.NewTee(
	//	zapcore.NewCore(getEncoder(), zapcore.NewMultiWriteSyncer(zapcore.AddSync(writer), zapcore.AddSync(os.Stdout)), zap.InfoLevel),
	//	zapcore.NewCore(getEncoder(), zapcore.NewMultiWriteSyncer(zapcore.AddSync(writer), zapcore.AddSync(os.Stdout)), zapLevel),
	//)

	core := zapcore.NewCore(
		getEncoder(),
		zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(writer),
			zapcore.AddSync(os.Stdout),
		), zapLevel)

	zapLogger := zap.New(core, getOptions(cnf.Prod)...)
	logger = &zapLog{log: zapLogger}
	return zapLogger
}

// SetLevel set zapLog level
func SetLevel(Level string) Option {
	return func(option *LogConfig) {
		option.Level = Level
	}
}

// SetLogPath set zapLog path
func SetLogPath(path string) Option {
	return func(option *LogConfig) {
		option.Path = path
	}
}

// SetMaxAge set zapLog max age
func SetMaxAge(day int32) Option {
	return func(option *LogConfig) {
		option.MaxAge = day
	}
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeName = zapcore.FullNameEncoder
	encoderConfig.LineEnding = zapcore.DefaultLineEnding
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.StacktraceKey = "stacktrace"
	encoderConfig.LevelKey = "level"
	encoderConfig.TimeKey = "time"
	encoderConfig.MessageKey = "msg"
	return zapcore.NewJSONEncoder(encoderConfig)
}

func getOptions(prod bool) []zap.Option {
	opts := make([]zap.Option, 0)
	if !prod {
		opts = append(opts, zap.Development())
	}
	return opts
}

func getWriter(filename string, age, rotationTime int32) io.Writer {
	writer, err := rotatelogs.New(
		filename+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Hour*time.Duration(age*24)),
		rotatelogs.WithRotationTime(time.Hour*time.Duration(rotationTime*24)),
	)

	if err != nil {
		panic(err)
	}
	return writer
}
