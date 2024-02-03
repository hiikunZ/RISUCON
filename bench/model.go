package bench

import (
	"sync"
	"time"

	"github.com/isucon/isucandar/agent"
)

const (
	nulluserid = -1
)

// benchmarker 内部で使用するモデルを集約するファイル
type User struct {
	mu            sync.RWMutex
	ID            int    `json:"id"`
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`
	Description   string `json:"description"`
	Password      string `json:"password"` // Passhash は dump するときに計算する
	SubmissionIDs []int  `json:"submission_ids"`
	Agent         *agent.Agent
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
