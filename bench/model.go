package bench

import (
	"sync"

	"github.com/isucon/isucandar/agent"
)

type User struct {
	mu sync.RWMutex

	Name        string
	DisplayName string
	Description string
	Password    string

	Agent *agent.Agent
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

