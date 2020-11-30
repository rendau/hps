package zap

import (
	"log"

	"go.uber.org/zap"
)

const callerSkip = 1

type St struct {
	l  *zap.Logger
	sl *zap.SugaredLogger
}

func New(level string, debug, test bool) (*St, error) {
	var err error

	logger := &St{}

	switch {
	case test:
		logger.l = zap.NewExample(zap.AddCallerSkip(callerSkip))
	case debug:
		logger.l, err = zap.NewDevelopment(zap.AddCallerSkip(callerSkip))
		if err != nil {
			return nil, err
		}
	default:
		cfg := zap.NewProductionConfig()

		switch level {
		case "error":
			cfg.Level.SetLevel(zap.ErrorLevel)
		case "warn": // default
			cfg.Level.SetLevel(zap.WarnLevel)
		case "info":
			cfg.Level.SetLevel(zap.InfoLevel)
		case "debug":
			cfg.Level.SetLevel(zap.DebugLevel)
		default:
			cfg.Level.SetLevel(zap.WarnLevel)
		}

		logger.l, err = cfg.Build(zap.AddCallerSkip(callerSkip))
		if err != nil {
			return nil, err
		}
	}

	logger.sl = logger.l.Sugar()

	return logger, nil
}

// Fatal is for Fatal
func (lg *St) Fatal(args ...interface{}) {
	lg.sl.Fatal(args...)
}

// Fatalf is for Fatalf
func (lg *St) Fatalf(tmpl string, args ...interface{}) {
	lg.sl.Fatalf(tmpl, args...)
}

// Fatalw is for Fatalw
func (lg *St) Fatalw(msg string, err interface{}, args ...interface{}) {
	kvs := make([]interface{}, 0, len(args)+2)
	kvs = append(kvs, "error", err)
	kvs = append(kvs, args...)
	lg.sl.Fatalw(msg, kvs...)
}

// Error is for Error
func (lg *St) Error(args ...interface{}) {
	lg.sl.Error(args...)
}

// Errorf is for Errorf
func (lg *St) Errorf(tmpl string, args ...interface{}) {
	lg.sl.Errorf(tmpl, args...)
}

// Errorw is for Errorw
func (lg *St) Errorw(msg string, err interface{}, args ...interface{}) {
	kvs := make([]interface{}, 0, len(args)+2)
	kvs = append(kvs, "error", err)
	kvs = append(kvs, args...)
	lg.sl.Errorw(msg, kvs...)
}

// Warn is for Warn
func (lg *St) Warn(args ...interface{}) {
	lg.sl.Warn(args...)
}

// Warnf is for Warnf
func (lg *St) Warnf(tmpl string, args ...interface{}) {
	lg.sl.Warnf(tmpl, args...)
}

// Warnw is for Warnw
func (lg *St) Warnw(msg string, args ...interface{}) {
	lg.sl.Warnw(msg, args...)
}

// Info is for Info
func (lg *St) Info(args ...interface{}) {
	lg.sl.Info(args...)
}

// Infof is for Infof
func (lg *St) Infof(tmpl string, args ...interface{}) {
	lg.sl.Infof(tmpl, args...)
}

// Infow is for Infow
func (lg *St) Infow(msg string, args ...interface{}) {
	lg.sl.Infow(msg, args...)
}

// Debug is for Debug
func (lg *St) Debug(args ...interface{}) {
	lg.sl.Debug(args...)
}

// Debugf is for Debugf
func (lg *St) Debugf(tmpl string, args ...interface{}) {
	lg.sl.Debugf(tmpl, args...)
}

// Debugw is for Debugw
func (lg *St) Debugw(msg string, args ...interface{}) {
	lg.sl.Debugw(msg, args...)
}

// Sync is for sync
func (lg *St) Sync() {
	err := lg.sl.Sync()
	if err != nil {
		log.Println("Fail to sync zap-logger", err)
	}
}
