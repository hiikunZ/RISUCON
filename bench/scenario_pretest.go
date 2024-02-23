package main

import (
	"context"
	"math/rand"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
)

func (s *Scenario) loginValidateSuccessScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User) error {
	report := TimeReporter("ログイン成功 整合性チェック", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)

	if err != nil {
		return err
	}
	loginRes, err := PostLoginAction(ctx, agent, user.Name, user.Password)
	if err != nil {
		return failure.NewError(ValidationErrInvalidRequest, err)
	}
	defer loginRes.Body.Close()

	// HTTP status code のチェック
	if err := IsuAssertStatus(200, loginRes.StatusCode, Hint(PostLogin, "")); err != nil {
		return failure.NewError(ValidationErrInvalidStatusCode, err)
	}
	loginResponse := &LoginResponse{}

	loginValidation := ValidateResponse(
		loginRes,
		validateLoginUser(loginResponse, user),
	)
	loginValidation.Add(step)

	if loginValidation.IsEmpty() {
		return nil
	} else {
		return loginValidation
	}
}

// ベンチ実行前の整合性検証シナリオ
// isucandar.ValidateScenarioを満たすメソッド
// isucandar.Benchmark の PrePare ステップで実行される

func (sc *Scenario) PretestScenario(ctx context.Context, step *isucandar.BenchmarkStep) error {
	report := TimeReporter("pretest", sc.Option)
	defer report()
	ContestantLogger.Println("[PretestScenario] 整合性チェックを開始します")
	defer ContestantLogger.Printf("[PretestScenario] 整合性チェックを終了します")

	// User 取り出し
	var user *User
	for {
		trial := rand.Intn(sc.Users.Len()-1) + 1 // 0 は admin なので除外
		if !sc.ConsumedUserIDs.Exists(int64(trial)) {
			sc.ConsumedUserIDs.Add(int64(trial))
			user = sc.Users.At(trial)
			break
		}
	}

	// ログイン
	ContestantLogger.Println("整合性チェック Request:POST /login")
	if err := sc.loginValidateSuccessScenario(ctx, step, user); err != nil {
		return err
	}

	// データを登録、反映されるかチェック

	// Admin 取り出し

	// etc...

	return nil
}
