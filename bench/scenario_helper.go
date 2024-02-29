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

func (s *Scenario) RecordVisitorCount(count int) {
	s.ScenarioControlWg.Add(1)
	go func(count int) {
		defer s.ScenarioControlWg.Done()
		s.VisitorMu.Lock()
		defer s.VisitorMu.Unlock()
		s.VisitorCount += count
	}(count)
}

// ベンチマーカーの出すエラーの状況を考慮しながら、並列数などを徐々に追加していく。
// 設定した数以上のエラーを検出すると負荷テストを打ち切るようになっている。
func (s *Scenario) loadAdjustor(ctx context.Context, step *isucandar.BenchmarkStep, submitWorker, userRegistrationWorker, visitor *worker.Worker) {
	tk := time.NewTicker(time.Second * 10)
	var prevErrors, prevtimeout int64
	bef_totalActiveUser := 0
	bef_totalRegister := 0
	bef_totalvisitor := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
		}

		errors := step.Result().Errors.Count()
		total := errors["load"]
		timeout := errors["timeout"] / 2 // step.Adderror 内でも NewError が呼ばれるため、timeout タグが重複してついてしまう
		if total-timeout >= int64(MaxErrors) {
			ContestantLogger.Printf("負荷走行を打ち切ります (タイムアウトを除くエラー数:%d)", total-timeout)
			AdminLogger.Printf("%#v", errors)
			step.Result().Score.Close()
			step.Cancel()
			return
		}
		loginParallels := int32(0)
		userRegistrationParallels := int32(0)
		diff := total - prevErrors
		timeoutdiff := timeout - prevtimeout
		if diff > 0 {
			ContestantLogger.Printf("エラーが%d件(うちタイムアウト%d件)増えました (現在%d件 うちタイムアウト%d件)", diff, timeoutdiff, total, timeout)
		}
		totalActiveUser := s.LoginSuccessCount
		activeuserCount := totalActiveUser - bef_totalActiveUser
		bef_totalActiveUser = totalActiveUser
		ContestantLogger.Printf("現在の行動成功ユーザー数: %d人 (+%d人)", totalActiveUser, activeuserCount)

		totalRegister := s.UserRegistrationCount
		userRegistrationCount := totalRegister - bef_totalRegister
		bef_totalRegister = totalRegister
		ContestantLogger.Printf("現在のユーザー登録成功数: %d人 (+%d人)", totalRegister, userRegistrationCount)

		totalvisitorStandings := s.VisitorCount
		visitorCount := totalvisitorStandings - bef_totalvisitor
		bef_totalvisitor = totalvisitorStandings
		ContestantLogger.Printf("現在の観戦者数: %d人 (+%d人)", totalvisitorStandings, visitorCount)

		if diff >= 5 {
			ContestantLogger.Print("エラーが発生しすぎているため、ユーザーは増えません")
		} else {
			// loginParallels, userRegistrationParallels を変更する
			ContestantLogger.Print("処理成功数に応じてユーザーが増えます")
			loginParallels = 1
			userRegistrationParallels = 1
		}

		if loginParallels > 0 {
			submitWorker.AddParallelism(loginParallels)
		}
		if userRegistrationParallels > 0 {
			userRegistrationWorker.AddParallelism(userRegistrationParallels)
		}
		if userRegistrationParallels+loginParallels > 0 {
			visitor.AddParallelism(userRegistrationParallels + loginParallels)
		}
		prevErrors = total
		prevtimeout = timeout
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

func SleepWithCtx(ctx context.Context, sleepTime time.Duration) {
	tick := time.After(sleepTime)
	select {
	case <-ctx.Done():
	case <-tick:
	}
}
