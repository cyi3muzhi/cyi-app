// pkg/util/log/zerolog.go
package log

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/rs/zerolog"
)

// LogConfig 日志配置结构体
type LogConfig struct {
	LogDir              string `json:"log_dir"`               // 日志目录
	ServiceName         string `json:"service_name"`          // 服务名
	Level               string `json:"level"`                 // 日志级别 (debug, info, warn, error, fatal)
	EnableFileOutput    bool   `json:"enable_file_output"`    // 是否启用文件输出
	EnableConsoleOutput bool   `json:"enable_console_output"` // 是否启用控制台输出
	TimeFormat          string `json:"time_format"`           // 时间格式
	EnableColor         bool   `json:"enable_color"`          // 是否启用颜色
}

// DefaultLogConfig 默认日志配置
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		LogDir:              "logs",
		ServiceName:         "service",
		Level:               "info",
		EnableFileOutput:    true,
		EnableConsoleOutput: true,
		TimeFormat:          "2006-01-02 15:04:05",
		EnableColor:         true,
	}
}

// GetZerologLevel 将字符串级别转换为 zerolog.Level
func GetZerologLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// ZerologAdapter zerolog 适配器
type ZerologAdapter struct {
	logger zerolog.Logger
	level  zerolog.Level
}

// NewZerologAdapter 创建 zerolog 适配器
func NewZerologAdapter() *ZerologAdapter {
	return NewZerologAdapterWithConfig(DefaultLogConfig())
}

// NewZerologAdapterWithConfig 根据配置创建 zerolog 适配器
func NewZerologAdapterWithConfig(config *LogConfig) *ZerologAdapter {
	// 使用默认配置如果没有提供
	if config == nil {
		config = DefaultLogConfig()
	}

	var outputs []io.Writer

	// 配置控制台输出
	if config.EnableConsoleOutput {
		consoleOutput := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: config.TimeFormat,
			NoColor:    !config.EnableColor,
		}
		outputs = append(outputs, consoleOutput)
	}

	// 配置文件输出
	if config.EnableFileOutput {
		fileOutput, err := setupFileOutputWithConfig(config)
		if err == nil {
			outputs = append(outputs, fileOutput)
		}
	}

	// 如果没有输出配置，使用默认控制台输出
	if len(outputs) == 0 {
		consoleOutput := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: config.TimeFormat,
			NoColor:    !config.EnableColor,
		}
		outputs = append(outputs, consoleOutput)
	}

	// 合并输出
	multiOutput := zerolog.MultiLevelWriter(outputs...)

	// 设置日志级别
	level := GetZerologLevel(config.Level)

	// 创建 logger
	logger := zerolog.New(multiOutput).With().Timestamp().Logger().Level(level)

	return &ZerologAdapter{
		logger: logger,
		level:  level,
	}
}

// setupFileOutput 配置文件输出（兼容旧版本）
func setupFileOutput() (io.Writer, error) {
	return setupFileOutputWithConfig(DefaultLogConfig())
}

// setupFileOutputWithConfig 根据配置配置文件输出
func setupFileOutputWithConfig(config *LogConfig) (io.Writer, error) {
	// 确保日志目录存在
	logDir := config.LogDir
	if logDir == "" {
		logDir = "logs"
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	// 创建日志文件，使用服务名+日期命名
	serviceName := config.ServiceName
	if serviceName == "" {
		serviceName = "service"
	}
	logFile := fmt.Sprintf("%s/%s-%s.log", logDir, serviceName, time.Now().Format("2006-01-02"))
	fmt.Printf("logFile: %s\n", logFile)
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// 返回文件输出
	timeFormat := config.TimeFormat
	if timeFormat == "" {
		timeFormat = "2006-01-02 15:04:05"
	}

	return zerolog.ConsoleWriter{
		Out:        file,
		TimeFormat: timeFormat,
		NoColor:    true, // 文件输出禁用颜色
	}, nil
}

// Log 实现 log.Logger 接口
func (z *ZerologAdapter) Log(level log.Level, keyvals ...interface{}) error {
	// 创建带键值对的 logger
	logger := z.logger

	// 处理键值对，确保所有字段都有值
	keyvalsMap := make(map[string]interface{})
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 >= len(keyvals) {
			break
		}
		key, val := keyvals[i], keyvals[i+1]
		keyStr := fmt.Sprintf("%v", key)
		keyvalsMap[keyStr] = val
	}

	// 确保必要字段存在
	if _, ok := keyvalsMap["service.id"]; !ok {
		keyvalsMap["service.id"] = "-"
	}
	if _, ok := keyvalsMap["trace.id"]; !ok {
		keyvalsMap["trace.id"] = "-"
	}
	if _, ok := keyvalsMap["span.id"]; !ok {
		keyvalsMap["span.id"] = "-"
	}

	// 构建 logger
	for key, val := range keyvalsMap {
		logger = logger.With().Interface(key, val).Logger()
	}

	// 根据级别输出日志
	switch level {
	case log.LevelDebug:
		if z.level <= zerolog.DebugLevel {
			logger.Debug().Msg("")
		}
	case log.LevelInfo:
		if z.level <= zerolog.InfoLevel {
			logger.Info().Msg("")
		}
	case log.LevelWarn:
		if z.level <= zerolog.WarnLevel {
			logger.Warn().Msg("")
		}
	case log.LevelError:
		if z.level <= zerolog.ErrorLevel {
			logger.Error().Msg("")
		}
	case log.LevelFatal:
		if z.level <= zerolog.FatalLevel {
			logger.Fatal().Msg("")
		}
	default:
		if z.level <= zerolog.InfoLevel {
			logger.Info().Msg("")
		}
	}

	return nil
}

// With 实现 log.Logger 接口
func (z *ZerologAdapter) With(keyvals ...interface{}) log.Logger {
	// 创建带键值对的 logger
	logger := z.logger
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 >= len(keyvals) {
			break
		}
		key, val := keyvals[i], keyvals[i+1]
		logger = logger.With().Interface(fmt.Sprintf("%v", key), val).Logger()
	}

	return &ZerologAdapter{
		logger: logger,
	}
}
