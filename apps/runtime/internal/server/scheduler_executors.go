package server

import (
	"context"
	"encoding/json"
	"fmt"

	"pryx-core/internal/bus"
	"pryx-core/internal/scheduler"
)

type taskEventExecutor struct {
	bus *bus.Bus
}

func (e *taskEventExecutor) Execute(ctx context.Context, task *scheduler.ScheduledTask) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	var payload interface{}
	if task.Payload != "" {
		if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
			return "", fmt.Errorf("invalid task payload: %w", err)
		}
	}

	if payload == nil {
		payload = map[string]interface{}{}
	}

	if e.bus != nil {
		e.bus.Publish(bus.NewEvent(bus.EventTraceEvent, "", map[string]interface{}{
			"kind":      "scheduler.task.executed",
			"task_id":   task.ID,
			"task_name": task.Name,
			"task_type": task.TaskType,
			"payload":   payload,
		}))
	}

	return fmt.Sprintf("executed %s task", task.TaskType), nil
}

func (s *Server) registerSchedulerExecutors() {
	if s.scheduler == nil {
		return
	}

	executor := &taskEventExecutor{bus: s.bus}
	s.scheduler.RegisterExecutor(scheduler.TaskTypeMessage, executor)
	s.scheduler.RegisterExecutor(scheduler.TaskTypeWorkflow, executor)
	s.scheduler.RegisterExecutor(scheduler.TaskTypeReminder, executor)
	s.scheduler.RegisterExecutor(scheduler.TaskTypeWebhook, executor)
}
