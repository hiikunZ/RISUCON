package bench

import (
	"context"
	"sync"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/worker"
)

type Scenario struct {
	mu sync.RWMutex

	Option Option
	// 競技者が使用した言語。ポータルへのレポーティングで使用される。
	Language string

}
