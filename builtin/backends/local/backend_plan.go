package local

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/backend"
	"github.com/hashicorp/terraform/terraform"
)

func (b *Local) opPlan(
	ctx context.Context,
	op *backend.Operation,
	runningOp *backend.RunningOperation) {
	// Get our state
	state, err := b.State()
	if err != nil {
		runningOp.Err = errwrap.Wrapf("Error loading state: {{err}}", err)
		return
	}
	if err := state.RefreshState(); err != nil {
		runningOp.Err = errwrap.Wrapf("Error loading state: {{err}}", err)
		return
	}
	runningOp.State = state.State()

	// Get our context
	tfCtx, err := b.Context(op, state)
	if err != nil {
		runningOp.Err = err
		return
	}

	// If we're refreshing before plan, perform that
	if op.PlanRefresh {
		_, err := tfCtx.Refresh()
		if err != nil {
			runningOp.Err = errwrap.Wrapf("Error refreshing state: {{err}}", err)
			return
		}
	}

	// Perform the plan
	plan, err := tfCtx.Plan()
	if err != nil {
		runningOp.Err = errwrap.Wrapf("Error running plan: {{err}}", err)
		return
	}

	// Save the plan to disk
	if path := op.PlanOutPath; path != "" {
		log.Printf("[INFO] backend/local: writing plan output to: %s", path)
		f, err := os.Create(path)
		if err == nil {
			err = terraform.WritePlan(plan, f)
		}
		f.Close()
		if err != nil {
			runningOp.Err = fmt.Sprintf("Error writing plan file: %s", err)
			return
		}
	}
}
