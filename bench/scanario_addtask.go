package main

import (
	"context"
	"net/http"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/worker"
)

// いい感じに発火させるマスター更新バージョンアップのシナリオ。

func (s *Scenario) FireAddTask(step *isucandar.BenchmarkStep) (*worker.Worker, error) {
	// あとで実装
	worker, err := worker.NewWorker(func(ctx context.Context, _ int) {
		PrintScenarioStarted("タスク 追加チェック シナリオ")
		defer PrintScenarioFinished("タスク 追加チェック シナリオ")
		time.Sleep(3 * time.Second)
		if s.AddTask && s.Tasks.Len() < s.Prepared_Tasks.Len(){
			PrintScenarioStarted("タスク 追加 シナリオ")
			s.AddTask = false
			adminuser, _ := s.Users.Get(1)
			s.LoginSuccessScenario(ctx, step, adminuser)
			s.AddTaskScanario(ctx, step, adminuser)
		}

	}, loopConfig(s), worker.WithMaxParallelism(1))

	worker.SetParallelism(1)
	return worker, err
}

func (s *Scenario) AddTaskScanario(ctx context.Context, step *isucandar.BenchmarkStep, user *User) {
	report := TimeReporter("タスク 追加 シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return
	}

	task, _ := s.Prepared_Tasks.Get(s.Tasks.Len() + 1)
	subtaskdata := []SubtaskRequest{}
	for _, subtask := range task.SubTasks {
		ans := []AnswerRequest{}
		for _, answer := range subtask.Answers {
			ans = append(ans, AnswerRequest{
				Answer: answer.Answer,
				Score:  answer.Score,
			})
		}
		subtaskdata = append(subtaskdata, SubtaskRequest{
			Name:        subtask.Name,
			DisplayName: subtask.DisplayName,
			Statement:   subtask.Statement,
			Answers:     ans,
		})
	}
	taskdata := CreateTaskRequest{
		Name:            task.Name,
		DisplayName:     task.DisplayName,
		Statement:       task.Statement,
		SubmissionLimit: task.SubmissionLimit,
		Subtasks:        subtaskdata,
	}
	res, err := PostCreateTaskAction(ctx, agent, taskdata)
	if err != nil {
		return
	}
	defer res.Body.Close()

	validation := ValidateResponse(
		res,
		WithStatusCode(http.StatusCreated),
	)

	validation.Add(step)

	if validation.IsEmpty() {
		s.Tasks.Add(task)
	}
}
