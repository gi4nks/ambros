package commands

import (
    "testing"

    "github.com/gi4nks/ambros/v3/internal/errors"
)

func TestIsValidPluginName(t *testing.T) {
    valid := []string{
        "test-plugin",
        "plugin_1",
        "my.plugin",
        "plugin123",
        "a",
        "plugin-123_name.v1",
    }
    for _, name := range valid {
        if !isValidPluginName(name) {
            t.Fatalf("expected valid plugin name: %s", name)
        }
    }

    invalid := []string{
        "../evil",
        "bad/name",
        ".hidden/evil",
        "with space",
        "semi;rm -rf /",
        "",
    }
    for _, name := range invalid {
        if isValidPluginName(name) {
            t.Fatalf("expected invalid plugin name: %s", name)
        }
    }
}

func TestInstallCreateInvalidPluginNameReturnsAppError(t *testing.T) {
    pc := &PluginCommand{}

    // invalid name should return an AppError (without creating files)
    if err := pc.installPlugin("bad/name"); err == nil {
        t.Fatalf("expected error for invalid plugin name")
    } else {
        if !errors.IsInvalidInput(err) {
            t.Fatalf("expected AppError with invalid input code, got: %v", err)
        }
    }

    if err := pc.createPlugin("../evil"); err == nil {
        t.Fatalf("expected error for invalid plugin name")
    } else {
        if !errors.IsInvalidInput(err) {
            t.Fatalf("expected AppError with invalid input code, got: %v", err)
        }
    }
}
