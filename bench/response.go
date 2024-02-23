package main

// リクエストの結果返ってくる JSON レスポンスを集約するファイル

type InitializeResponse struct {
	Language string `json:"language"`
}

type LoginResponse struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type UserResponse struct {
	Name            string `json:"name"`
	DisplayName     string `json:"display_name"`
	Description     string `json:"description"`
	SubmissionCount int    `json:"submission_count"`
	TeamName        string `json:"team_name"`
	TeamDisplayName string `json:"team_display_name"`
}