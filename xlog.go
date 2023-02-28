package xlog

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"time"
)

var textLogger *zap.SugaredLogger
var jsonLogger *zap.SugaredLogger

func zapLogLevel(level string) zap.AtomicLevel {

	switch level {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	}
	// default is info level
	return zap.NewAtomicLevelAt(zap.InfoLevel)
}

// Option 是日志配置选项
type Option struct {
	Name            string // 代表logger
	NoFile          bool   // 是否为开发模式，如果是true，只显示到标准输出，同旧的 NoFile
	Format          string
	WritableStack   bool // 是否需要打印error及以上级别的堆栈信息
	Skip            int
	WritableCaller  bool // 是否需要打印行号函数信息
	Level           string
	Path            string
	FileName        string
	PackageLevel    map[string]string // 包级别日志等级设置
	ErrLogLevel     string            // 错误日志级别，默认error
	ErrorPath       string
	MaxAge          int64 // 保留旧文件的最大天数，默认7天
	MaxBackups      uint  // 保留旧文件的最大个数，默认7个
	MaxSize         int64 // 在进行切割之前，日志文件的最大大小（以MB为单位）默认1024
	Compress        bool  // 是否压缩/归档旧文件
	packageLogLevel map[string]zapcore.Level
}

const (
	formatTXT         = "txt"
	formatJSON        = "json"
	formatBlank       = "blank"
	timeFormat        = "2006-01-02 15:04:05.999"
	DefaultLoggerName = "xlog"
)

func DefaultOption() *Option {
	return &Option{
		Name:           DefaultLoggerName,
		NoFile:         false,
		Format:         formatTXT,
		WritableCaller: true,
		Skip:           1,
		WritableStack:  false,
		Level:          zap.InfoLevel.String(),
		Path:           "./logs",
		FileName:       "xlog",
		PackageLevel:   make(map[string]string),
		ErrLogLevel:    zap.ErrorLevel.String(),
		ErrorPath:      "./logs",
		MaxAge:         1,
		MaxBackups:     3,
		MaxSize:        1024 * 1024 * 10,
		Compress:       false,
	}
}

func InitLevel(opt *Option) {
	cfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "file",
		FunctionKey:    "func",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(timeFormat),
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
	consolecft := cfg
	if !opt.NoFile {
		cfg.EncodeLevel = zapcore.CapitalLevelEncoder
		consolecft.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 打印所有级别的日志
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapLogLevel(opt.Level).Level()
	})
	consoleDebugging := zapcore.Lock(os.Stdout)
	jsonInfoWriter, _ := newRotateLog(opt, opt.Path, ".json.log")
	jsonEncoderZap := NewJSONEncoder(opt, cfg)

	jsonCore := zapcore.NewTee(
		zapcore.NewCore(jsonEncoderZap, zapcore.AddSync(jsonInfoWriter), lowPriority),
		zapcore.NewCore(jsonEncoderZap, consoleDebugging, lowPriority),
	)

	zapOption := make([]zap.Option, 0)
	if opt.WritableCaller {
		zapOption = append(zapOption, zap.AddCaller(), zap.AddCallerSkip(opt.Skip))
	}
	if opt.WritableStack {
		zapOption = append(zapOption, zap.AddStacktrace(zapcore.ErrorLevel))
	}
	jsonLog := zap.New(jsonCore, zapOption...)
	jsonLogger = jsonLog.Sugar()

	textEncoder := NewConsoleEncoder(opt, cfg)
	textEncoderConsole := NewConsoleEncoder(opt, consolecft)
	textInfoWriter, _ := newRotateLog(opt, opt.Path, ".log")
	textCore := zapcore.NewTee(
		zapcore.NewCore(textEncoder, zapcore.AddSync(textInfoWriter), lowPriority),
		zapcore.NewCore(textEncoderConsole, consoleDebugging, lowPriority),
	)
	testLog := zap.New(textCore, zapOption...) // 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数, 有点小坑
	textLogger = testLog.Sugar()
}

