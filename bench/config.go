package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/isucon/isucandar/agent"
)

const (
	PostLogin string = "POST /api/login"
	GetUser   string = "GET /api/user/:username"
	GetTeam   string = "GET /api/team/:teamname"
)

// 起動時の設定

type Option struct {
	TargetHost               string
	RequestTimeout           time.Duration
	InitializeRequestTimeout time.Duration
	ExitErrorOnFail          bool
	Stage                    string
	Parallelism              int
	PrepareOnly              bool
}

const (
	MaxErrors = 50
)

const (
	ScenarioLogin            string = "ログイン"
	ScenarioUserRegistration string = "新規ユーザー登録"
	ScenarioVisitor          string = "観戦"
)

func (o Option) String() string {
	args := []string{
		"benchmarker",
		fmt.Sprintf("--target-host=%s", o.TargetHost),
		fmt.Sprintf("--request-timeout=%s", o.RequestTimeout.String()),
		fmt.Sprintf("--initialize-request-timeout=%s", o.InitializeRequestTimeout.String()),
		fmt.Sprintf("--exit-error-on-fail=%v", o.ExitErrorOnFail),
		fmt.Sprintf("--stage=%s", o.Stage),
		// fmt.Sprintf("--max-parallelism=%d", o.Parallelism),
		// fmt.Sprintf("--prepare-only=%v", o.PrepareOnly),
	}

	return strings.Join(args, " ")
}

func (o Option) NewAgent(forInitialize bool) (*agent.Agent, error) {
	agentOptions := []agent.AgentOption{
		agent.WithBaseURL(fmt.Sprintf("http://%s/", o.TargetHost)),
		agent.WithCloneTransport(agent.DefaultTransport),
	}

	if forInitialize {
		agentOptions = append(agentOptions, agent.WithTimeout(o.InitializeRequestTimeout))
	} else {
		agentOptions = append(agentOptions, agent.WithTimeout(o.RequestTimeout))
	}

	return agent.NewAgent(agentOptions...)
}
