package bench

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/failure"
)

func (s *Set[T]) LoadJSON(jsonFile string) error {
	file, err := os.Open(jsonFile)
	if err != nil {
		return err
	}
	defer file.Close()

	models := []T{}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&models); err != nil {
		return err
	}

	for _, model := range models {
		if !s.Add(model) {
			return fmt.Errorf("unexpected error on dump loading: %v", model)
		}
	}

	return nil
}
func (s *Scenario) LoadInitialData() error {
	if err := s.Tasks.LoadJSON("./data/teams.json"); err != nil {
		ContestantLogger.Println("初期データ (tasks) のロードに失敗しました")
		return failure.NewError(ErrFailedToLoadJson, err)
	}
	if err := s.Teams.LoadJSON("./data/users.json"); err != nil {
		ContestantLogger.Println("初期データ (users) のロードに失敗しました")
		return failure.NewError(ErrFailedToLoadJson, err)
	}
	if err := s.Users.LoadJSON("./data/teams.json"); err != nil {
		ContestantLogger.Println("初期データ (teams) のロードに失敗しました")
		return failure.NewError(ErrFailedToLoadJson, err)
	}
	return nil
}

// webapp の POST /initialize を叩く。
func (s *Scenario) DoInitialize(ctx context.Context, step *isucandar.BenchmarkStep, agent *agent.Agent) error {
	res, err := PostInitializeAction(ctx, agent)
	if err != nil {
		return failure.NewError(ErrPrepareInvalidRequest, err)
	}
	defer res.Body.Close()

	initializeResponse := &InitializeResponse{}

	validationError := ValidateResponse(
		res,
		WithInitializationSuccess(initializeResponse),
	)
	validationError.Add(step)

	// 後の統計用に使用言語を取得し、ロギングしておく。
	s.Language = initializeResponse.Language
	AdminLogger.Printf("[LANGUAGE] %s", initializeResponse.Language)

	if !validationError.IsEmpty() {
		ContestantLogger.Printf("初期化リクエストに失敗しました")
		return validationError
	} else {
		return nil
	}
}
