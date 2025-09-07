package utils

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

type Utilities struct {
	logger *zap.Logger
}

func NewUtilities(logger *zap.Logger) *Utilities {
	return &Utilities{logger: logger}
}

func (u *Utilities) AsJson(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func (u *Utilities) Random() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 12
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return ""
		}
		result[i] = letters[num.Int64()]
	}
	return string(result)
}

func (u *Utilities) Tail(slice []string) []string {
	if len(slice) < 2 {
		return []string{}
	}
	return slice[1:]
}

func (u *Utilities) Check(err error) {
	if err != nil {
		u.logger.Warn("Check error", zap.Error(err))
	}
}

func (u *Utilities) Fatal(err error) {
	if err != nil {
		u.logger.Error("Fatal error", zap.Error(err))
		panic(err)
	}
}

func (u *Utilities) ExistsPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (u *Utilities) CreatePath(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		u.logger.Error("CreatePath error", zap.Error(err))
		return err
	}
	return nil
}

func (u *Utilities) GetAbsolutePath(path string) (string, error) {
	if path == "" {
		return "", errors.New("path is empty")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		u.logger.Error("GetAbsolutePath error", zap.Error(err))
		return "", err
	}
	return abs, nil
}
