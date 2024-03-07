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

type TeamsStandingsSub struct {
	TaskName     string `json:"task_name"`
	HasSubmitted bool   `json:"has_submitted"`
	Score        int    `json:"score"`
}
type TeamsStandings struct {
	Rank               int                 `json:"rank"`
	TeamName           string              `json:"team_name"`
	TeamDisplayName    string              `json:"team_display_name"`
	LeaderName         string              `json:"leader_name"`
	LeaderDisplayName  string              `json:"leader_display_name"`
	Member1Name        string              `json:"member1_name,omitempty"`
	Member1DisplayName string              `json:"member1_display_name,omitempty"`
	Member2Name        string              `json:"member2_name,omitempty"`
	Member2DisplayName string              `json:"member2_display_name,omitempty"`
	ScoringData        []TeamsStandingsSub `json:"scoring_data"`
	TotalScore         int                 `json:"total_score"`
}
type Standings struct {
	TasksData     []TaskAbstract   `json:"tasks_data"`
	StandingsData []TeamsStandings `json:"standings_data"`
}

type JoinTeamResponse struct {
	TeamName        string `json:"team_name"`
	TeamDisplayName string `json:"team_display_name"`
}

type SubmissionDetail struct {
	TaskName           string `json:"task_name"`
	TaskDisplayName    string `json:"task_display_name"`
	SubTaskName        string `json:"subtask_name"`
	SubTaskDisplayName string `json:"subtask_display_name"`
	SubTaskMaxScore    int    `json:"subtask_max_score"`
	UserName           string `json:"user_name"`
	UserDisplayName    string `json:"user_display_name"`
	SubmittedAt        int64  `json:"submitted_at"`
	Answer             string `json:"answer"`
	Score              int    `json:"score"`
}

type submissionresponse struct {
	Submissions     []SubmissionDetail `json:"submissions"`
	SubmissionCount int                `json:"submission_count"`
}

type SubtaskDetail struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Statement   string `json:"statement"`
	MaxScore    int    `json:"max_score"`
	Score       int    `json:"score"`
}
type TaskDetail struct {
	Name            string          `json:"name"`
	DisplayName     string          `json:"display_name"`
	Statement       string          `json:"statement"`
	MaxScore        int             `json:"max_score"`
	Score           int             `json:"score"`
	SubmissionLimit int             `json:"submission_limit"`
	SubmissionCount int             `json:"submission_count"`
	Subtasks        []SubtaskDetail `json:"subtasks"`
}
