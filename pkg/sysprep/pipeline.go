package sysprep

import (
	"fmt"
	"io"
	"os"
)

// Context holds runtime configuration for the sysprep pipeline.
type Context struct {
	DryRun           bool
	Verbose          bool
	PoweroffOnFinish bool
	SkipNetplan      bool
	Out              io.Writer
}

// Step represents a single idempotent preparation/sanitization task.
type Step interface {
	Name() string
	Run(ctx *Context) error
}

// Pipeline orchestrates step execution.
type Pipeline struct {
	steps []Step
}

// NewPipeline creates a new pipeline with the provided steps.
func NewPipeline(steps ...Step) *Pipeline {
	return &Pipeline{steps: steps}
}

// Run executes all steps sequentially.
func (p *Pipeline) Run(ctx *Context) error {
	if ctx.Out == nil {
		ctx.Out = os.Stdout
	}

	total := len(p.steps)
	for idx, step := range p.steps {
		fmt.Fprintf(ctx.Out, "[%d/%d] %s...\n", idx+1, total, step.Name())
		if ctx.DryRun {
			fmt.Fprintf(ctx.Out, "       (dry-run: step simulated)\n")
			continue
		}

		if err := step.Run(ctx); err != nil {
			return fmt.Errorf("step '%s' failed: %w", step.Name(), err)
		}
	}
	return nil
}
