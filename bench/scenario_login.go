package main

// ログイン処理〜解答提出までのシナリオを一通り管理するファイル。

import (
	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/worker"
)

func (s *Scenario) NewLoginScenarioWorker(step *isucandar.BenchmarkStep, p int32) (*worker.Worker, error) {
	// あとで実装

	return nil, nil
}

func (s *Scenario) GetAgentFromUser(step *isucandar.BenchmarkStep, user *User) (*agent.Agent, error) {
	agent, err := user.GetAgent(s.Option)
	if err != nil {
		step.AddError(failure.NewError(ErrCannotCreateNewAgent, err))
		return nil, err
	}
	return agent, nil
}
