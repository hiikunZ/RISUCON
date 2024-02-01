package bench

import (
	"sync"
	"time"

	"github.com/isucon/isucandar/agent"
)

// benchmarker 内部で使用するモデルを集約するファイル

type User struct {
	mu sync.RWMutex

	Name        string
	DisplayName string
	Description string
	Password    string

	Agent *agent.Agent
}

type Team struct {
	Name               string
	DisplayName        string
	LeaderName         string
	LeaderDisplayName  string
	Member1Name        string
	Member1DisplayName string
	Member2Name        string
	Member2DisplayName string
	Description        string
	InvitationCode     string
}

type Task struct {
	Name            string
	DisplayName     string
	Statement       string
	SubmissionLimit int
	Subtasks        []Subtask
}

type Subtask struct {
	Name        string
	DisplayName string
	Statement   string
	Answers     []Answer
}

type Answer struct {
	Answer string
	Score  int
}

type Submission struct {
	UserName        string
	UserDisplayName string
	TaskName        string
	TeamDisplayname string
	Subtaskname     string
	SubmittedAt     time.Time
	Answer          string
}

func (u *User) GetAgent(o Option) (*agent.Agent, error) {
	u.mu.RLock()
	agent := u.Agent
	u.mu.RUnlock()

	if agent != nil {
		return agent, nil
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	agent, err := o.NewAgent(false)
	if err != nil {
		return nil, err
	}

	u.Agent = agent

	return agent, nil
}

func (u *User) ClearAgent() {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.Agent = nil
}
