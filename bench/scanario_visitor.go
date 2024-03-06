package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/worker"
)

// 観戦のシナリオを一通り管理するファイル。

func (s *Scenario) NewVisitorScenarioWorker(step *isucandar.BenchmarkStep, p int32) (*worker.Worker, error) {
	// あとで実装
	visitor, err := worker.NewWorker(func(ctx context.Context, _ int) {
		PrintScenarioStarted(ScenarioVisitor)
		defer PrintScenarioFinished(ScenarioVisitor)
		time.Sleep(1 * time.Second)

		agent, err := s.GetAgent(step)
		if err != nil {
			return
		}

		// tasks (確率的に)
		if rand.Float64() < 0.5 {
			s.GetTasksSuccessScenario_guest(ctx, step, agent)
			gettaskcnt := rand.Intn(3)
			for i := 0; i < gettaskcnt; i++ {
				task := s.Tasks.At(rand.Intn(s.Tasks.Len()))
				s.GetTaskSuccessScenario_guest(ctx, step, agent, task)
			}
		}

		// standings
		s.GetStandingsSuccessScenario_guest(ctx, step, agent)
		// 上位チームの情報を見る

		// 成功
		s.RecordVisitorCount(1)
	}, loopConfig(s), parallelismConfig(s))

	visitor.SetParallelism(p)

	return visitor, err
}

func (s *Scenario) GetTasksSuccessScenario_guest(ctx context.Context, step *isucandar.BenchmarkStep, agent *agent.Agent) {
	report := TimeReporter("tasks 取得 シナリオ", s.Option)
	defer report()

	gettasksRes, err := GetTasksAction(ctx, agent)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return
	}
	defer gettasksRes.Body.Close()

	gettasksResponse := &[]TaskAbstract{}

	gettasksValidation := ValidateResponse(
		gettasksRes,
		WithStatusCode(200),
		WithJsonBody(gettasksResponse),
	)
	gettasksValidation.Add(step)
}

func (s *Scenario) GetTaskSuccessScenario_guest(ctx context.Context, step *isucandar.BenchmarkStep, agent *agent.Agent, task *Task) {
	report := TimeReporter("task 取得 シナリオ", s.Option)
	defer report()

	gettaskRes, err := GetTaskAction(ctx, agent, task.Name)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return
	}
	defer gettaskRes.Body.Close()

	gettaskResponse := &Task{}

	gettaskValidation := ValidateResponse(
		gettaskRes,
		WithStatusCode(200),
		WithJsonBody(gettaskResponse),
	)
	gettaskValidation.Add(step)
}

func (s *Scenario) GetStandingsSuccessScenario_guest(ctx context.Context, step *isucandar.BenchmarkStep, agent *agent.Agent) {
	report := TimeReporter("standings 取得 シナリオ", s.Option)
	defer report()

	getstandingsRes, err := GetStandingsAction(ctx, agent)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return
	}
	defer getstandingsRes.Body.Close()

	getstandingsresponce := &Standings{}

	getstandingsValidation := ValidateResponse(
		getstandingsRes,
		WithStatusCode(200),
		WithJsonBody(getstandingsresponce),
	)
	getstandingsValidation.Add(step)
}
