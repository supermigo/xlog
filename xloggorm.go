package xlog

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	gormlog "gorm.io/gorm/logger"
	"time"
)

type GormLogger struct {
	logger *zap.SugaredLogger
	gormlog.Config
}

func (g *GormLogger) LogMode(logLevel gormlog.LogLevel) gormlog.Interface {
	var zapLevel zapcore.Level
	switch logLevel {
	case gormlog.Silent:
		zapLevel = zapcore.DPanicLevel
	case gormlog.Error:
		zapLevel = zapcore.ErrorLevel
	case gormlog.Warn:
		zapLevel = zapcore.WarnLevel
	case gormlog.Info:
		zapLevel = zapcore.InfoLevel
	}
	zapLevel = zapLevel
	//level.SetLevel(zapLevel)
	return g
}

func (g *GormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	//TODO implement me
	g.logger.Infof(s, i...)
}

func (g *GormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	//TODO implement me
	g.logger.Warnf(s, i...)
}

func (g *GormLogger) Error(ctx context.Context, s string, i ...interface{}) {
	//TODO implement me
	g.logger.Errorf(s, i...)
}

func (g *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	//TODO implement me
	if g.logger.Level() >= zapcore.DPanicLevel {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	switch {
	case err != nil && g.logger.Level() <= zapcore.ErrorLevel && (!errors.Is(err, gormlog.ErrRecordNotFound) || !g.IgnoreRecordNotFoundError):
		g.logger.Errorw(sql, "elapsed", elapsed, "rows", rows, "error", err.Error())
	case elapsed > g.SlowThreshold && g.SlowThreshold != 0 && g.logger.Level() >= zapcore.WarnLevel:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", g.SlowThreshold)
		g.logger.Warnw(sql, "elapsed", elapsed, "rows", rows, "warn", slowLog)
	case g.logger.Level() < zapcore.InfoLevel:
		g.logger.Debugw(sql, "elapsed", elapsed, "rows", rows)
	}
}

func NewGormLogger(conf gormlog.Config) gormlog.Interface {
	return &GormLogger{
		logger: textLogger,
		Config: conf,
	}
}
