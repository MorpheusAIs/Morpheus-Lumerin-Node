package lib

import (
	"io"
	"os"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ILogger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	DPanic(args ...interface{}) // Development panic - panics only in development mode
	Panic(args ...interface{})
	Fatal(args ...interface{})

	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	DPanicf(template string, args ...interface{})
	Panicf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})

	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	DPanicw(msg string, keysAndValues ...interface{})
	Panicw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})

	Log(lvl zapcore.Level, msg string, fields ...zapcore.Field)

	Sync() error
	Close() error
	Named(name string) ILogger
	With(args ...interface{}) ILogger
}

const timeLayout = "2006-01-02T15:04:05.999999999"

func NewLogger(level string, color, isProd bool, isJSON bool, filepath string) (*Logger, error) {
	log, file, al, err := newLogger(level, color, isProd, isJSON, filepath, nil)
	if err != nil {
		return nil, err
	}

	return &Logger{
		SugaredLogger: log.Sugar(),
		file:          file,
		atomicLevel:   al,
	}, nil
}

func NewLoggerMemory(level string, color, isProd bool, isJSON bool, filepath string, wr io.Writer) (*Logger, error) {
	log, file, al, err := newLogger(level, color, isProd, isJSON, filepath, wr)

	if err != nil {
		return nil, err
	}

	return &Logger{
		SugaredLogger: log.Sugar(),
		file:          file,
		atomicLevel:   al,
	}, nil
}

// NewTestLogger logs only to stdout
func NewTestLogger() *Logger {
	log, file, al, _ := newLogger("debug", false, false, false, "", nil)
	return &Logger{
		SugaredLogger: log.Sugar(),
		file:          file,
		atomicLevel:   al,
	}
}

func newLogger(levelStr string, color bool, isProd bool, isJSON bool, filepath string, extraWriter io.Writer) (*zap.Logger, *os.File, zap.AtomicLevel, error) {
	level, err := zapcore.ParseLevel(levelStr)
	if err != nil {
		return nil, nil, zap.AtomicLevel{}, err
	}

	atomicLevel := zap.NewAtomicLevelAt(level)

	var cores []zapcore.Core
	var file *os.File

	if filepath != "" {
		fileCore, fd, err := newFileCore(atomicLevel, isProd, isJSON, filepath)
		if err != nil {
			return nil, nil, zap.AtomicLevel{}, err
		}
		file = fd
		cores = append(cores, fileCore)
	}
	if extraWriter != nil {
		memoryCore := zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), zapcore.AddSync(extraWriter), atomicLevel)
		cores = append(cores, memoryCore)
	}

	consoleCore := newConsoleCore(atomicLevel, color, isProd, isJSON)
	cores = append(cores, consoleCore)

	var core zapcore.Core
	if len(cores) > 1 {
		core = zapcore.NewTee(cores...)
	} else {
		core = cores[0]
	}

	opts := []zap.Option{
		zap.AddStacktrace(zap.ErrorLevel),
	}
	if !isProd {
		opts = append(opts, zap.Development())
	}

	return zap.New(core, opts...), file, atomicLevel, nil
}

func newConsoleCore(level zapcore.LevelEnabler, color bool, isProd bool, isJSON bool) zapcore.Core {
	encoderCfg := newEncoderCfg(isProd, color, isJSON)

	var encoder zapcore.Encoder
	if isJSON {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}
	return zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
}

func newEncoderCfg(isProd bool, color bool, isJSON bool) zapcore.EncoderConfig {
	var encoderCfg zapcore.EncoderConfig
	if isProd {
		encoderCfg = zap.NewProductionEncoderConfig()
	} else {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
		encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout(timeLayout)
	}

	if color && !isJSON {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	return encoderCfg
}

func newFileCore(level zapcore.LevelEnabler, isProd bool, isJSON bool, path string) (zapcore.Core, *os.File, error) {
	encoderCfg := newEncoderCfg(isProd, false, isJSON)
	if !isJSON {
		encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout(timeLayout)
	}

	var encoder zapcore.Encoder
	if isJSON {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return nil, nil, err
	}

	return zapcore.NewCore(encoder, zapcore.AddSync(file), level), file, nil
}

type Logger struct {
	*zap.SugaredLogger
	file        *os.File
	atomicLevel zap.AtomicLevel
}

// SetLevel changes the log level at runtime for all cores sharing this logger's AtomicLevel.
func (l *Logger) SetLevel(levelStr string) error {
	lvl, err := zapcore.ParseLevel(levelStr)
	if err != nil {
		return err
	}
	l.atomicLevel.SetLevel(lvl)
	return nil
}

// GetLevel returns the current log level string.
func (l *Logger) GetLevel() string {
	return l.atomicLevel.Level().String()
}

func (l *Logger) Named(name string) ILogger {
	return &Logger{
		SugaredLogger: l.SugaredLogger.Named(name),
		atomicLevel:   l.atomicLevel,
	}
}

func (l *Logger) With(args ...interface{}) ILogger {
	return &Logger{
		SugaredLogger: l.SugaredLogger.With(args...),
		atomicLevel:   l.atomicLevel,
	}
}

func (l *Logger) Log(lvl zapcore.Level, msg string, fields ...zapcore.Field) {
	l2 := l.SugaredLogger.Desugar()
	l2.Log(lvl, msg, fields...)
}

func (l *Logger) Close() error {
	return multierr.Combine(
		l.Sync(),
		l.file.Close(),
	)
}
