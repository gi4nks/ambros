// Package logging provides standardized logging utilities for the Ambros application.
package logging

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config represents logging configuration
type Config struct {
	Level      string
	OutputPath string
	DevMode    bool
}

// NewLogger creates a new configured logger
func NewLogger(cfg Config) (*zap.Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level: %v", err)
	}

	// Ensure log directory exists
	if cfg.OutputPath != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.OutputPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %v", err)
		}
	}

	var config zap.Config
	if cfg.DevMode {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	config.Level = zap.NewAtomicLevelAt(level)

	if cfg.OutputPath != "" {
		config.OutputPaths = []string{cfg.OutputPath}
	}

	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %v", err)
	}

	return logger, nil
}

// Fields represents a collection of log fields
type Fields map[string]interface{}

// With adds fields to a logger
func With(logger *zap.Logger, fields Fields) *zap.Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return logger.With(zapFields...)
}

// Common field keys
const (
	FieldKeyCommand   = "command"
	FieldKeyCommandID = "command_id"
	FieldKeyError     = "error"
	FieldKeyDuration  = "duration"
	FieldKeyStatus    = "status"
	FieldKeyUser      = "user"
	FieldKeyPath      = "path"
	FieldKeyOperation = "operation"
	FieldKeyComponent = "component"
)

// Component names for logging
const (
	ComponentAPI       = "api"
	ComponentScheduler = "scheduler"
	ComponentAnalytics = "analytics"
	ComponentCLI       = "cli"
	ComponentRepo      = "repository"
)

// Standard fields for common scenarios
func CommandFields(id, cmd string) Fields {
	return Fields{
		FieldKeyCommandID: id,
		FieldKeyCommand:   cmd,
	}
}

func ErrorFields(err error) Fields {
	return Fields{
		FieldKeyError: err.Error(),
	}
}

func OperationFields(component, operation string) Fields {
	return Fields{
		FieldKeyComponent: component,
		FieldKeyOperation: operation,
	}
}

// Convenience methods for common logging patterns
func LogCommandExecution(logger *zap.Logger, id, cmd string, err error, duration float64) {
	fields := CommandFields(id, cmd)
	fields[FieldKeyDuration] = duration
	fields[FieldKeyStatus] = "success"

	if err != nil {
		fields[FieldKeyStatus] = "error"
		fields[FieldKeyError] = err.Error()
		logger.Error("Command execution failed", toZapFields(fields)...)
		return
	}

	logger.Info("Command execution completed", toZapFields(fields)...)
}

func LogOperationResult(logger *zap.Logger, component, operation string, err error) {
	fields := OperationFields(component, operation)

	if err != nil {
		fields[FieldKeyError] = err.Error()
		logger.Error("Operation failed", toZapFields(fields)...)
		return
	}

	logger.Info("Operation completed", toZapFields(fields)...)
}

// Helper to convert Fields to zap.Field slice
func toZapFields(fields Fields) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return zapFields
}
