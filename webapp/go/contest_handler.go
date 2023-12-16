package main

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type Task struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	DisplayName string `db:"display_name"`
	Statement   string `db:"statement"`
}
type Question struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	DisplayName string `db:"display_name"`
	TaskID      int    `db:"task_id"`
	Statement   string `db:"statement"`
}

type TaskAbstract struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	MaxScore    int    `json:"max_score"`
}

func gettaskabstarcts(ctx context.Context, tx *sqlx.Tx) ([]TaskAbstract, error) {
	tasks := []Task{}
	if err := tx.SelectContext(ctx, &tasks, "SELECT * FROM tasks ORDER BY name"); err != nil {
		return []TaskAbstract{}, err
	}
	res := []TaskAbstract{}
	for _, task := range tasks {
		maxscore := 0
		questions := []Question{}
		if err := tx.SelectContext(ctx, &questions, "SELECT * FROM questions WHERE task_id = ?", task.ID); err != nil {
			return []TaskAbstract{}, err
		}
		for _, question := range questions {
			maxscore_for_question := 0
			if err := tx.GetContext(ctx, &maxscore_for_question, "SELECT MAX(score) FROM answers WHERE question_id = ?", question.ID); err != nil {
				return []TaskAbstract{}, err
			}
			maxscore += maxscore_for_question
		}
		res = append(res, TaskAbstract{
			Name:        task.Name,
			DisplayName: task.DisplayName,
			MaxScore:    maxscore,
		})
	}

	return res, nil
}

// GET /api/tasks
func getTasksHandler(c echo.Context) error {
	ctx := c.Request().Context()

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	taskabstarcts, err := gettaskabstarcts(ctx, tx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get taskabstarcts: "+err.Error())
	}

	return c.JSON(http.StatusOK, taskabstarcts)
}


type StandingsData struct {
	TeamName string `json:"team_name"`
	TeamDisplayName string `json:"team_display_name"`
	Team
}
// GET /api/stanings
func getStandingsHandler(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}
