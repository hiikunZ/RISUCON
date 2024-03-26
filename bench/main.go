package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/score"
)

const (
	DefaultTargetHost               = "localhost:8080"
	DefaultRequestTimeout           = 3 * time.Second
	DefaultInitializeRequestTimeout = 10 * time.Second
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

type scoreSummary struct {
	total     int64
	addition  int64
	deduction int64
	breakdown score.ScoreTable
}

func main() {
	option := Option{}

	flag.StringVar(&option.TargetHost, "target-host", DefaultTargetHost, "Benchmark target host with port")
	flag.DurationVar(&option.RequestTimeout, "request-timeout", DefaultRequestTimeout, "Default request timeout")
	flag.DurationVar(&option.InitializeRequestTimeout, "initialize-request-timeout", DefaultInitializeRequestTimeout, "Initialize request timeout")
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

	errorSummary := aggregateErrors(result)
	fatal := handleErrors(errorSummary, option.PrepareOnly, option.TargetHost)

	var score *scoreSummary
	var passed bool

	if fatal {
		// エラーハンドリングの結果、fatal 判定だった場合はここで処理を終了する
		ContestantLogger.Println("続行不可能なエラーが検出されたので、ここで処理を終了します。")
		score = &scoreSummary{
			total:     0,
			addition:  0,
			deduction: 0,
			breakdown: ConstructBreakdown(result),
		}
		passed = false
	} else {
		score = sumScore(result, errorSummary)
		passed = true
	}

	ContestantLogger.Printf("[PASSED]: %v", passed)
	ContestantLogger.Printf("[SCORE]: %d (addition: %d, deduction: %d)", score.total, score.addition, score.deduction)
	AdminLogger.Printf("[SCORE] %v", score.breakdown)

	AdminLogger.Printf("[PASSED]: %v,[SCORE]: %d", passed, score.total)
}

func sumScore(result *isucandar.BenchmarkResult, errorSummary *errorSummary) *scoreSummary {
	score := result.Score
	score = MakeScoreTable(score)

	addition := score.Sum()
	deduction := int64(len(errorSummary.scenarioError) * ErrorDeduction)

	sum := addition - deduction
	if sum < 0 {
		sum = 0
	}

	return &scoreSummary{
		total:     sum,
		addition:  addition,
		deduction: deduction,
		breakdown: ConstructBreakdown(result),
	}
}

func aggregateErrors(result *isucandar.BenchmarkResult) *errorSummary {
	initializeError := []error{}
	scenarioError := []error{}
	validationError := []error{}
	internalError := []error{}
	unexpectedError := []error{}

	for _, err := range result.Errors.All() {
		category := ClassifyError(err)

		switch category {
		case InitializeErr:
			initializeError = append(initializeError, err)
		case ScenarioErr:
			scenarioError = append(scenarioError, err)
		case ValidationErr:
			validationError = append(validationError, err)
		case InternalErr:
			internalError = append(internalError, err)
		case IsucandarMarked:
			continue
		default:
			unexpectedError = append(unexpectedError, err)
		}
	}

	return &errorSummary{
		initializeError,
		scenarioError,
		validationError,
		internalError,
		unexpectedError,
	}
}

// たとえば初期化処理のみ実行するモード（prepare-only mode）の際にエラーが発生した際は「続行不可能（fatal）」として判定させるが、
// fatal と判定された場合、この関数は true として値を返す。false を返す場合は、続行可能を意味する。
func handleErrors(summary *errorSummary, prepareOnly bool, targethost string) bool {
	for _, err := range summary.internalError {
		ContestantLogger.Printf("[INTERNAL] %v", err)
	}
	for _, err := range summary.unexpectedError {
		ContestantLogger.Printf("[UNEXPECTED] %v\n", err)
	}
	for _, err := range summary.initializeError {
		ContestantLogger.Printf("[INITIALIZATION_ERR] %v\n", err)
	}
	for _, err := range summary.validationError {
		ContestantLogger.Printf("[VALIDATION_ERR] %v\n", err)
	}

	// シナリオエラーは数が大量になりえるので、あまりに数が膨大になった場合には件数を絞って表示する
	var printErrorWindow []error
	aboveThreshold := false
	if len(summary.scenarioError) > MaxErrors {
		printErrorWindow = summary.scenarioError[0:MaxErrors]
		aboveThreshold = true
	} else {
		printErrorWindow = summary.scenarioError[0:]
	}

	if aboveThreshold {
		ContestantLogger.Printf("負荷走行で発生したエラー%d件のうち、最初から%d件を表示します", len(summary.scenarioError), MaxErrors)
	} else {
		ContestantLogger.Printf("負荷走行で発生したエラー%d件を表示します", len(summary.scenarioError))
	}
	for i, err := range printErrorWindow {
		AdminLogger.Printf("ERROR[%d] %+v", i+1, err)
		AdminLogger.Printf("ERROR[%d] %v", i+1, err)
		if failure.IsCode(err, failure.TimeoutErrorCode) {
			var nerr net.Error
			failure.As(err, &nerr)
			errcode := nerr.Error()
			method := strings.ToUpper(strings.Split(errcode, " ")[0])
			AdminLogger.Print(errcode)
			p := strings.Split(errcode, targethost)
			if len(p) >= 2 {
				path := strings.Split(p[1], "\"")[0]
				ContestantLogger.Printf("ERROR[%d] %s %s : タイムアウトが発生しました", i+1, method, path)
			} else {
				ContestantLogger.Printf("ERROR[%d] タイムアウトが発生しました", i+1)
			}

		} else if failure.IsCode(err, ErrInvalidStatusCode) {
			message := strings.Split(fmt.Sprintf("%v", err), "scenario-error-status-code: ")[1]
			ContestantLogger.Printf("ERROR[%d] %v", i+1, message)
		} else {
			ContestantLogger.Printf("ERROR[%d] %v", i+1, err)
		}
	}

	// prepare only モードの場合は、エラーが1件でもあればエラーで終了させる
	if prepareOnly {
		if summary.containsError() {
			return true
		}
	}

	return summary.containsFatal()
}
