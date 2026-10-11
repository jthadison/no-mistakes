package cli

import (
	"strings"
	"testing"

	"github.com/kunchenguid/no-mistakes/internal/custody"
	"github.com/kunchenguid/no-mistakes/internal/db"
	"github.com/kunchenguid/no-mistakes/internal/paths"
	"github.com/kunchenguid/no-mistakes/internal/types"
)

func TestAxiSyncAdoptReconciledLocalEndToEnd(t *testing.T) {
	f := newCLIRecoverFixture(t)
	cliGit(t, f.local, "fetch", f.gate, "refs/heads/feature/recover:refs/remotes/test/preserved")
	cliGit(t, f.gate, "update-ref", "refs/heads/feature/recover", f.submitted, f.preserved)
	cliGit(t, f.gate, "update-ref", custody.RecoveryGateRef(f.runID), f.submitted)
	cliGit(t, f.local, "reset", "--hard", f.preserved)
	cliGit(t, f.local, "commit", "--allow-empty", "-m", "lossless reconciliation")
	localHead := cliGit(t, f.local, "rev-parse", "HEAD")

	p, err := paths.New()
	if err != nil {
		t.Fatal(err)
	}
	database, err := db.Open(p.DB())
	if err != nil {
		t.Fatal(err)
	}
	if err := database.UpdateRunStatusWithVerifiedHead(f.runID, types.RunFailed, f.preserved); err != nil {
		database.Close()
		t.Fatal(err)
	}
	if err := database.SetRunCustodyReturned(f.runID); err != nil {
		database.Close()
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	out, err := executeCmd("axi", "sync", "--adopt-reconciled-local")
	if err != nil {
		t.Fatalf("adopt reconciled local: %v\n%s", err, out)
	}
	if !strings.Contains(out, "safety: adopted_reconciled_local") {
		t.Fatalf("unexpected output:\n%s", out)
	}
	if got := cliGit(t, f.gate, "rev-parse", "refs/heads/feature/recover"); got != localHead {
		t.Fatalf("gate lane = %s, want %s", got, localHead)
	}
}
