package pc

import (
	"cmp"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type Log struct {
	File      string `yaml:"file"`      // 默认值为 pc.log
	AddSource bool   `yaml:"addSource"` // 是否打印错误位置
	Level     int    `yaml:"level"`     // 日志级别，默认为 slog.LevelInfo
}

// Logger 根据配置，构建 slog.Logger
// todo 日志滚动切割
func (l *Log) Logger() (*slog.Logger, error) {
	if l == nil {
		return slog.Default(), nil
	}

	path := cmp.Or(l.File, "pc.log")

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o664)
	if err != nil {
		return nil, fmt.Errorf("open file: %+v, error: %w", l.File, err)
	}

	logger := slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{
		AddSource:   l.AddSource,
		Level:       slog.Level(l.Level),
		ReplaceAttr: nil,
	}))

	return logger, nil
}
