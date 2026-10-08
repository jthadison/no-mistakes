package branchsync

import "testing"

func TestCustodyReturnedSymbolicGateLaneIsNotReadyForPipeline(t *testing.T) {
	t.Parallel()
	f, localHead := newReconciledLocalFixture(t)
	branchRef := "refs/heads/feature/recover"
	unrelatedRef := "refs/heads/unrelated"

	mustRun(t, f.gate, "fetch", f.local, "HEAD:"+unrelatedRef)
	mustRun(t, f.gate, "symbolic-ref", branchRef, unrelatedRef)

	inspected := f.service.InspectCached(f.ctx)
	if inspected.Safety == "gate_ready" {
		t.Fatalf("symbolic gate lane was reported ready: %#v", inspected)
	}
	if inspected.NextAction != nil && inspected.NextAction.Code == "adopt_reconciled_local" {
		t.Fatalf("symbolic gate lane was offered reconciled-local adoption: %#v", inspected)
	}

	state := f.service.AdoptReconciledLocal(f.ctx)
	if state.Changed || state.Safety == "adopted_reconciled_local" || state.Safety == "already_adopted_reconciled_local" {
		t.Fatalf("symbolic gate lane was adopted: %#v", state)
	}
	if got := mustRun(t, f.gate, "symbolic-ref", branchRef); got != unrelatedRef {
		t.Fatalf("selected lane changed to %s, want symbolic target %s", got, unrelatedRef)
	}
	if got := mustRun(t, f.gate, "rev-parse", unrelatedRef); got != localHead {
		t.Fatalf("unrelated lane changed to %s, want local head %s", got, localHead)
	}
}
