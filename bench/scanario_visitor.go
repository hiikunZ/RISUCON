package main

import (
	"context"
	"fmt"
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

		// 静的ファイル
		s.GetIndexScenario_guest(ctx, step, agent)
		s.GetJSScenario_guest(ctx, step, agent)
		s.GetCSSScenario_guest(ctx, step, agent)
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
		topteamname, ok := s.GetStandingsSuccessScenario_guest(ctx, step, agent)
		if !ok {
			return
		}
		// 上位チームの情報を見る
		if rand.Float64() < 0.8 {
			teamname := ""
			if rand.Float64() < 0.5 {
				teamname = topteamname[0]
			} else {
				teamname = topteamname[rand.Intn(2)+1]
			}
			s.GetTeamSuccessScenario_guest(ctx, step, agent, teamname)
		}
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

	gettaskResponse := &TaskDetail{}

	gettaskValidation := ValidateResponse(
		gettaskRes,
		WithStatusCode(200),
		WithJsonBody(gettaskResponse),
	)
	gettaskValidation.Add(step)
}

func (s *Scenario) GetStandingsSuccessScenario_guest(ctx context.Context, step *isucandar.BenchmarkStep, agent *agent.Agent) ([]string, bool) {
	report := TimeReporter("standings 取得 シナリオ", s.Option)
	defer report()

	getstandingsRes, err := GetStandingsAction(ctx, agent)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return nil, false
	}
	defer getstandingsRes.Body.Close()

	getstandingsresponce := &Standings{}

	getstandingsValidation := ValidateResponse(
		getstandingsRes,
		WithStatusCode(200),
		WithJsonBody(getstandingsresponce),
	)
	getstandingsValidation.Add(step)

	if getstandingsValidation.IsEmpty() {
		topteamname := []string{}
		if len(getstandingsresponce.StandingsData) < 3 {
			AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidResponse, fmt.Errorf("GET /api/standings : StandingsData が正しくありません")))
			return nil, false
		}
		for i := 0; i < 3; i++ {
			if getstandingsresponce.StandingsData[i].TeamName == "" {
				AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidResponse, fmt.Errorf("GET /api/standings : StandingsData が正しくありません")))
				return nil, false
			}
			topteamname = append(topteamname, getstandingsresponce.StandingsData[i].TeamName)
		}
		return topteamname, true
	}
	return nil, false
}

func (s *Scenario) GetTeamSuccessScenario_guest(ctx context.Context, step *isucandar.BenchmarkStep, agent *agent.Agent, teamname string) {
	report := TimeReporter("task 取得 シナリオ", s.Option)
	defer report()

	getteamRes, err := GetTeamAction(ctx, agent, teamname)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return
	}
	defer getteamRes.Body.Close()

	getteamResponse := &TeamResponse{}

	getteamValidation := ValidateResponse(
		getteamRes,
		WithStatusCode(200),
		WithJsonBody(getteamResponse),
	)
	getteamValidation.Add(step)
}
func (s *Scenario) GetIndexScenario_guest(ctx context.Context, step *isucandar.BenchmarkStep, agent *agent.Agent) {
	report := TimeReporter("index 取得 シナリオ", s.Option)
	defer report()

	indexRes, err := GetIndexAction(ctx, agent)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return
	}
	defer indexRes.Body.Close()

	if rand.Intn(10) == 0 { // 10% の確率で確認 
		indexValidation := ValidateResponse(
			indexRes,
			ValidateStaticFile(indexhash),
		)
		indexValidation.Add(step)
	}
}

func (s *Scenario) GetJSScenario_guest(ctx context.Context, step *isucandar.BenchmarkStep, agent *agent.Agent) {
	report := TimeReporter("js 取得 シナリオ", s.Option)
	defer report()

	jsRes, err := GetJSAction(ctx, agent)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return
	}
	defer jsRes.Body.Close()

	if rand.Intn(10) == 0 { // 10% の確率で確認
		jsValidation := ValidateResponse(
			jsRes,
			ValidateStaticFile(jsfilehash),
		)
		jsValidation.Add(step)
	}
}

func (s *Scenario) GetCSSScenario_guest(ctx context.Context, step *isucandar.BenchmarkStep, agent *agent.Agent) {
	report := TimeReporter("css 取得 シナリオ", s.Option)
	defer report()

	cssRes, err := GetCSSAction(ctx, agent)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return
	}
	defer cssRes.Body.Close()

	if rand.Intn(10) == 0 { // 10% の確率で確認
		cssValidation := ValidateResponse(
			cssRes,
			ValidateStaticFile(cssfilehash),
		)
		cssValidation.Add(step)
	}
}
