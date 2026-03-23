package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

func Init(level string) error {
	var err error
	cfg := zap.NewProductionConfig()
	if cfg.Level, err = zap.ParseAtomicLevel(level); err != nil {
		return err
	}

	if Log, err = cfg.Build(); err != nil {
		return err
	}

	return nil
}
