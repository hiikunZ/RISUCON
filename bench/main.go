package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
)

var (
	ContestantLogger = log.New(os.Stdout, "", log.Ltime|log.Lmicroseconds)
	AdminLogger = log.New(os.Stderr, "[ADMIN] ", log.Ltime|log.Lmicroseconds)
)

const (
	DefaultTargetHost               = "localhost:8080"
	DefaultRequestTimeout           = 3 * time.Second
	DefaultInitializeRequestTimeout = 10 * time.Second
	DefaultExitErrorOnFail          = true
)

func init() {
	failure.BacktraceCleaner.Add(failure.SkipGOROOT)
}

func main() {
	option := Option{}

	flag.StringVar(&option.TargetHost, "target-host", DefaultTargetHost, "Benchmark target host with port")
	flag.DurationVar(&option.RequestTimeout, "request-timeout", DefaultRequestTimeout, "Default request timeout")
	flag.DurationVar(&option.InitializeRequestTimeout, "initialize-request-timeout", DefaultInitializeRequestTimeout, "Initialize request timeout")
	flag.BoolVar(&option.ExitErrorOnFail, "exit-error-on-fail", DefaultExitErrorOnFail, "Exit with error if benchmark fails")

	flag.Parse()

	AdminLogger.Print(option)

	scenario := &Scenario{
		Option: option,
		ConsumedUserIDs: NewLightSet(),
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
	

	addition := score.Sum()


	deduction := len(result.Errors.All())


	sum := addition - int64(deduction)
	if sum < 0 {
		sum = 0
	}

	return sum
}