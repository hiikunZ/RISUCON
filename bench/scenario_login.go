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

		// ログイン
		ok := s.LoginSuccessScenario(ctx, step, user)
		if !ok {
			return
		}
		// user / team 取得 (確率的)
		if rand.Float64() < 0.2 {
			s.GetUserSuccessScenario(ctx, step, user)
		}
		if rand.Float64() < 0.2 {
			s.GetTeamSuccessScenario(ctx, step, user, team.Name)
		}
		// 以下、何回かやる
		tasksubmitcnt := rand.Intn(3) + 1
		for i := 0; i < tasksubmitcnt; i++ {
			// tasks
			s.GetTasksSuccessScenario(ctx, step, user)

			// task
			task := s.ChooseTask(team)
			s.GetTaskSuccessScenario(ctx, step, user, task)

			// submit
			s.PostSubmitScenario(ctx, step, user, team, task)
		}
		// submission (確率的)
		if rand.Float64() < 0.5 {
			s.GetSubmissionsScenario(ctx, step, user, team)
		}
		// standings (確率的)
		if rand.Float64() < 0.5 {
			s.GetStandingsSuccessScenario(ctx, step, user)
		}
		// logout
		s.LogoutSuccessScenario(ctx, step, user)
		// ここまでできたら成功
		s.RecordLoginSuccessCount(1)

		user.ClearAgent()
	}, loopConfig(s), parallelismConfig(s))

	loginSuccess.SetParallelism(p)

	return loginSuccess, err

}

// リクエストを送ってステータスコードが成功状態であることと、レスポンスボディの形式が正しいかを確認する。
func (s *Scenario) LoginSuccessScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User) bool {
	report := TimeReporter("ログイン成功シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return false
	}

	loginRes, err := PostLoginAction(ctx, agent, user.Name, user.Password)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return false
	}
	defer loginRes.Body.Close()

	loginResponse := &LoginResponse{}

	loginValidation := ValidateResponse(
		loginRes,
		WithStatusCode(200),
		WithJsonBody(loginResponse),
	)
	loginValidation.Add(step)

	if loginValidation.IsEmpty() {
		return true
	} else {
		return false
	}
}

func (s *Scenario) GetUserSuccessScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User) {
	report := TimeReporter("user 取得 シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return
	}

	getuserRes, err := GetUserAction(ctx, agent, user.Name)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return
	}
	defer getuserRes.Body.Close()

	getuserResponse := &UserResponse{}

	getuserValidation := ValidateResponse(
		getuserRes,
		WithStatusCode(200),
		WithJsonBody(getuserResponse),
	)
	getuserValidation.Add(step)

}

func (s *Scenario) GetTeamSuccessScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User, teamname string) {
	report := TimeReporter("team 取得 シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return
	}

	getteamRes, err := GetTeamAction(ctx, agent, teamname)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return
	}
	defer getteamRes.Body.Close()

	getteamResponse := &TeamResponse{}

	getteamValidation := ValidateResponse(
		getteamRes,
		WithStatusCode(200),
		WithJsonBody(getteamResponse),
	)
	getteamValidation.Add(step)
}

func (s *Scenario) GetTasksSuccessScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User) {
	report := TimeReporter("tasks 取得 シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return
	}

	gettasksRes, err := GetTasksAction(ctx, agent)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return
	}
	defer gettasksRes.Body.Close()

	gettasksResponse := &[]TaskAbstract{}

	gettasksValidation := ValidateResponse(
		gettasksRes,
		WithStatusCode(200),
		WithJsonBody(gettasksResponse),
	)
	gettasksValidation.Add(step)
}

func (s *Scenario) GetTaskSuccessScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User, task *Task) {
	report := TimeReporter("task 取得 シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return
	}

	gettaskRes, err := GetTaskAction(ctx, agent, task.Name)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return
	}
	defer gettaskRes.Body.Close()

	gettaskResponse := &TaskDetail{}

	gettaskValidation := ValidateResponse(
		gettaskRes,
		WithStatusCode(200),
		WithJsonBody(gettaskResponse),
	)
	gettaskValidation.Add(step)

}

