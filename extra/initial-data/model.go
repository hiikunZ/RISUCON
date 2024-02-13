package main

import (
	"time"
)

const (
	nulluserid = -1
	nullteamid = -1
)

type User struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`
	Description   string `json:"description"`
	Password      string `json:"password"` // Passhash は dump するときに計算する
	SubmissionIDs []int  `json:"submission_ids"`
	TeamID        int    `json:"team_id"`
}

type Team struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	LeaderID       int    `json:"leader_id"`
	Member1ID      int    `json:"member1_id"`
	Member2ID      int    `json:"member2_id"`
	Description    string `json:"description"`
	InvitationCode string `json:"invitation_code"`
}

type Task struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	DisplayName     string    `json:"display_name"`
	Statement       string    `json:"statement"`
	SubmissionLimit int       `json:"submission_limit"`
	SubTasks        []Subtask `json:"subtasks"`
	MaxScore        int       `json:"max_score"`
}

type Subtask struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	TaskID      int      `json:"task_id"`
	Statement   string   `json:"statement"`
	Answers     []Answer `json:"answers"`
	MaxScore    int      `json:"max_score"`
}

type Answer struct {
	ID        int    `json:"id"`
	TaskID    int    `json:"task_id"`
	SubtaskID int    `json:"subtask_id"`
	Answer    string `json:"answer"`
	Score     int    `json:"score"`
}

type Submission struct {
	ID          int       `json:"id"`
	TaskID      int       `json:"task_id"`
	UserID      int       `json:"user_id"`
	SubmittedAt time.Time `json:"submitted_at"`
	Answer      string    `json:"answer"`
	SubTaskid   int       `json:"subtask_id"`
	Score       int       `json:"score"`
}
