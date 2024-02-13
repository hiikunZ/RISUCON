package bench

import (
	"sync"
)

type Scenario struct {
	mu sync.RWMutex

	Option Option
	// 競技者が使用した言語。ポータルへのレポーティングで使用される。
	Language string

	Tasks Set[*Task]
	Users Set[*User]
	Teams Set[*Team]

	ScenarioControlWg  sync.WaitGroup
	SubmitCountMu      sync.Mutex
	SubmitSuccessCount int

	UserRegistrationMu    sync.Mutex
	UserRegistrationCount int
}
