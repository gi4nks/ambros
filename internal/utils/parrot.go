package utils

import (
	"fmt"

	"go.uber.org/zap"
)

// Parrot is a logging utility that provides console output capabilities
type ParrotLogger struct {
	logger *zap.Logger
}

// NewParrot creates a new Parrot instance
func NewParrot(logger *zap.Logger) *ParrotLogger {
	return &ParrotLogger{
		logger: logger,
	}
}

// Println prints a line to stdout and logs at info level
func (p *ParrotLogger) Println(msg interface{}) {
	fmt.Println(msg)
	if p.logger != nil {
		p.logger.Info(fmt.Sprintf("%v", msg))
	}
}

// Print prints to stdout and logs at info level
func (p *ParrotLogger) Print(msg interface{}) {
	fmt.Print(msg)
	if p.logger != nil {
		p.logger.Info(fmt.Sprintf("%v", msg))
	}
}

// Debug prints debug message and logs at debug level
func (p *ParrotLogger) Debug(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	fmt.Println("[DEBUG]", formatted)
	if p.logger != nil {
		p.logger.Debug(formatted)
	}
}

// Error prints error message and logs at error level
func (p *ParrotLogger) Error(msg string, err error) {
	if err != nil {
		formatted := fmt.Sprintf("%s: %v", msg, err)
		fmt.Println("[ERROR]", formatted)
		if p.logger != nil {
			p.logger.Error(msg, zap.Error(err))
		}
	} else {
		fmt.Println("[ERROR]", msg)
		if p.logger != nil {
			p.logger.Error(msg)
		}
	}
}

// Info prints info message and logs at info level
func (p *ParrotLogger) Info(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	fmt.Println("[INFO]", formatted)
	if p.logger != nil {
		p.logger.Info(formatted)
	}
}

// Warn prints warning message and logs at warn level
func (p *ParrotLogger) Warn(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	fmt.Println("[WARN]", formatted)
	if p.logger != nil {
		p.logger.Warn(formatted)
	}
}