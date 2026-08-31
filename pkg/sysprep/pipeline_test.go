package sysprep

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

type mockStep struct {
	name      string
	runCalled bool
	err       error
}

func (m *mockStep) Name() string { return m.name }
func (m *mockStep) Run(ctx *Context) error {
	m.runCalled = true
	return m.err
}

func TestPipeline_SuccessfulRun(t *testing.T) {
	step1 := &mockStep{name: "Step 1"}
	step2 := &mockStep{name: "Step 2"}

	pipeline := NewPipeline(step1, step2)
	var buf bytes.Buffer
	ctx := &Context{Out: &buf}

	err := pipeline.Run(ctx)
	if err != nil {
		t.Fatalf("expected pipeline to succeed, got: %v", err)
	}

	if !step1.runCalled || !step2.runCalled {
		t.Errorf("expected all steps to be called, step1: %v, step2: %v", step1.runCalled, step2.runCalled)
	}

	output := buf.String()
	if !strings.Contains(output, "[1/2] Step 1...") || !strings.Contains(output, "[2/2] Step 2...") {
		t.Errorf("output does not contain step headers: %s", output)
	}
}

func TestPipeline_DryRun(t *testing.T) {
	step := &mockStep{name: "Destructive Step"}
	pipeline := NewPipeline(step)
	var buf bytes.Buffer
	ctx := &Context{DryRun: true, Out: &buf}

	err := pipeline.Run(ctx)
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}

	if step.runCalled {
		t.Errorf("step.Run() should NOT be called during dry-run")
	}

	if !strings.Contains(buf.String(), "(dry-run: step simulated)") {
		t.Errorf("expected dry-run notification in output: %s", buf.String())
	}
}

func TestPipeline_ErrorPropagation(t *testing.T) {
	step1 := &mockStep{name: "Failing Step", err: errors.New("simulated fatal error")}
	step2 := &mockStep{name: "Should Not Run"}

	pipeline := NewPipeline(step1, step2)
	var buf bytes.Buffer
	ctx := &Context{Out: &buf}

	err := pipeline.Run(ctx)
	if err == nil {
		t.Fatalf("expected pipeline to fail")
	}

	if !strings.Contains(err.Error(), "simulated fatal error") {
		t.Errorf("unexpected error message: %v", err)
	}

	if step2.runCalled {
		t.Errorf("subsequent steps should not run after a failure")
	}
}
