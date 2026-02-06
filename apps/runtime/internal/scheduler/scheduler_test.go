package scheduler

import (
	"context"
	"testing"
	"time"

	"pryx-core/internal/store"
)

type testExecutor struct {
	ch  chan *ScheduledTask
	err error
}

func (e *testExecutor) Execute(ctx context.Context, task *ScheduledTask) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if e.ch != nil {
		select {
		case e.ch <- task:
		default:
		}
	}

	if e.err != nil {
		return "", e.err
	}

	return "ok", nil
}

func TestValidateCronExpressionSupportsCronIntervalAndEvent(t *testing.T) {
	valid := []string{
		"0 * * * *",
		"@every 5m",
		"every 5 minutes",
		"event:mesh.sync",
	}

	for _, expr := range valid {
		if err := ValidateCronExpression(expr); err != nil {
			t.Fatalf("expected valid expression %q, got error: %v", expr, err)
		}
	}

	invalid := []string{
		"",
		"every bananas",
		"event:",
		"*/5 * *",
	}

	for _, expr := range invalid {
		if err := ValidateCronExpression(expr); err == nil {
			t.Fatalf("expected invalid expression %q", expr)
		}
	}
}

func TestSaveRunUpsert(t *testing.T) {
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer st.Close()

	s := New(st.DB)

	task := &ScheduledTask{
		Name:           "upsert-test",
		CronExpression: "@every 1s",
		TaskType:       TaskTypeMessage,
		Enabled:        false,
	}
	if err := s.CreateTask(task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	run := &TaskRun{
		ID:        "run-1",
		TaskID:    task.ID,
		StartedAt: time.Now(),
		Status:    RunStatusRunning,
	}
	if err := s.saveRun(run); err != nil {
		t.Fatalf("failed to save run start: %v", err)
	}

	now := time.Now()
	run.CompletedAt = &now
	run.Status = RunStatusSuccess
	run.Output = "done"
	if err := s.saveRun(run); err != nil {
		t.Fatalf("failed to save run completion: %v", err)
	}

	var count int
	if err := st.DB.QueryRow("SELECT COUNT(*) FROM scheduled_task_runs WHERE id = ?", run.ID).Scan(&count); err != nil {
		t.Fatalf("failed to query run count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 run row after upsert, got %d", count)
	}
}

func TestTriggerEventExecutesEventTasks(t *testing.T) {
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer st.Close()

	s := New(st.DB)
	exec := &testExecutor{ch: make(chan *ScheduledTask, 1)}
	s.RegisterExecutor(TaskTypeMessage, exec)

	task := &ScheduledTask{
		Name:           "event-task",
		CronExpression: "event:user.login",
		TaskType:       TaskTypeMessage,
		Enabled:        true,
	}
	if err := s.CreateTask(task); err != nil {
		t.Fatalf("failed to create event task: %v", err)
	}

	triggered, err := s.TriggerEvent("user.login")
	if err != nil {
		t.Fatalf("failed to trigger event: %v", err)
	}
	if triggered != 1 {
		t.Fatalf("expected 1 triggered task, got %d", triggered)
	}

	select {
	case <-exec.ch:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event task execution")
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		runs, runErr := s.GetTaskRuns(task.ID, 10)
		if runErr == nil && len(runs) > 0 && runs[0].Status == RunStatusSuccess {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}

	t.Fatal("expected successful run record for triggered event task")
}

func TestSchedulerLoadsEnabledTasksOnStart(t *testing.T) {
	st, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer st.Close()

	s1 := New(st.DB)
	task := &ScheduledTask{
		Name:           "restart-recovery-task",
		CronExpression: "@every 1s",
		TaskType:       TaskTypeMessage,
		Enabled:        true,
	}
	if err := s1.CreateTask(task); err != nil {
		t.Fatalf("failed to create persisted task: %v", err)
	}

	s2 := New(st.DB)
	exec := &testExecutor{ch: make(chan *ScheduledTask, 1)}
	s2.RegisterExecutor(TaskTypeMessage, exec)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := s2.Start(ctx); err != nil {
		t.Fatalf("failed to start scheduler: %v", err)
	}
	defer s2.Stop()

	select {
	case <-exec.ch:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for scheduled task execution after start")
	}

	runs, err := s2.GetTaskRuns(task.ID, 10)
	if err != nil {
		t.Fatalf("failed to fetch task runs: %v", err)
	}
	if len(runs) == 0 {
		t.Fatal("expected at least one persisted task run after scheduler start")
	}
}
