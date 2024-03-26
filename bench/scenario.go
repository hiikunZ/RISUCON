package main

import (
	"context"
	"sync"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/failure"
	"github.com/isucon/isucandar/worker"
)

type Scenario struct {
	mu sync.RWMutex

	Option Option
	// 競技者が使用した言語。ポータルへのレポーティングで使用される。
	Language string

	Prepared_Tasks Set[*Task]

	Tasks       Set[*Task]
	Users       Set[*User]
	Teams       Set[*Team]
	Submissions Set[*Submission]

	ConsumedUserIDs *LightSet

	ScenarioControlWg sync.WaitGroup

	LoginSuccessCountMu sync.Mutex
	LoginSuccessCount   int

	UserRegistrationMu    sync.Mutex
	UserRegistrationCount int

	VisitorMu    sync.Mutex
	VisitorCount int

	AddTask bool

	NowClock time.Time
}

// 初期化処理を行うが、初期化処理を正しく実行しているかをチェックする。
// 初期化処理自体は `main.DefaultInitializeRequestTimeout` 秒以内に終了する必要がある。
func (s *Scenario) Prepare(ctx context.Context, step *isucandar.BenchmarkStep) error {
	// 初期データの読み込み
	err := s.LoadInitialData()
	if err != nil {
		return err
	}

	s.addexistnametoset()

	ctx, cancel := context.WithTimeout(ctx, s.Option.InitializeRequestTimeout)
	defer cancel()

	report := TimeReporter("初期化処理", s.Option)
	defer report()

	agent, err := s.Option.NewAgent(true)
	if err != nil {
		return failure.NewError(ErrCannotCreateNewAgent, err)
	}

	err = s.DoInitialize(ctx, step, agent)
	if err != nil {
		return err
	}
	ContestantLogger.Println("初期化処理が成功しました！")

	// Pretest を1回まわす
	if err := s.PretestScenario(ctx, step); err != nil {
		ContestantLogger.Println("整合性チェックに失敗しました")
		return err
	}

	ContestantLogger.Println("整合性チェックに成功しました！")

	return nil
}

// 主なシナリオとしては次の通り
// 1. 参加者シナリオ
// 2. 観戦者シナリオ
// 3. ユーザー/チーム登録シナリオ
// 4. 運営シナリオ
func (s *Scenario) Load(ctx context.Context, step *isucandar.BenchmarkStep) error {
	if s.Option.PrepareOnly {
		return nil
	}
	ContestantLogger.Println("アプリケーションへの負荷走行を開始します")

	wg := &sync.WaitGroup{}
	// 各シナリオを走らせる。
	loginSuccess, err := s.NewLoginScenarioWorker(step, 3)
	if err != nil {
		return err
	}

	userRegistration, err := s.NewUserRegistrationScenarioWorker(step, 1)
	if err != nil {
		return err
	}

	visitor, err := s.NewVisitorScenarioWorker(step, 1)
	if err != nil {
		return err
	}

	addTask, err := s.FireAddTask(step)
	if err != nil {
		return err
	}

	workers := []*worker.Worker{loginSuccess, userRegistration, visitor, addTask}

	for _, w := range workers {
		wg.Add(1)
		worker := w
		go func() {
			defer wg.Done()
			worker.Process(ctx)
		}()
	}

	// ベンチマーカー走行中の負荷調整を10秒ごとにかける
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.loadAdjustor(ctx, step, loginSuccess, userRegistration, visitor)
	}()

	wg.Wait()
	s.ScenarioControlWg.Wait()

	ContestantLogger.Println("負荷走行がすべて終了しました")
	AdminLogger.Println("負荷走行がすべて終了しました")

	return nil
}
func (s *Scenario) Validation(ctx context.Context, step *isucandar.BenchmarkStep) error {
	if s.Option.PrepareOnly {
		return nil
	}

	return nil
}
