package main

// リクエストを送信する際にリクエストボディに詰める JSON を書くファイル

type RegisterRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Password    string `json:"password"` // ハッシュ化されていない
}

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"` // ハッシュ化されていない
}

type CreateTeamRequest struct {
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	Description    string `json:"description"`
}

type JoinTeamRequest struct {
	TeamName       string `json:"team_name"`
	InvitationCode string `json:"invitation_code"`
}

type SubmitRequest struct {
	TaskName  string `json:"task_name"`
	Answer    string `json:"answer"`
	Timestamp int64  `json:"timestamp"`
}

type AnswerRequest struct {
	Answer string `json:"answer"`
	Score  int    `json:"score"`
}
type SubtaskRequest struct {
	Name        string          `json:"name"`
	DisplayName string          `json:"display_name"`
	Statement   string          `json:"statement"`
	Answers     []AnswerRequest `json:"answers"`
}
type CreateTaskRequest struct {
	Name            string           `json:"name"`
	DisplayName     string           `json:"display_name"`
	Statement       string           `json:"statement"`
	SubmissionLimit int              `json:"submission_limit"`
	Subtasks        []SubtaskRequest `json:"subtasks"`
}
