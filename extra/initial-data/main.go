package main

import (
	"math/rand"
	"time"
)

const (
	TeamCount         = 50
	SubmissionPerUser = 3
	validanswerprob   = 0.8
	defaulttasksfile  = "default_tasks.json"
	maxtaskcnt        = 6
)

var (
	rng = rand.New(rand.NewSource(634))
)

func main() {
	tasks := LoadDeafultTasks(defaulttasksfile)
	users := []User{}
	teams := []Team{}
	submissions := []Submission{}

	// Admin
	users = append(users, User{
		ID:            1,
		Name:          "admin",
		DisplayName:   "管理者",
		Description:   "管理者アカウントです。",
		Password:      "admin",
		SubmissionIDs: []int{},
		TeamID:        nullteamid,
	})

	// テスト用ユーザー
	users = append(users, User{
		ID:            2,
		Name:          "risucon",
		DisplayName:   "risucon",
		Description:   "テスト用アカウントです。",
		Password:      "risucon",
		SubmissionIDs: []int{},
		TeamID:        1,
	})

	// テスト用チーム
	teams = append(teams, Team{
		ID:             1,
		Name:           "risucon",
		DisplayName:    "risucon",
		LeaderID:       2,
		Member1ID:      nulluserid,
		Member2ID:      nulluserid,
		Description:    "テスト用チームです。",
		InvitationCode: "72697375636f6e21",
	})
	teams[0].SubmissionCounts = make([]int, maxtaskcnt)

	for i := 0; i < TeamCount; i++ {
		usercnt := rng.Intn(3) + 1
		team := Teamgen()
		team.ID = len(teams) + 1
		leader := Usergen()
		leader.ID = len(users) + 1
		leader.TeamID = team.ID
		users = append(users, leader)
		team.LeaderID = leader.ID
		team.SubmissionCounts = make([]int, maxtaskcnt)

		if usercnt > 1 {
			member1 := Usergen()
			member1.ID = len(users) + 1
			member1.TeamID = team.ID
			users = append(users, member1)
			team.Member1ID = member1.ID
			if usercnt > 2 {
				member2 := Usergen()
				member2.ID = len(users) + 1
				member2.TeamID = team.ID
				users = append(users, member2)
				team.Member2ID = member2.ID
			}
		}
		teams = append(teams, team)
	}

	// 提出
	answerletters := []string{
		"あ", "い", "う", "え", "お", "か", "き", "く", "け", "こ", "さ", "し", "す", "せ", "そ", "た", "ち", "つ", "て", "と", "な", "に", "ぬ", "ね", "の", "は", "ひ", "ふ", "へ", "も", "ま", "み", "む", "め", "も", "や", "ゆ", "よ", "ら", "り", "る", "れ", "ろ", "わ", "を", "ん",
	}
	for i := 1; i < len(users); i++ {
		user := users[i]
		for j := 0; j < SubmissionPerUser; j++ {
			task := tasks[rng.Intn(len(tasks))]
			submission := Submission{
				TaskID: task.ID,
				UserID: user.ID,
			}
			if rng.Float64() < validanswerprob {
				// 正解
				subtask := task.SubTasks[rng.Intn(len(task.SubTasks))]
				answer := subtask.Answers[rng.Intn(len(subtask.Answers))]
				submission.Answer = answer.Answer
				submission.Score = answer.Score
				submission.SubTaskid = subtask.ID
			} else {
				// 不正解
				answer := ""
				anslen := rng.Intn(10) + 10
				for k := 0; k < anslen; k++ {
					answer += answerletters[rng.Intn(len(answerletters))]
				}
				submission.Answer = answer
				submission.Score = 0
				submission.SubTaskid = -1
			}
			submissions = append(submissions, submission)
		}
	}

	// submissions を シャッフル
	rng.Shuffle(len(submissions), func(i, j int) {
		submissions[i], submissions[j] = submissions[j], submissions[i]
	})

	// id と submitted_at を設定 (JST)
	time.Local = time.FixedZone("Asia/Tokyo", 9*60*60) // JST
	firstsubtime := time.Date(2024, 3, 26, 18, 0, 0, 0, time.Local)
	for i := 0; i < len(submissions); i++ {
		submissions[i].ID = i + 1
		submissions[i].SubmittedAt = firstsubtime.Add(time.Duration(i) * time.Second)
		submituserid := submissions[i].UserID
		users[submituserid-1].SubmissionIDs = append(users[submituserid-1].SubmissionIDs, i+1)
		teams[users[submituserid-1].TeamID-1].SubmissionIDs = append(teams[users[submituserid-1].TeamID-1].SubmissionIDs, i+1)
		teams[users[submituserid-1].TeamID-1].SubmissionCounts[submissions[i].TaskID-1]++
	}

	DumpUserstoJSON(users)
	DumpTeamstoJSON(teams)
	DumpSubmissionstoJSON(submissions)
	DumpDatatoSQL(users, teams, submissions, tasks)
}
