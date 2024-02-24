package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
)

const (
	DefaultTargetHost               = "localhost:8080"
	DefaultRequestTimeout           = 3 * time.Second
	DefaultInitializeRequestTimeout = 10 * time.Second
	DefaultExitErrorOnFail          = true
	DefaultStage                    = "test"
)

func init() {
	failure.BacktraceCleaner.Add(failure.SkipGOROOT)
}

type errorSummary struct {
	initializeError []error
	scenarioError   []error
	validationError []error
	internalError   []error
	unexpectedError []error
}

func (e errorSummary) containsFatal() bool {
	return len(e.internalError) != 0 || len(e.unexpectedError) != 0 || len(e.initializeError) != 0 || len(e.validationError) != 0
}

func (e errorSummary) containsError() bool {
	return e.containsFatal() || len(e.scenarioError) != 0 || len(e.internalError) != 0 || len(e.validationError) != 0 || len(e.unexpectedError) != 0
}

func main() {
	option := Option{}

	flag.StringVar(&option.TargetHost, "target-host", DefaultTargetHost, "Benchmark target host with port")
	flag.DurationVar(&option.RequestTimeout, "request-timeout", DefaultRequestTimeout, "Default request timeout")
	flag.DurationVar(&option.InitializeRequestTimeout, "initialize-request-timeout", DefaultInitializeRequestTimeout, "Initialize request timeout")
	flag.BoolVar(&option.ExitErrorOnFail, "exit-error-on-fail", DefaultExitErrorOnFail, "Exit with error if benchmark fails")
	flag.StringVar(&option.Stage, "stage", DefaultStage, "Set stage which affects the amount of request")

	flag.Parse()

	AdminLogger.Print(option)

	time.Local = time.FixedZone("Asia/Tokyo", 9*60*60) // JST
	scenario := &Scenario{
		Option:          option,
		ConsumedUserIDs: NewLightSet(),
		NowClock:        time.Date(2024, 3, 28, 9, 0, 0, 0, time.Local),
	}

	benchmark, err := isucandar.NewBenchmark(
		isucandar.WithoutPanicRecover(),
		isucandar.WithLoadTimeout(1*time.Minute),
	)
	if err != nil {
		AdminLogger.Fatal(err)
	}

	benchmark.AddScenario(scenario)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := benchmark.Start(ctx)

	// 結果の集計のために少し待つ
	result.Errors.Wait()
	time.Sleep(3 * time.Second)
	

	for _, err := range result.Errors.All() {
		ContestantLogger.Printf("%v", err)
		AdminLogger.Printf("%+v", err)
	}

	for tag, count := range result.Score.Breakdown() {
		AdminLogger.Printf("%s: %d", tag, count)
	}

	score := SumScore(result)
	ContestantLogger.Printf("score: %d", score)

	if option.ExitErrorOnFail && score <= 0 {
		os.Exit(1)
	}
}

func SumScore(result *isucandar.BenchmarkResult) int64 {
	score := result.Score
	// 各タグに倍率を設定
	/*
		score.Set(ScoreGETRoot, 1)
		score.Set(ScoreGETLogin, 1)
		score.Set(ScorePOSTLogin, 2)
		score.Set(ScorePOSTRoot, 5)
	*/
	score.Set(ScoreSubmission, 1)

	addition := score.Sum()

	deduction := len(result.Errors.All())

	sum := addition - int64(deduction)
	if sum < 0 {
		sum = 0
	}

	return sum
}
