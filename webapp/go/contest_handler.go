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
type Subtask struct {
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
		subtasks := []Subtask{}
		if err := tx.SelectContext(ctx, &subtasks, "SELECT * FROM subtasks WHERE task_id = ?", task.ID); err != nil {
			return []TaskAbstract{}, err
		}
		for _, subtask := range subtasks {
			maxscore_for_subtask := 0
			if err := tx.GetContext(ctx, &maxscore_for_subtask, "SELECT MAX(score) FROM answers WHERE subtask_id = ?", subtask.ID); err != nil {
				return []TaskAbstract{}, err
			}
			maxscore += maxscore_for_subtask
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

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit transaction: "+err.Error())
	}

	return c.JSON(http.StatusOK, taskabstarcts)
}

type TeamsStandingsSub struct {
	TaskName     string `json:"task_name"`
	HasSubmitted bool   `json:"has_submitted"`
	Score        int    `json:"score"`
}
type TeamsStandings struct {
	Rank               int                 `json:"rank"`
	TeamName           string              `json:"team_name"`
	TeamDisplayName    string              `json:"team_display_name"`
	LeaderName         string              `json:"leader_name"`
	LeaderDisplayName  string              `json:"leader_display_name"`
	Member1Name        string              `json:"member1_name"`
	Member1DisplayName string              `json:"member1_display_name"`
	Member2Name        string              `json:"member2_name"`
	Member2DisplayName string              `json:"member2_display_name"`
	ScoringData        []TeamsStandingsSub `json:"scoring_data"`
	TotalScore         int                 `json:"total_score"`
}
type Standings struct {
	TasksData     []TaskAbstract   `json:"tasks_data"`
	StandingsData []TeamsStandings `json:"standings_data"`
}

func getstandings(ctx context.Context, tx *sqlx.Tx) (Standings, error) {
	standings := Standings{}

	tasks := []Task{}
	if err := tx.SelectContext(ctx, &tasks, "SELECT * FROM tasks ORDER BY name"); err != nil {
		return Standings{}, err
	}
	for _, task := range tasks {
		standings.TasksData = append(standings.TasksData, TaskAbstract{
			Name:        task.Name,
			DisplayName: task.DisplayName,
		})
	}

	teams := []Team{}
	if err := tx.SelectContext(ctx, &teams, "SELECT * FROM teams ORDER BY name"); err != nil {
		return Standings{}, err
	}
	for _, team := range teams {
		teamstandings := TeamsStandings{}
		leader := User{}
		if err := tx.GetContext(ctx, &leader, "SELECT * FROM users WHERE id = ?", team.LeaderID); err != nil {
			return Standings{}, err
		}
		teamstandings.LeaderName = leader.Name
		teamstandings.LeaderDisplayName = leader.DisplayName
		if team.Member1ID != nil {
			member1 := User{}
			if err := tx.GetContext(ctx, &member1, "SELECT * FROM users WHERE id = ?", team.Member1ID); err != nil {
				return Standings{}, err
			}
			teamstandings.Member1Name = member1.Name
			teamstandings.Member1DisplayName = member1.DisplayName
		}
		if team.Member2ID != nil {
			member2 := User{}
			if err := tx.GetContext(ctx, &member2, "SELECT * FROM users WHERE id = ?", team.Member1ID); err != nil {
				return Standings{}, err
			}
			teamstandings.Member1Name = member2.Name
			teamstandings.Member1DisplayName = member2.DisplayName
		}

		scoringdata := []TeamsStandingsSub{}
		for _, task := range tasks {
			taskscoringdata := TeamsStandingsSub{}
			taskscoringdata.TaskName = task.Name
			taskscoringdata.HasSubmitted = false
			taskscoringdata.Score = 0

			subtasks := []Subtask{}
			if err := tx.SelectContext(ctx, &subtasks, "SELECT * FROM subtasks WHERE task_id = ?", task.ID); err != nil {
				return Standings{}, err
			}
			submissioncount := 0
			if err := tx.GetContext(ctx, &submissioncount, "SELECT COUNT(*) FROM submissions WHERE task_id = ? AND user_id = ?", task.ID, team.LeaderID); err != nil {
				return Standings{}, err
			}
			if submissioncount > 0 {
				taskscoringdata.HasSubmitted = true
			}
			if team.Member1ID != nil {
				if err := tx.GetContext(ctx, &submissioncount, "SELECT COUNT(*) FROM submissions WHERE task_id = ? AND user_id = ?", task.ID, team.Member1ID); err != nil {
					return Standings{}, err
				}
				if submissioncount > 0 {
					taskscoringdata.HasSubmitted = true
				}
			}
			if team.Member2ID != nil {
				if err := tx.GetContext(ctx, &submissioncount, "SELECT COUNT(*) FROM submissions WHERE task_id = ? AND user_id = ?", task.ID, team.Member2ID); err != nil {
					return Standings{}, err
				}
				if submissioncount > 0 {
					taskscoringdata.HasSubmitted = true
				}
			}
			
			for _, subtask := range subtasks {
				subtaskscore := 0

				leaderscore := 0
				if err := tx.GetContext(ctx, &leaderscore, "SELECT MAX(score) FROM answers WHERE subtask_id = ? AND EXISTS (SELECT * FROM submissions WHERE task_id = ? AND user_id = ? AND submissions.answer = answers.answer)", subtask.ID, task.ID, team.LeaderID); err != nil {
					return Standings{}, err
				}
				subtaskscore = max(subtaskscore, leaderscore)

				if team.Member1ID != nil {
					member1score := 0
					if err := tx.GetContext(ctx, &member1score, "SELECT MAX(score) FROM answers WHERE subtask_id = ? AND EXISTS (SELECT * FROM submissions WHERE task_id = ? AND user_id = ? AND submissions.answer = answers.answer)", subtask.ID, task.ID, team.Member1ID); err != nil {
						return Standings{}, err
					}
					subtaskscore = max(subtaskscore, member1score)
				}
				if team.Member2ID != nil {
					member2score := 0
					if err := tx.GetContext(ctx, &member2score, "SELECT MAX(score) FROM answers WHERE subtask_id = ? AND EXISTS (SELECT * FROM submissions WHERE task_id = ? AND user_id = ? AND submissions.answer = answers.answer)", subtask.ID, task.ID, team.Member2ID); err != nil {
						return Standings{}, err
					}
					subtaskscore = max(subtaskscore, member2score)
				}
				taskscoringdata.Score += subtaskscore
			}
			scoringdata = append(scoringdata, taskscoringdata)
		}
		teamstandings.ScoringData = scoringdata
		standings.StandingsData = append(standings.StandingsData, teamstandings)
	}
	return standings, nil
}

// GET /api/stanings
func getStandingsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	standings, err := getstandings(ctx, tx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get standings: "+err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit transaction: "+err.Error())
	}

	return c.JSON(http.StatusOK, standings)
}
