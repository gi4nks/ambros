package commands

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/gi4nks/ambros/v3/internal/models"
	mockrepo "github.com/gi4nks/ambros/v3/internal/repos/mocks"
)

// wipeRepo is a small wrapper around the generated mock to add DeleteSchema
type wipeRepo struct{ *mockrepo.MockRepository }

func (w *wipeRepo) DeleteSchema(wipe bool) error {
	args := w.Called(wipe)
	return args.Error(0)
}

func TestDBInit_WipeYes(t *testing.T) {
	logger := zap.NewNop()
	// create a mock that also implements DeleteSchema via a small wrapper
	base := mockrepo.NewMockRepository()
	wr := &wipeRepo{base}
	wr.On("DeleteSchema", true).Return(nil)
	// The init flow probes DB accessibility via GetAllCommands
	base.On("GetAllCommands").Return([]models.Command{}, nil)

	// create command with wrapper repo
	dc := NewDBCommand(logger, wr)
	cmd := dc.Command()

	// run with --wipe and send 'yes' on stdin
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetIn(bytes.NewBufferString("yes\n"))
	// also set outputs for subcommands to ensure they inherit
	for _, sc := range cmd.Commands() {
		sc.SetOut(buf)
		sc.SetErr(buf)
		sc.SetIn(cmd.InOrStdin())
	}
	cmd.SetArgs([]string{"init", "--wipe"})

	// cobra ExecuteC will run the command
	_, err := cmd.ExecuteC()
	// We expect no error
	assert.NoError(t, err)
	wr.AssertExpectations(t)
}

func TestDBPrune_DryRunAndPrune(t *testing.T) {
	logger := zap.NewNop()
	mr := mockrepo.NewMockRepository()

	// Prepare commands: two old, one new
	now := time.Now()
	old1 := models.Command{Entity: models.Entity{ID: "old1", CreatedAt: now.AddDate(0, -2, 0)}}
	old2 := models.Command{Entity: models.Entity{ID: "old2", CreatedAt: now.AddDate(0, -1, -1)}}
	recent := models.Command{Entity: models.Entity{ID: "new1", CreatedAt: now}}

	mr.On("GetAllCommands").Return([]models.Command{old1, old2, recent}, nil)
	// Expect Delete called for the two old ones when not dry-run
	mr.On("Delete", "old1").Return(nil)
	mr.On("Delete", "old2").Return(nil)

	dc := NewDBCommand(logger, mr)
	cmd := dc.Command()

	// Dry-run first
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	for _, sc := range cmd.Commands() {
		sc.SetOut(buf)
		sc.SetErr(buf)
	}
	cmd.SetArgs([]string{"prune", "--before", now.AddDate(0, -1, 0).Format("2006-01-02"), "--dry-run"})
	_, err := cmd.ExecuteC()
	assert.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "Found 2 commands to delete")

	// Now real prune
	buf.Reset()
	// create a fresh command instance to avoid leftover flags/state
	dc2 := NewDBCommand(logger, mr)
	cmd2 := dc2.Command()
	cmd2.SetOut(buf)
	cmd2.SetErr(buf)
	for _, sc := range cmd2.Commands() {
		sc.SetOut(buf)
		sc.SetErr(buf)
	}
	cmd2.SetArgs([]string{"prune", "--before", now.AddDate(0, -1, 0).Format("2006-01-02")})
	_, err = cmd2.ExecuteC()
	assert.NoError(t, err)
	out = buf.String()
	assert.True(t, strings.Contains(out, "Pruned") || strings.Contains(out, "pruned"))

	mr.AssertExpectations(t)
}
