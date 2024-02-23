package main

import (
	"context"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/worker"
)

// いい感じに発火させるマスター更新バージョンアップのシナリオ。

func (s *Scenario) FireAddTask(step *isucandar.BenchmarkStep) (*worker.Worker, error) {
	// あとで実装
	worker, err := worker.NewWorker(func(ctx context.Context, _ int) {
	}, worker.WithLoopCount(1), worker.WithMaxParallelism(1))

	worker.SetParallelism(1)
	return worker, err
}
