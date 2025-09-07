package commands_test

import (
	"testing"

	"go.uber.org/zap"
)

type testFixture struct {
	logger *zap.Logger
}

func setupTest(t *testing.T) *testFixture {
	logger := zap.NewNop()
	return &testFixture{
		logger: logger,
	}
}

func (f *testFixture) tearDown() {

}
