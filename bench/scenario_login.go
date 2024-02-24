package main

// ログイン処理〜解答提出までのシナリオを一通り管理するファイル。

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/worker"
)

const (
	validanswerprob = 0.8
)

func (s *Scenario) ChooseTask(team *Team) *Task {
	// lock 取得
	team.mu.RLock()
	defer team.mu.RUnlock()

	unfull_idx := []int{}

	for i := 0; i < s.Tasks.Len(); i++ {
		task := s.Tasks.At(i)
		if team.SubmissionCounts[task.ID-1] < task.SubmissionLimit {
			unfull_idx = append(unfull_idx, i)
		}
	}

	if len(unfull_idx) == 0 || rand.Float64() < 0.1 { // 10% の確率でランダムに選ぶ
		return s.Tasks.At(rand.Intn(s.Tasks.Len()))
	} else {
		return s.Tasks.At(unfull_idx[rand.Intn(len(unfull_idx))])
	}
}

func (s *Scenario) NewLoginScenarioWorker(step *isucandar.BenchmarkStep, p int32) (*worker.Worker, error) {
	// あとで実装
	loginSuccess, err := worker.NewWorker(func(ctx context.Context, _ int) {
		PrintScenarioStarted(ScenarioLogin)
		defer PrintScenarioFinished(ScenarioLogin)

		var user *User
		for {
			trial := rand.Intn(s.Users.Len()-1) + 2 // id 1 は admin なので除外
			if !s.ConsumedUserIDs.Exists(int64(trial)) {
				s.ConsumedUserIDs.Add(int64(trial))
				user, _ = s.Users.Get(trial)
				break
			}
		}
		defer s.ConsumedUserIDs.Remove(int64(user.ID))

		team, _ := s.Teams.Get(user.TeamID)

	Rewind:
		// ログイン
		result, ok := s.LoginSuccessScenario(ctx, step, user)
		if !ok {
			return
		}
		if result.Rewind {
			goto Rewind
		}
		// user / team 取得 (確率的)
		if rand.Float64() < 0.2 {
			result = s.GetUserSuccessScenario(ctx, step, user)
			if result.Rewind {
				goto Rewind
			}
		}
		if rand.Float64() < 0.2 {
			result = s.GetTeamSuccessScenario(ctx, step, user, team.Name)
			if result.Rewind {
				goto Rewind
			}
		}
		// 以下、何回かやる
		tasksubmitcnt := rand.Intn(3) + 1
		for i := 0; i < tasksubmitcnt; i++ {
			// tasks
			result = s.GetTasksSuccessScenario(ctx, step, user)
			if result.Rewind {
				goto Rewind
			}
			// task

			task := s.ChooseTask(team)
			result = s.GetTaskSuccessScenario(ctx, step, user, team, task)
			if result.Rewind {
				goto Rewind
			}

			// submit

			result = s.PostSubmitScenario(ctx, step, user, team, task)
			if result.Rewind {
				goto Rewind
			}

		}
		// submission (確率的)
		// standings (確率的)

		// logout

		// ここまでできたら成功
		s.RecordLoginSuccessCount(1)

		user.ClearAgent()
	}, loopConfig(s), parallelismConfig(s))

	loginSuccess.SetParallelism(p)

	return loginSuccess, err

}

// リクエストを送ってステータスコードが成功状態であることと、レスポンスボディの形式が正しいかを確認する。
func (s *Scenario) LoginSuccessScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User) (ScenarioResult, bool) {
	report := TimeReporter("ログイン成功シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return NoRewind(), false
	}

	loginRes, err := PostLoginAction(ctx, agent, user.Name, user.Password)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return NoRewind(), false
	}
	defer loginRes.Body.Close()

	if loginRes.StatusCode == http.StatusUnprocessableEntity {
		return Rewind(), false
	}

	loginResponse := &LoginResponse{}

	loginValidation := ValidateResponse(
		loginRes,
		WithStatusCode(200),
		WithJsonBody(loginResponse),
	)
	loginValidation.Add(step)

	if loginValidation.IsEmpty() {
		return NoRewind(), true
	} else {
		return NoRewind(), false
	}
}

func (s *Scenario) GetUserSuccessScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User) ScenarioResult {
	report := TimeReporter("user 取得 シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return NoRewind()
	}

	getuserRes, err := GetUserAction(ctx, agent, user.Name)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return NoRewind()
	}
	defer getuserRes.Body.Close()

	if getuserRes.StatusCode == http.StatusUnprocessableEntity {
		return Rewind()
	}

	getuserResponse := &UserResponse{}

	getuserValidation := ValidateResponse(
		getuserRes,
		WithStatusCode(200),
		WithJsonBody(getuserResponse),
	)
	getuserValidation.Add(step)

	if getuserValidation.IsEmpty() {
		return NoRewind()
	} else {
		return NoRewind()
	}
}

func (s *Scenario) GetTeamSuccessScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User, teamname string) ScenarioResult {
	report := TimeReporter("team 取得 シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return NoRewind()
	}

	getteamRes, err := GetTeamAction(ctx, agent, teamname)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return NoRewind()
	}
	defer getteamRes.Body.Close()

	if getteamRes.StatusCode == http.StatusUnprocessableEntity {
		return Rewind()
	}

	getteamResponse := &TeamResponse{}

	getteamValidation := ValidateResponse(
		getteamRes,
		WithStatusCode(200),
		WithJsonBody(getteamResponse),
	)
	getteamValidation.Add(step)

	if getteamValidation.IsEmpty() {
		return NoRewind()
	} else {
		return NoRewind()
	}
}

