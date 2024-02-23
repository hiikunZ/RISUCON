package main

import (
	"context"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/worker"
)

// ユーザー/チーム登録のシナリオを一通り管理するファイル。

func (s *Scenario) NewUserRegistrationScenarioWorker(step *isucandar.BenchmarkStep, p int32) (*worker.Worker, error) {
	// あとで実装
	userRegistration, err := worker.NewWorker(func(ctx context.Context, _ int) {
		// チーム内の人数を決めて、その人数分 register
		// createteam
		// join
	}, loopConfig(s), parallelismConfig(s))

	userRegistration.SetParallelism(p)

	return userRegistration, err
}
