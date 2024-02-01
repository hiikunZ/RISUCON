package initialdata_generator

import (
	"time"
)

const (
	nulluserid = -1
)

type User struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	DisplayName string `db:"display_name"`
	Description string `db:"description"`
	Passhash    string `db:"passhash"`
}

type Team struct {
	ID             int    `db:"id"`
	Name           string `db:"name"`
	DisplayName    string `db:"display_name"`
	LeaderID       int    `db:"leader_id"`
	Member1ID      int    `db:"member1_id"`
	Member2ID      int    `db:"member2_id"`
	Description    string `db:"description"`
	InvitationCode string `db:"invitation_code"`
}

type Task struct {
	ID              int    `db:"id"`
	Name            string `db:"name"`
	DisplayName     string `db:"display_name"`
	Statement       string `db:"statement"`
	SubmissionLimit int    `db:"submission_limit"`
}

type Subtask struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	DisplayName string `db:"display_name"`
	TaskID      int    `db:"task_id"`
	Statement   string `db:"statement"`
}

type Answer struct {
	ID        int    `db:"id"`
	TaskID    int    `db:"task_id"`
	SubtaskID int    `db:"subtask_id"`
	Answer    string `db:"answer"`
	Score     int    `db:"score"`
}

type Submission struct {
	ID          int       `db:"id"`
	TaskID      int       `db:"task_id"`
	UserID      int       `db:"user_id"`
	SubmittedAt time.Time `db:"submitted_at"`
	Answer      string    `db:"answer"`
}