func (s *Scenario) GetTasksSuccessScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User) ScenarioResult {
	report := TimeReporter("tasks 取得 シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return NoRewind()
	}

	gettasksRes, err := GetTasksAction(ctx, agent)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return NoRewind()
	}
	defer gettasksRes.Body.Close()

	if gettasksRes.StatusCode == http.StatusUnprocessableEntity {
		return Rewind()
	}

	gettasksResponse := &[]TaskAbstract{}

	gettasksValidation := ValidateResponse(
		gettasksRes,
		WithStatusCode(200),
		WithJsonBody(gettasksResponse),
	)
	gettasksValidation.Add(step)

	if gettasksValidation.IsEmpty() {
		return NoRewind()
	} else {
		return NoRewind()
	}
}

func (s *Scenario) GetTaskSuccessScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User, team *Team, task *Task) ScenarioResult {
	report := TimeReporter("task 取得 シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return NoRewind()
	}

	gettaskRes, err := GetTaskAction(ctx, agent, task.Name)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return NoRewind()
	}
	defer gettaskRes.Body.Close()

	if gettaskRes.StatusCode == http.StatusUnprocessableEntity {
		return Rewind()
	}

	gettaskResponse := &Task{}

	gettaskValidation := ValidateResponse(
		gettaskRes,
		WithStatusCode(200),
		WithJsonBody(gettaskResponse),
	)
	gettaskValidation.Add(step)

	if gettaskValidation.IsEmpty() {
		return NoRewind()
	} else {
		return NoRewind()
	}
}

func (s *Scenario) PostSubmitScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User, team *Team, task *Task) ScenarioResult {
	report := TimeReporter("submit シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return NoRewind()
	}
	submission := &Submission{}
	// 全体で時間のロックを取るのは諦めました
	timediff := time.Duration(rand.Intn(10)+1) * time.Second
	s.NowClock = s.NowClock.Add(timediff)
	submission.SubmittedAt = s.NowClock
	submission.ID = int(rand.Int31())
	submission.TaskID = task.ID
	submission.UserID = user.ID
	if rand.Float64() < validanswerprob {
		// 正解の場合
		subtask := task.SubTasks[rand.Intn(len(task.SubTasks))]
		answer := subtask.Answers[rand.Intn(len(subtask.Answers))]
		submission.Answer = answer.Answer
		submission.Score = answer.Score
		submission.SubTaskID = subtask.ID
	} else {
		answerletters := []string{
			"あ", "い", "う", "え", "お", "か", "き", "く", "け", "こ", "さ", "し", "す", "せ", "そ", "た", "ち", "つ", "て", "と", "な", "に", "ぬ", "ね", "の", "は", "ひ", "ふ", "へ", "も", "ま", "み", "む", "め", "も", "や", "ゆ", "よ", "ら", "り", "る", "れ", "ろ", "わ", "を", "ん",
		}
		answer := ""
		anslen := rand.Intn(10) + 10
		for k := 0; k < anslen; k++ {
			answer += answerletters[rand.Intn(len(answerletters))]
		}
		submission.Answer = answer
		submission.Score = 0
		submission.SubTaskID = -1
	}
	// 送信前に lock
	team.mu.Lock()
	defer team.mu.Unlock()

	submitRes, err := PostSubmitAction(ctx, agent, task.Name, submission.Answer, submission.SubmittedAt.Unix())
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return NoRewind()
	}
	defer submitRes.Body.Close()

	if submitRes.StatusCode == http.StatusUnprocessableEntity {
		return Rewind()
	}

	// 一杯
	if team.SubmissionCounts[task.ID-1] >= task.SubmissionLimit {
		submitValidation := ValidateResponse(
			submitRes,
			WithStatusCode(http.StatusBadRequest),
		)
		submitValidation.Add(step)

		if submitValidation.IsEmpty() {
			return NoRewind()
		} else {
			return NoRewind()
		}
	} else {

		submitresponse := &SubmitResponse{}

		submitValidation := ValidateResponse(
			submitRes,
			WithStatusCode(200),
			WithJsonBody(submitresponse),
		)
		submitValidation.Add(step)

		if submitValidation.IsEmpty() {
			// 成功
			s.Submissions.Add(submission)
			team.SubmissionCounts[task.ID-1]++
			team.SubmissionIDs = append(team.SubmissionIDs, int(submission.ID))
			user.SubmissionIDs = append(user.SubmissionIDs, int(submission.ID))
			step.AddScore(ScoreSubmission)
			return NoRewind()
		} else {
			return NoRewind()
		}
	}
}

func (s *Scenario) GetAgentFromUser(step *isucandar.BenchmarkStep, user *User) (*agent.Agent, error) {
	agent, err := user.GetAgent(s.Option)
	if err != nil {
		step.AddError(failure.NewError(ErrCannotCreateNewAgent, err))
		return nil, err
	}
	return agent, nil
}
