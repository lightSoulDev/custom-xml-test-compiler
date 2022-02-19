package consoleLogger

import (
	"fmt"
	"strings"
	"time"
)

var (
	INFO string = "INF"
	WARN string = "WRN"
	ERROR string = "ERR"
)

// ConsoleLogger ...
type ConsoleLogger struct {
	
}

// New ...
func New() *ConsoleLogger {
	return &ConsoleLogger{}
}

func (l *ConsoleLogger) log(message, prefix string) {
	t := strings.Split(time.Now().UTC().String(), " ")
	timeStamp := strings.Join(t[:2], "|")
	fmt.Printf("[%s][%s] %s\n", prefix, timeStamp, message)
}

func (l *ConsoleLogger) LogInfo(message string) error {
	l.log(message, INFO)
	return nil
}

func (l *ConsoleLogger) LogWarn(message string) error {
	l.log(message, WARN)
	return nil
}

func (l *ConsoleLogger) LogError(message string) error {
	l.log(message, ERROR)
	return nil
}