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

	var team *Team = nil

	if user.TeamID != nullteamid {
		team, _ = s.Teams.Get(user.TeamID)
	}

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
		validateLoginUser(loginResponse, user, team),
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

	var team *Team = nil

	if user.TeamID != nullteamid {
		team, _ = s.Teams.Get(user.TeamID)
	}

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

func (s *Scenario) getteamValidateScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User) error {
	if user.TeamID == nullteamid {
		return nil
	}

	report := TimeReporter("team 取得 整合性チェック", s.Option)
	defer report()

	team, _ := s.Teams.Get(user.TeamID)
	leader, _ := s.Users.Get(team.LeaderID)

	var member1, member2 *User = nil, nil
	if team.Member1ID != nulluserid {
		member1, _ = s.Users.Get(team.Member1ID)
	}
	if team.Member2ID != nulluserid {
		member2, _ = s.Users.Get(team.Member2ID)
	}

	agent, err := s.GetAgentFromUser(step, user)

	if err != nil {
		return err
	}
	getteamRes, err := GetTeamAction(ctx, agent, team.Name)
	if err != nil {
		return failure.NewError(ValidationErrInvalidRequest, err)
	}
	defer getteamRes.Body.Close()

	getteamResponse := &TeamResponse{}

	getteamValidation := ValidateResponse(
		getteamRes,
		validategetTeam(getteamResponse, team, leader, member1, member2, user.Name == leader.Name),
	)
	getteamValidation.Add(step)

	if getteamValidation.IsEmpty() {
		return nil
	} else {
		return getteamValidation
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

	checkuserIDs := []int{2, 4, 10}
	for cnt := 0; cnt < 4; cnt++ {
		// User 取り出し
		var user *User
		for {
			var trial int
			if cnt < 3 {
				trial = checkuserIDs[cnt] // 仕様通りかののためなので、決め打ち
			} else {
				// データが消されていないかのチェックなので、ランダム
				trial = rand.Intn(sc.Users.Len()-1) + 2 // id 1 は admin なので除外
			}
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
		// team
		if err := sc.getteamValidateScenario(ctx, step, user); err != nil {
			return err
		}
		// submit
		// submission
		// submission 検索
		// logout
		sc.ConsumedUserIDs.Remove(int64(user.ID))
	}
	// 情報がこわれてないか
	// 非ログインユーザー
	// tasks
	// task
	// standings
	// submission が見れないことを確認

	// Admin 取り出し

	// createtask
	// 反映されているかのチェック
	// tasks
	// task
	// submission (admin)
	// submission 検索 (admin)

	// ユーザー、チームの新規作成
	// register (もう存在するユーザーで失敗)
	// register
	// login (名前の typo で失敗)
	// login (パスワードの typo で失敗)
	// create
	// join (一杯のチームに入れないことを確認)
	// join
	// もうチームに所属しているのに create できないことを確認
	// もうチームに所属しているのに join できないことを確認

	return nil
}
