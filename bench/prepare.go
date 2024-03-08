package main

import (
	"context"
	"crypto/sha256"
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
	if err := s.Prepared_Tasks.LoadJSON("./data/tasks.json"); err != nil {
		ContestantLogger.Println("初期データ (tasks) のロードに失敗しました")
		return failure.NewError(ErrFailedToLoadJson, err)
	}

	for i := 1; i <= 2; i++ {
		task, _ := s.Prepared_Tasks.Get(i)
		s.Tasks.Add(task)
	}

	if err := s.Users.LoadJSON("./data/users.json"); err != nil {
		ContestantLogger.Println("初期データ (users) のロードに失敗しました")
		return failure.NewError(ErrFailedToLoadJson, err)
	}
	if err := s.Teams.LoadJSON("./data/teams.json"); err != nil {
		ContestantLogger.Println("初期データ (teams) のロードに失敗しました")
		return failure.NewError(ErrFailedToLoadJson, err)
	}
	if err := s.Submissions.LoadJSON("./data/submissions.json"); err != nil {
		ContestantLogger.Println("初期データ (submissions) のロードに失敗しました")
		return failure.NewError(ErrFailedToLoadJson, err)
	}

	// nonce の設定
	nonce = []byte("RiSuCoN")

	// ファイルのハッシュ値を計算
	indexfile, err := os.ReadFile("./data/index.html")
	if err != nil {
		ContestantLogger.Println("初期データ (index.html) の読み込みに失敗しました")
		return failure.NewError(ErrFailedToLoadJson, err)
	}
	indexfile = append(indexfile, nonce...)
	indexhash = sha256.Sum256(indexfile)

	jsfile, err := os.ReadFile("./data/" + jsfilename)
	if err != nil {
		ContestantLogger.Println("初期データ (js) の読み込みに失敗しました")
		return failure.NewError(ErrFailedToLoadJson, err)
	}
	jsfile = append(jsfile, nonce...)
	jsfilehash = sha256.Sum256(jsfile)

	cssfile, err := os.ReadFile("./data/" + cssfilename)
	if err != nil {
		ContestantLogger.Println("初期データ (css) の読み込みに失敗しました")
		return failure.NewError(ErrFailedToLoadJson, err)
	}
	cssfile = append(cssfile, nonce...)
	cssfilehash = sha256.Sum256(cssfile)

	return nil
}

// webapp の POST /initialize を叩く。
func (s *Scenario) DoInitialize(ctx context.Context, step *isucandar.BenchmarkStep, agent *agent.Agent) error {
	res, err := PostInitializeAction(ctx, agent)
	if err != nil {
		ContestantLogger.Printf("初期化リクエストに失敗しました")
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