func newRotateLog(opt *Option, p, suffix string) (io.Writer, error) {

	hook, err := rotatelogs.New(
		p+"/"+opt.FileName+".%Y%m%d%H"+suffix, // 没有使用go风格反人类的format格式
		rotatelogs.WithLinkName(opt.FileName),
		//rotatelogs.WithMaxAge(time.Hour*24*time.Duration(opt.MaxAge)),
		rotatelogs.WithRotationTime(time.Hour*24),
		rotatelogs.WithRotationSize(opt.MaxSize),
		rotatelogs.WithRotationCount(opt.MaxBackups),
	)
	if err != nil {
		panic(err)
	}
	return hook, err
}
func Debug(args ...interface{}) {
	jsonLogger.Debug(args...)
}
func Debugf(template string, args ...interface{}) {
	jsonLogger.Debugf(template, args...)
}
func Info(args ...interface{}) {
	jsonLogger.Info(args...)
}
func Infof(template string, args ...interface{}) {
	jsonLogger.Infof(template, args...)
}

// Warn ...
func Warn(args ...interface{}) {
	jsonLogger.Warn(args...)
}

// Warnf ...
func Warnf(template string, args ...interface{}) {
	jsonLogger.Warnf(template, args...)
}

// Error ...
func Error(args ...interface{}) {
	jsonLogger.Error(args...)
}

// Errorf ...
func Errorf(template string, args ...interface{}) {
	jsonLogger.Errorf(template, args...)
}

// DPanic ...
func DPanic(args ...interface{}) {
	jsonLogger.DPanic(args...)
}

// DPanicf ...
func DPanicf(template string, args ...interface{}) {
	jsonLogger.DPanicf(template, args...)
}

// Panic ...
func Panic(args ...interface{}) {
	jsonLogger.Panic(args...)
}

// Panicf ...
func Panicf(template string, args ...interface{}) {
	jsonLogger.Panicf(template, args...)
}

// Fatal ...
func Fatal(args ...interface{}) {
	jsonLogger.Fatal(args...)
}

// Fatalf ...
func Fatalf(template string, args ...interface{}) {
	jsonLogger.Fatalf(template, args...)
}

// CDebug ....
func CDebug(args ...interface{}) {
	textLogger.Debug(args...)
}

// CDebugf ...
func CDebugf(template string, args ...interface{}) {
	textLogger.Debugf(template, args...)
}

// CInfo ...
func CInfo(args ...interface{}) {
	textLogger.Info(args...)
}

// CInfof ...
func CInfof(template string, args ...interface{}) {
	textLogger.Infof(template, args...)
}

// CWarn ...
func CWarn(args ...interface{}) {
	textLogger.Warn(args...)
}

// CWarnf ...
func CWarnf(template string, args ...interface{}) {
	textLogger.Warnf(template, args...)
}

// CError ...
func CError(args ...interface{}) {
	textLogger.Error(args...)
}

// CErrorf ...
func CErrorf(template string, args ...interface{}) {
	textLogger.Errorf(template, args...)
}

// CDPanic ...
func CDPanic(args ...interface{}) {
	textLogger.DPanic(args...)
}

// CDPanicf ...
func CDPanicf(template string, args ...interface{}) {
	textLogger.DPanicf(template, args...)
}

// CPanic ...
func CPanic(args ...interface{}) {
	textLogger.Panic(args...)
}

// CPanicf ...
func CPanicf(template string, args ...interface{}) {
	textLogger.Panicf(template, args...)
}

// CFatal ...
func CFatal(args ...interface{}) {
	textLogger.Fatal(args...)
}

// CFatalf ...
func CFatalf(template string, args ...interface{}) {
	textLogger.Fatalf(template, args...)
}
func GetTextLogger() *zap.SugaredLogger {
	return textLogger
}
