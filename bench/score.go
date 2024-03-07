package main

import (
	"context"
	"net"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/score"
)

const (
	ScoreSubmission score.ScoreTag = "提出成功"
	ScoreRegisteration score.ScoreTag = "ユーザ登録成功"
)

var ScoreRateTable = map[score.ScoreTag]int64{
	ScoreSubmission: 10,
	ScoreRegisteration: 15,
}

// シナリオ中に発生したエラーは1つ15点減点する
const ErrorDeduction = 15

func MakeScoreTable(score *score.Score) *score.Score {
	for tag, rate := range ScoreRateTable {
		score.Set(tag, rate)
	}
	return score
}

func isTimeout(err error) bool {
	var nerr net.Error
	if failure.As(err, &nerr) {
		if nerr.Timeout() || nerr.Temporary() {
			return true
		}
	}
	if failure.Is(err, context.DeadlineExceeded) ||
		failure.Is(err, context.Canceled) {
		return true
	}
	return failure.IsCode(err, failure.TimeoutErrorCode)
}

func ConstructBreakdown(result *isucandar.BenchmarkResult) score.ScoreTable {
	bd := result.Score.Breakdown()
	for tag := range ScoreRateTable {
		if _, ok := bd[tag]; !ok {
			bd[tag] = int64(0)
		}
	}
	return bd
}
