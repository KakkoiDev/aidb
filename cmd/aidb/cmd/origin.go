package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type originPin struct {
	Toplevel string `json:"toplevel"`
	Remote   string `json:"remote"`
}

// pinOrigin records which repo owns a project namespace and refuses adds from
// a different repo that shares the directory basename. Matching remote URLs
// (both non-empty) or a matching toplevel path identify the same project.
func pinOrigin(dbDir, namespaceDir, toplevel string) error {
	pinPath := filepath.Join(namespaceDir, ".origin")
	current := originPin{Toplevel: toplevel, Remote: GetRemoteURL(toplevel)}

	data, err := os.ReadFile(pinPath)
	if os.IsNotExist(err) {
		out, err := json.Marshal(current)
		if err != nil {
			return err
		}
		if err := os.WriteFile(pinPath, out, 0644); err != nil {
			return err
		}
		// Stage it so the pin travels with the store (and the tree stays clean)
		if err := exec.Command("git", "-C", dbDir, "add", pinPath).Run(); err != nil {
			return fmt.Errorf("failed to stage %s: %w", pinPath, err)
		}
		return nil
	}
	if err != nil {
		return err
	}

	var pinned originPin
	if err := json.Unmarshal(data, &pinned); err != nil {
		return fmt.Errorf("unreadable %s: %w", pinPath, err)
	}

	if pinned.Toplevel == current.Toplevel {
		return nil
	}
	if pinned.Remote != "" && current.Remote != "" && pinned.Remote == current.Remote {
		return nil
	}
	return fmt.Errorf("namespace %q belongs to %s (remote %q); refusing add from %s (remote %q) - rename one project directory",
		filepath.Base(namespaceDir), pinned.Toplevel, pinned.Remote, current.Toplevel, current.Remote)
}
