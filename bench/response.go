package main

// リクエストの結果返ってくる JSON レスポンスを集約するファイル

type InitializeResponse struct {
	Language string `json:"language"`
}

type LoginResponse struct {
	Name            string `json:"name"`
	DisplayName     string `json:"display_name"`
	TeamName        string `json:"team_name,omitempty"`
	TeamDisplayName string `json:"team_display_name,omitempty"`
}

type UserResponse struct {
	Name            string `json:"name"`
	DisplayName     string `json:"display_name"`
	Description     string `json:"description"`
	SubmissionCount int    `json:"submission_count"`
	TeamName        string `json:"team_name"`
	TeamDisplayName string `json:"team_display_name"`
}

type TeamResponse struct {
	Name               string `json:"name"`
	DisplayName        string `json:"display_name"`
	LeaderName         string `json:"leader_name"`
	LeaderDisplayName  string `json:"leader_display_name"`
	Member1Name        string `json:"member1_name,omitempty"`
	Member1DisplayName string `json:"member1_display_name,omitempty"`
	Member2Name        string `json:"member2_name,omitempty"`
	Member2DisplayName string `json:"member2_display_name,omitempty"`
	Description        string `json:"description"`
	SubmissionCount    int    `json:"submission_count"`
	InvitationCode     string `json:"invitation_code,omitempty"`
}

type TaskAbstract struct {
	Name            string `json:"name"`
	DisplayName     string `json:"display_name"`
	MaxScore        int    `json:"max_score"`
	Score           int    `json:"score,omitempty"`
	SubmissionLimit int    `json:"submission_limit,omitempty"`
	SubmissionCount int    `json:"submission_count,omitempty"`
}

type SubmitResponse struct {
	IsScored             bool   `json:"is_scored"`
	Score                int    `json:"score"`
	SubtaskName          string `json:"subtask_name"`
	SubTaskDisplayName   string `json:"subtask_display_name"`
	SubTaskMaxScore      int    `json:"subtask_max_score"`
	RemainingSubmissions int    `json:"remaining_submissions"`
}
