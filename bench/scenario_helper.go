package main

// scenario に使用できる便利関数をまとめておくファイル

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/worker"
)

// JSON 形式のレスポンスボディのパースを行う。
func parseJsonBody[T any](res *http.Response, dest *T) error {
	if !strings.Contains(res.Header.Get("Content-Type"), "application/json") {
		return errors.New("response doesn't have the header of `Content-Type: application/json`")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return failure.NewError(ErrInvalidJson, err)
	}
	err = json.Unmarshal(body, &dest)
	if err != nil {
		return failure.NewError(ErrInvalidJson, err)
	}
	return nil
}

func parseMasterJson[T any](jsonFile string, dest *T) error {
	file, err := os.Open(jsonFile)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&dest); err != nil {
		return err
	}
	return nil
}

// 1つの agent 内で処理を次のステップに進める際に、context が途中で中断しているケースがある。
// context が終了している場合には true を返す関数。利用側で処理の離脱を行うようにして使用する。
func checkIfContextOver(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
func (s *Scenario) RecordLoginSuccessCount(count int) {
	s.ScenarioControlWg.Add(1)
	go func(count int) {
		defer s.ScenarioControlWg.Done()
		s.LoginSuccessCountMu.Lock()
		defer s.LoginSuccessCountMu.Unlock()
		s.LoginSuccessCount += count
	}(count)
}

func (s *Scenario) RecordUserRegistrationCount(count int) {
	s.ScenarioControlWg.Add(1)
	go func(count int) {
		defer s.ScenarioControlWg.Done()
		s.UserRegistrationMu.Lock()
		defer s.UserRegistrationMu.Unlock()
		s.UserRegistrationCount += count
	}(count)
}

func (s *Scenario) RecordVisitorStandingsCount(count int) {
	s.ScenarioControlWg.Add(1)
	go func(count int) {
		defer s.ScenarioControlWg.Done()
		s.VisitorStandingsMu.Lock()
		defer s.VisitorStandingsMu.Unlock()
		s.VisitorStandingsCount += count
	}(count)
}

// ベンチマーカーの出すエラーの状況を考慮しながら、並列数などを徐々に追加していく。
// 設定した数以上のエラーを検出すると負荷テストを打ち切るようになっている。
func (s *Scenario) loadAdjustor(ctx context.Context, step *isucandar.BenchmarkStep, submitWorker, userRegistrationWorker, visitor *worker.Worker) {
	tk := time.NewTicker(time.Second * 10)
	var prevErrors int64
	activeuserCount := 0
	userRegistrationCount := 0
	visitorStandingsCount := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
		}

		errors := step.Result().Errors.Count()
		total := errors["load"]
		if total >= int64(MaxErrors) {
			ContestantLogger.Printf("負荷走行を打ち切ります (エラー数:%d)", total)
			AdminLogger.Printf("%#v", errors)
			step.Result().Score.Close()
			step.Cancel()
			return
		}

		loginParallels := int32(1)
		userRegistrationParallels := int32(1)
		if diff := total - prevErrors; diff > 5 {
			ContestantLogger.Printf("エラーが%d件増えました(現在%d件)", diff, total)
		} else {
			totalActiveUser := s.LoginSuccessCount
			activeuserCount = totalActiveUser - activeuserCount
			ContestantLogger.Printf("現在のアクティブユーザー数: [total: %d (+%d人)]", totalActiveUser, activeuserCount)

			totalRegister := s.UserRegistrationCount
			userRegistrationCount = totalRegister - userRegistrationCount
			ContestantLogger.Printf("現在のユーザー登録成功数: [total: %d (+%d人)]", totalRegister, userRegistrationCount)

			totalvisitorStandings := s.VisitorStandingsCount
			visitorStandingsCount = totalvisitorStandings - visitorStandingsCount
			ContestantLogger.Printf("現在の観戦者数: [total: %d (+%d人)]", totalvisitorStandings, visitorStandingsCount)

			// loginParallels, userRegistrationParallels を変更する

		}

		submitWorker.AddParallelism(loginParallels)
		userRegistrationWorker.AddParallelism(userRegistrationParallels)
		visitor.AddParallelism(userRegistrationParallels + loginParallels)
		prevErrors = total
	}
}

// シナリオを開始後、時刻がどの程度経過したかを通知します。
func TimeReporter(name string, o Option) func() {
	if !(o.Stage == "test") {
		return func() {}
	}
	start := time.Now()
	return func() {
		AdminLogger.Printf("Scenario:%s elapsed:%s", name, time.Since(start))
	}
}

var loopConfig = func(s *Scenario) worker.WorkerOption {
	if s.Option.Stage == "test" {
		return worker.WithLoopCount(1)
	} else if s.Option.Stage == "prod" {
		return worker.WithInfinityLoop()
	} else {
		panic("please set --stage option")
	}
}

var parallelismConfig = func(s *Scenario) worker.WorkerOption {
	if s.Option.Stage == "test" {
		return worker.WithMaxParallelism(1)
	} else if s.Option.Stage == "prod" {
		return worker.WithMaxParallelism(int32(s.Option.Parallelism))
	} else {
		panic("please set --stage option")
	}
}

type ScenarioResult struct {
	Rewind bool
}

func NoRewind() ScenarioResult {
	return ScenarioResult{
		Rewind: false,
	}
}

func Rewind() ScenarioResult {
	return ScenarioResult{
		Rewind: true,
	}
}

func SleepWithCtx(ctx context.Context, sleepTime time.Duration) {
	tick := time.After(sleepTime)
	select {
	case <-ctx.Done():
	case <-tick:
	}
}
