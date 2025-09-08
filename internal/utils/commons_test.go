package utils_test

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// testFixture provides test utilities for utils package tests
type testFixture struct {
	logger *zap.Logger
}

func setupTest(t *testing.T) *testFixture {
	logger := zaptest.NewLogger(t)

	return &testFixture{
		logger: logger,
	}
}

func (f *testFixture) tearDown() {
	// No cleanup needed for these tests
}
