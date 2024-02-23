package main

import (
	"context"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/worker"
)

// 観戦のシナリオを一通り管理するファイル。

func (s *Scenario) NewVisitorScenarioWorker(step *isucandar.BenchmarkStep, p int32) (*worker.Worker, error) {
	// あとで実装
	visitor, err := worker.NewWorker(func(ctx context.Context, _ int) {
		PrintScenarioStarted(ScenarioVisitor)
		defer PrintScenarioFinished(ScenarioVisitor)

		// tasks (確率的に)
		// task (確率的に)

		// standings
		// 上位チームの情報を見る

	}, loopConfig(s), parallelismConfig(s))

	visitor.SetParallelism(p)

	return visitor, err
}
