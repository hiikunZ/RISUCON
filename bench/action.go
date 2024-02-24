package main

// リクエストを送る動作 "Action" を中心に集約しているファイル

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/isucon/isucandar/agent"
)

// POST /api/initialize にリクエストを送る
func PostInitializeAction(ctx context.Context, agent *agent.Agent) (*http.Response, error) {
	req, err := agent.POST("/api/initialize", nil)
	if err != nil {
		return nil, err
	}

	setContentType(req)

	return agent.Do(ctx, req)
}

// POST /api/register
func PostRegisterAction(ctx context.Context, agent *agent.Agent, username string, userdisplayname string, password string) (*http.Response, error) {
	json, err := json.Marshal(RegisterRequest{Name: username, DisplayName: userdisplayname, Password: password})
	if err != nil {
		return nil, err
	}

	req, err := agent.POST("/api/register", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	setContentType(req)

	return agent.Do(ctx, req)
}

// POST /api/login
func PostLoginAction(ctx context.Context, agent *agent.Agent, username string, password string) (*http.Response, error) {
	json, err := json.Marshal(LoginRequest{Name: username, Password: password})
	if err != nil {
		// このエラーは実装上の問題でエラーになるはずなので、もし送出される場合は何かがおかしい。
		return nil, err
	}

	req, err := agent.POST("/api/login", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	setContentType(req)

	return agent.Do(ctx, req)
}

// POST /api/logout
func PostLogoutAction(ctx context.Context, agent *agent.Agent) (*http.Response, error) {
	req, err := agent.POST("/api/logout", nil)
	if err != nil {
		return nil, err
	}
	setContentType(req)

	return agent.Do(ctx, req)
}

// GET /api/user/:username
func GetUserAction(ctx context.Context, agent *agent.Agent, username string) (*http.Response, error) {
	req, err := agent.GET("/api/user/" + username)
	if err != nil {
		return nil, err
	}
	setContentType(req)

	return agent.Do(ctx, req)
}

// POST /api/team/create
func PostCreateTeamAction(ctx context.Context, agent *agent.Agent, teamname string, teamdisplayname string, teamdescription string) (*http.Response, error) {
	json, err := json.Marshal(CreateTeamRequest{Name: teamname, DisplayName: teamdisplayname, Description: teamdescription})
	if err != nil {
		return nil, err
	}

	req, err := agent.POST("/api/team/create", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	setContentType(req)

	return agent.Do(ctx, req)
}

// POST /api/team/join
func PostJoinTeamAction(ctx context.Context, agent *agent.Agent, teamname string, invitationcode string) (*http.Response, error) {
	json, err := json.Marshal(JoinTeamRequest{TeamName: teamname, InvitationCode: invitationcode})
	if err != nil {
		return nil, err
	}

	req, err := agent.POST("/api/team/join", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	setContentType(req)

	return agent.Do(ctx, req)
}

// GET /api/team/:teamname
func GetTeamAction(ctx context.Context, agent *agent.Agent, teamname string) (*http.Response, error) {
	req, err := agent.GET("/api/team/" + teamname)
	if err != nil {
		return nil, err
	}
	setContentType(req)

	return agent.Do(ctx, req)
}

// GET /api/tasks
func GetTasksAction(ctx context.Context, agent *agent.Agent) (*http.Response, error) {
	req, err := agent.GET("/api/tasks")
	if err != nil {
		return nil, err
	}
	setContentType(req)

	return agent.Do(ctx, req)
}

// GET /api/standings
func GetStandingsAction(ctx context.Context, agent *agent.Agent) (*http.Response, error) {
	req, err := agent.GET("/api/standings")
	if err != nil {
		return nil, err
	}
	setContentType(req)

	return agent.Do(ctx, req)
}

// GET /api/tasks/:taskname
func GetTaskAction(ctx context.Context, agent *agent.Agent, taskname string) (*http.Response, error) {
	req, err := agent.GET("/api/tasks/" + taskname)
	if err != nil {
		return nil, err
	}
	setContentType(req)

	return agent.Do(ctx, req)
}

// POST /api/submit
func PostSubmitAction(ctx context.Context, agent *agent.Agent, taskname string, answer string, timestamp int64) (*http.Response, error) {
	json, err := json.Marshal(SubmitRequest{TaskName: taskname, Answer: answer, Timestamp: timestamp})
	if err != nil {
		return nil, err
	}

	req, err := agent.POST("/api/submit", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	setContentType(req)

	return agent.Do(ctx, req)
}

// GET /api/submissions
func GetSubmissionsAction(ctx context.Context, agent *agent.Agent, page int, username string, teamname string, taskname string, subtaskname string, answerfilter string) (*http.Response, error) {
	url := "/api/submissions"
	param := []string{}

	if page != 1 {
		param = append(param, fmt.Sprintf("page=%d", page))
	}
	if username != "" {
		param = append(param, fmt.Sprintf("user_name=%s", username))
	}
	if teamname != "" {
		param = append(param, fmt.Sprintf("team_name=%s", teamname))
	}
	if taskname != "" {
		param = append(param, fmt.Sprintf("task_name=%s", taskname))
	}
	if subtaskname != "" {
		param = append(param, fmt.Sprintf("subtask_name=%s", subtaskname))
	}
	if answerfilter != "" {
		param = append(param, fmt.Sprintf("filter=%s", answerfilter))
	}

	if len(param) > 0 {
		url += "?" + strings.Join(param, "&")
	}

	req, err := agent.GET(url)
	if err != nil {
		return nil, err
	}
	setContentType(req)

	return agent.Do(ctx, req)
}

// POST /api/admin/createtask
func PostCreateTaskAction(ctx context.Context, agent *agent.Agent, reqdata CreateTaskRequest) (*http.Response, error) {
	json, err := json.Marshal(reqdata)
	if err != nil {
		return nil, err
	}

	req, err := agent.POST("/api/admin/createtask", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	setContentType(req)

	return agent.Do(ctx, req)
}
func setContentType(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
}
