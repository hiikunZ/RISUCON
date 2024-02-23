package main

// ログイン処理〜解答提出までのシナリオを一通り管理するファイル。

import (
	"context"
	"math/rand"
	"net/http"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/worker"
)

func (s *Scenario) NewLoginScenarioWorker(step *isucandar.BenchmarkStep, p int32) (*worker.Worker, error) {
	// あとで実装
	loginSuccess, err := worker.NewWorker(func(ctx context.Context, _ int) {
		PrintScenarioStarted(ScenarioLogin)
		defer PrintScenarioFinished(ScenarioLogin)

		var user *User
		for {
			trial := rand.Intn(s.Users.Len()-1) + 2 // id 1 は admin なので除外
			if !s.ConsumedUserIDs.Exists(int64(trial)) {
				s.ConsumedUserIDs.Add(int64(trial))
				user, _ = s.Users.Get(trial)
				break
			}
		}
		defer s.ConsumedUserIDs.Remove(int64(user.ID))

	Rewind:
		// ログイン
		result, ok := s.LoginSuccessScenario(ctx, step, user)
		if !ok {
			return
		}
		if result.Rewind {
			goto Rewind
		}
		// user / team 取得 (確率的)
		// 以下、何回かやる
		// tasks
		// task
		// submit
		// submission (確率的)
		// standings (確率的)

		// logout


		user.ClearAgent()
	}, loopConfig(s), parallelismConfig(s))

	loginSuccess.SetParallelism(p)

	return loginSuccess, err

}

// リクエストを送ってステータスコードが成功状態であることと、レスポンスボディの形式が正しいかを確認する。
func (s *Scenario) LoginSuccessScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User) (ScenarioResult, bool) {
	report := TimeReporter("ログイン成功シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return NoRewind(), false
	}

	loginRes, err := PostLoginAction(ctx, agent, user.Name, user.Password)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
	}
	defer loginRes.Body.Close()

	if loginRes.StatusCode == http.StatusUnprocessableEntity {
		return Rewind(), false
	}

	loginResponse := &LoginResponse{}

	loginValidation := ValidateResponse(
		loginRes,
		WithStatusCode(200),
		WithJsonBody(loginResponse),
	)
	loginValidation.Add(step)

	if loginValidation.IsEmpty() {
		return NoRewind(), true
	} else {
		return NoRewind(), false
	}
}

func (s *Scenario) GetAgentFromUser(step *isucandar.BenchmarkStep, user *User) (*agent.Agent, error) {
	agent, err := user.GetAgent(s.Option)
	if err != nil {
		step.AddError(failure.NewError(ErrCannotCreateNewAgent, err))
		return nil, err
	}
	return agent, nil
}
