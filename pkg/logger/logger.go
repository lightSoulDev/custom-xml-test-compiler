package logger

type Logger interface {
	LogInfo(message string) error
	LogWarn(message string) error
	LogError(message string) error
}