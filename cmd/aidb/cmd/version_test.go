package cmd

import (
	"testing"
)

func TestVersionIsSetOnRootCmd(t *testing.T) {
	// After init(), rootCmd.Version must be set (not empty)
	if rootCmd.Version == "" {
		t.Error("rootCmd.Version should not be empty after init")
	}
}

func TestVersionOverridesDevDefault(t *testing.T) {
	// When built with ldflags (like goreleaser/make), version != "dev"
	// When run via go test, debug.ReadBuildInfo may return "(devel)"
	// Either way, rootCmd.Version must match the package-level version var
	if rootCmd.Version != version {
		t.Errorf("rootCmd.Version = %q, want %q (must match package var)", rootCmd.Version, version)
	}
}
