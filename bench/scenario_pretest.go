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

func (s *Scenario) getuserValidateScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User) error {
	report := TimeReporter("user 取得 整合性チェック", s.Option)
	defer report()

	team, _ := s.Teams.Get(user.TeamID)

	agent, err := s.GetAgentFromUser(step, user)

	if err != nil {
		return err
	}
	getuserRes, err := GetUserAction(ctx, agent, user.Name)
	if err != nil {
		return failure.NewError(ValidationErrInvalidRequest, err)
	}
	defer getuserRes.Body.Close()

	getuserResponse := &UserResponse{}

	getuserValidation := ValidateResponse(
		getuserRes,
		validategetUser(getuserResponse, user, team),
	)
	getuserValidation.Add(step)

	if getuserValidation.IsEmpty() {
		return nil
	} else {
		return getuserValidation
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
		trial := rand.Intn(sc.Users.Len()-1) + 2 // id 1 は admin なので除外
		if !sc.ConsumedUserIDs.Exists(int64(trial)) {
			sc.ConsumedUserIDs.Add(int64(trial))
			user, _ = sc.Users.Get(trial)
			break
		}
	}

	// 一般ユーザー
	// ログイン
	if err := sc.loginValidateSuccessScenario(ctx, step, user); err != nil {
		return err
	}
	// user
	if err := sc.getuserValidateScenario(ctx, step, user); err != nil {
		return err
	}
	// データを登録、反映されるかチェック

	// Admin 取り出し

	// etc...

	return nil
}
