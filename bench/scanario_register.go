package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/worker"
)

// ユーザー/チーム登録のシナリオを一通り管理するファイル。

func (s *Scenario) NewUserRegistrationScenarioWorker(step *isucandar.BenchmarkStep, p int32) (*worker.Worker, error) {
	// あとで実装
	userRegistration, err := worker.NewWorker(func(ctx context.Context, _ int) {
		PrintScenarioStarted(ScenarioUserRegistration)
		defer PrintScenarioFinished(ScenarioUserRegistration)
		time.Sleep(1 * time.Second)
		// データ作成
		// team := Teamgen()
		teammembercount := rand.Intn(3) + 2
		var leader, member1, member2 User

		leader = Usergen()
		if teammembercount >= 2 {
			member1 = Usergen()
			if teammembercount >= 3 {
				member2 = Usergen()
			}
		}
		// 静的ファイル * register + login
		s.GetIndexScenario(ctx, step, &leader)
		if ok := s.RegisterSuccessScenario(ctx, step, &leader); !ok {
			return
		}
		if ok := s.LoginSuccessScenario(ctx, step, &leader); !ok {
			return
		}
		if teammembercount >= 2 {
			s.GetIndexScenario(ctx, step, &member1)
			if ok := s.RegisterSuccessScenario(ctx, step, &member1); !ok {
				return
			}
			if ok := s.LoginSuccessScenario(ctx, step, &member1); !ok {
				return
			}
			if teammembercount >= 3 {
				s.GetIndexScenario(ctx, step, &member2)
				if ok := s.RegisterSuccessScenario(ctx, step, &member2); !ok {
					return
				}
				if ok := s.LoginSuccessScenario(ctx, step, &member2); !ok {
					return
				}
			}
		}
		// createteam
		team := Teamgen()
		team.SubmissionCounts = make([]int, s.Tasks.Len())
		if ok := s.CreateTeamSuccessScenario(ctx, step, &leader, &team); !ok {
			return
		}
		code, ok := s.GetTeam_GetTokenScenario(ctx, step, &leader, &team)
		if !ok {
			return
		}
		team.InvitationCode = code
		if !ok {
			return
		}
		// join
		if teammembercount >= 2 {
			if ok := s.TeamJoinScenario(ctx, step, &member1, &team); !ok {
				return
			}
			if teammembercount >= 3 {
				if ok := s.TeamJoinScenario(ctx, step, &member2, &team); !ok {
					return
				}
			}
		}
		// データを保存
		s.Users.mu.Lock()
		defer s.Users.mu.Unlock()
		leader.ID = s.Users.Len_without_lock() + 1
		team.LeaderID = leader.ID
		if teammembercount >= 2 {
			member1.ID = s.Users.Len_without_lock() + 2
			team.Member1ID = member1.ID
			if teammembercount >= 3 {
				member2.ID = s.Users.Len_without_lock() + 3
				team.Member2ID = member2.ID
			}
		}
		s.Teams.mu.Lock()
		defer s.Teams.mu.Unlock()
		team.ID = s.Teams.Len_without_lock() + 1
		leader.TeamID = team.ID
		s.Users.Add_without_lock(&leader)
		if teammembercount >= 2 {
			member1.TeamID = team.ID
			s.Users.Add_without_lock(&member1)
			if teammembercount >= 3 {
				member2.TeamID = team.ID
				s.Users.Add_without_lock(&member2)
			}
		}
		s.Teams.Add_without_lock(&team)
		s.RecordUserRegistrationCount(teammembercount)
	}, loopConfig(s), parallelismConfig(s))

	userRegistration.SetParallelism(p)

	return userRegistration, err
}

func (s *Scenario) RegisterSuccessScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User) bool {
	report := TimeReporter("ユーザー 登録 シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return false
	}

	registerres, err := PostRegisterAction(ctx, agent, user.Name, user.DisplayName, user.Description, user.Password)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return false
	}
	defer registerres.Body.Close()

	registerValidation := ValidateResponse(
		registerres,
		WithStatusCode(http.StatusCreated),
	)
	registerValidation.Add(step)

	if registerValidation.IsEmpty() {
		step.AddScore(ScoreRegisteration)
		return true
	} else {
		return false
	}
}

func (s *Scenario) CreateTeamSuccessScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User, team *Team) bool {
	report := TimeReporter("チーム 登録 シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return false
	}

	createres, err := PostCreateTeamAction(ctx, agent, team.Name, team.DisplayName, team.Description)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return false
	}
	defer createres.Body.Close()

	createValidation := ValidateResponse(
		createres,
		WithStatusCode(http.StatusCreated),
	)
	createValidation.Add(step)

	if createValidation.IsEmpty() {
		return true
	} else {
		return false
	}
}

func (s *Scenario) GetTeam_GetTokenScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User, team *Team) (string, bool) {
	report := TimeReporter("team 取得 シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return "", false
	}

	teamRes, err := GetTeamAction(ctx, agent, team.Name)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return "", false
	}
	defer teamRes.Body.Close()

	teamResponse := &TeamResponse{}

	teamValidation := ValidateResponse(
		teamRes,
		WithStatusCode(200),
		WithJsonBody(teamResponse),
	)
	teamValidation.Add(step)

	if teamValidation.IsEmpty() {
		if teamResponse.InvitationCode != "" {
			return teamResponse.InvitationCode, true
		}
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidResponse, fmt.Errorf("POST /api/teams/%v : InvitationCode が存在しません", team.Name)))
		return "", false
	} else {
		return "", false
	}
}

func (s *Scenario) TeamJoinScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User, team *Team) bool {
	report := TimeReporter("team join シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return false
	}

	joinRes, err := PostJoinTeamAction(ctx, agent, team.Name, team.InvitationCode)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return false
	}
	defer joinRes.Body.Close()

	joinResponse := &JoinTeamResponse{}

	joinValidation := ValidateResponse(
		joinRes,
		WithStatusCode(http.StatusCreated),
		WithJsonBody(joinResponse),
	)
	joinValidation.Add(step)

	if joinValidation.IsEmpty() {
		return true
	} else {
		return false
	}
}