func (s *Scenario) PostSubmitScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User, team *Team, task *Task) {
	report := TimeReporter("submit シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return
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
		return
	}
	defer submitRes.Body.Close()

	// 一杯
	if team.SubmissionCounts[task.ID-1] >= task.SubmissionLimit {
		submitValidation := ValidateResponse(
			submitRes,
			WithStatusCode(http.StatusBadRequest),
		)
		submitValidation.Add(step)

	} else {

		submitresponse := &SubmitResponse{}

		submitValidation := ValidateResponse(
			submitRes,
			WithStatusCode(http.StatusCreated),
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
		}
	}
}

func (s *Scenario) GetSubmissionsScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User, team *Team) {
	report := TimeReporter("submissions 取得 シナリオ", s.Option)
	defer report()

	submissionsperpage := 20

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return
	}
	
	page := 1
	if len(team.SubmissionIDs) > submissionsperpage && rand.Float64() < 0.05 {
		page = 2
	}

	username := ""
	if rand.Float64() < 0.3 {
		memberids := []int{}
		memberids = append(memberids, team.LeaderID)
		if team.Member1ID != nullteamid {
			memberids = append(memberids, team.Member1ID)
			if team.Member2ID != nullteamid {
				memberids = append(memberids, team.Member2ID)
			}
		}
		user, ok := s.Users.Get(memberids[rand.Intn(len(memberids))])
		if !ok {
			return
		}
		username = user.Name
	}
	taskname := ""
	subtaskname := ""
	if rand.Float64() < 0.3 {
		task := s.Tasks.At(rand.Intn(s.Tasks.Len()))
		taskname = task.Name
		if rand.Float64() < 0.3 {
			subtaskname = task.SubTasks[rand.Intn(len(task.SubTasks))].Name
		}
	}

	answerfilter := ""
	if rand.Float64() < 0.1 {
		sub, ok := s.Submissions.Get(team.SubmissionIDs[rand.Intn(len(team.SubmissionIDs))])
		if !ok {
			return
		}
		answerfilter = sub.Answer
		if rand.Float64() < 0.3 {
			l := rand.Intn(len(answerfilter))
			r := rand.Intn(len(answerfilter))
			if l > r {
				l, r = r, l
			}
			answerfilter = answerfilter[l : r+1]
		}
	}

	getsubmissionsRes, err := GetSubmissionsAction(ctx, agent, page, username, "", taskname, subtaskname, answerfilter)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return
	}
	defer getsubmissionsRes.Body.Close()

	getsubmissionsresponce := &submissionresponse{}

	getsubmissionsValidation := ValidateResponse(
		getsubmissionsRes,
		WithStatusCode(200),
		WithJsonBody(getsubmissionsresponce),
	)

	getsubmissionsValidation.Add(step)
}

func (s *Scenario) GetStandingsSuccessScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User) {
	report := TimeReporter("standings 取得 シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return
	}

	getstandingsRes, err := GetStandingsAction(ctx, agent)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return
	}
	defer getstandingsRes.Body.Close()

	getstandingsresponce := &Standings{}

	getstandingsValidation := ValidateResponse(
		getstandingsRes,
		WithStatusCode(200),
		WithJsonBody(getstandingsresponce),
	)
	getstandingsValidation.Add(step)
}

func (s *Scenario) LogoutSuccessScenario(ctx context.Context, step *isucandar.BenchmarkStep, user *User) {
	report := TimeReporter("ログアウト シナリオ", s.Option)
	defer report()

	agent, err := s.GetAgentFromUser(step, user)
	if err != nil {
		return
	}

	logoutRes, err := PostLogoutAction(ctx, agent)
	if err != nil {
		AddErrorIfNotCanceled(step, failure.NewError(ErrInvalidRequest, err))
		return
	}
	defer logoutRes.Body.Close()

	logoutValidation := ValidateResponse(
		logoutRes,
		WithStatusCode(200),
	)
	logoutValidation.Add(step)
}

func (s *Scenario) GetAgentFromUser(step *isucandar.BenchmarkStep, user *User) (*agent.Agent, error) {
	agent, err := user.GetAgent(s.Option)
	if err != nil {
		step.AddError(failure.NewError(ErrCannotCreateNewAgent, err))
		return nil, err
	}
	return agent, nil
}

func (s *Scenario) GetAgent(step *isucandar.BenchmarkStep) (*agent.Agent, error) {
	agent, err := s.Option.NewAgent(false)
	if err != nil {
		step.AddError(failure.NewError(ErrCannotCreateNewAgent, err))
		return nil, err
	}
	return agent, nil
}
