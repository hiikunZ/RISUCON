package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

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
type Answer struct {
	ID        int    `db:"id"`
	TaskID    int    `db:"task_id"`
	SubtaskID int    `db:"subtask_id"`
	Answer    string `db:"answer"`
	Score     int    `db:"score"`
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
		teamstandings.TeamName = team.Name
		teamstandings.TeamDisplayName = team.DisplayName
		teamstandings.TotalScore = 0

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
			teamstandings.Member2Name = member2.Name
			teamstandings.Member2DisplayName = member2.DisplayName
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
				if err := tx.GetContext(ctx, &leaderscore, "SELECT COALESCE(MAX(score),0) FROM answers WHERE subtask_id = ? AND EXISTS (SELECT * FROM submissions WHERE task_id = ? AND user_id = ? AND submissions.answer = answers.answer)", subtask.ID, task.ID, team.LeaderID); err != nil {
					return Standings{}, err
				}
				if subtaskscore < leaderscore {
					subtaskscore = leaderscore
				}

				if team.Member1ID != nil {
					member1score := 0
					if err := tx.GetContext(ctx, &member1score, "SELECT COALESCE(MAX(score),0) FROM answers WHERE subtask_id = ? AND EXISTS (SELECT * FROM submissions WHERE task_id = ? AND user_id = ? AND submissions.answer = answers.answer)", subtask.ID, task.ID, team.Member1ID); err != nil {
						return Standings{}, err
					}
					if subtaskscore < member1score {
						subtaskscore = member1score
					}
				}
				if team.Member2ID != nil {
					member2score := 0
					if err := tx.GetContext(ctx, &member2score, "SELECT COALESCE(MAX(score),0) FROM answers WHERE subtask_id = ? AND EXISTS (SELECT * FROM submissions WHERE task_id = ? AND user_id = ? AND submissions.answer = answers.answer)", subtask.ID, task.ID, team.Member2ID); err != nil {
						return Standings{}, err
					}
					if subtaskscore < member2score {
						subtaskscore = member2score
					}
				}
				taskscoringdata.Score += subtaskscore
			}
			scoringdata = append(scoringdata, taskscoringdata)
			teamstandings.TotalScore += taskscoringdata.Score
		}
		teamstandings.ScoringData = scoringdata
		standings.StandingsData = append(standings.StandingsData, teamstandings)
	}

	// sort
	for i := 0; i < len(standings.StandingsData); i++ {
		for j := i + 1; j < len(standings.StandingsData); j++ {
			if standings.StandingsData[i].TotalScore < standings.StandingsData[j].TotalScore {
				tmp := standings.StandingsData[i]
				standings.StandingsData[i] = standings.StandingsData[j]
				standings.StandingsData[j] = tmp
			}
		}
	}
	for i := 0; i < len(standings.StandingsData); i++ {
		standings.StandingsData[i].Rank = i + 1
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

type SubtaskDetail struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Statement   string `json:"statement"`
	Score       int    `json:"score"`
}
type TaskDetail struct {
	Name        string          `json:"name"`
	DisplayName string          `json:"display_name"`
	Statement   string          `json:"statement"`
	Score       int             `json:"score"`
	Subtasks    []SubtaskDetail `json:"subtasks"`
}

// GET /api/tasks/:taskname
func getTaskHandler(c echo.Context) error {
	taskname := c.Param("taskname")

	tx, err := dbConn.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	task := Task{}

	err = tx.GetContext(c.Request().Context(), &task, "SELECT * FROM tasks WHERE name = ?", taskname)

	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusBadRequest, "task not found")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get task: "+err.Error())
	}

	subtasks := []Subtask{}
	if err := tx.SelectContext(c.Request().Context(), &subtasks, "SELECT * FROM subtasks WHERE task_id = ?", task.ID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get subtasks: "+err.Error())
	}

	res := TaskDetail{
		Name:        task.Name,
		DisplayName: task.DisplayName,
		Statement:   task.Statement,
		Score:       0,
	}

	for _, subtask := range subtasks {
		subtaskdetail := SubtaskDetail{
			Name:        subtask.Name,
			DisplayName: subtask.DisplayName,
			Statement:   subtask.Statement,
		}
		if err := tx.GetContext(c.Request().Context(), &subtaskdetail.Score, "SELECT MAX(score) FROM answers WHERE subtask_id = ?", subtask.ID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get subtask score: "+err.Error())
		}
		res.Subtasks = append(res.Subtasks, subtaskdetail)
		res.Score += subtaskdetail.Score
	}
	return c.JSON(http.StatusOK, res)
}

type SubmitRequest struct {
	TaskName  string `json:"task_name"`
	Answer    string `json:"answer"`
	Timestamp int64  `json:"timestamp"`
}

type SubmitResponse struct {
	IsScored           bool   `json:"is_scored"`
	Score              int    `json:"score"`
	SubtaskName        string `json:"subtask_name"`
	SubTaskDisplayName string `json:"subtask_display_name"`
}

// POST /api/submit
func submitHandler(c echo.Context) error {
	ctx := c.Request().Context()
	defer c.Request().Body.Close()

	if err := verifyUserSession(c); err != nil {
		return err
	}

	username := c.Get("username").(string)

	tx, err := dbConn.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	user := User{}
	if err := tx.GetContext(c.Request().Context(), &user, "SELECT * FROM users WHERE name = ?", username); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user: "+err.Error())
	}

	team := Team{}
	err = tx.GetContext(c.Request().Context(), &team, "SELECT * FROM teams WHERE leader_id = ? OR member1_id = ? OR member2_id = ?", username, username, username)
	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusBadRequest, "you have not joined team")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get team: "+err.Error())
	}

	req := SubmitRequest{}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to bind request: "+err.Error())
	}

	task := Task{}
	err = tx.GetContext(c.Request().Context(), &task, "SELECT * FROM tasks WHERE name = ?", req.TaskName)
	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusBadRequest, "task not found")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get task: "+err.Error())
	}

	timestamp := time.Unix(req.Timestamp, 0)

	if _, err = tx.ExecContext(ctx, "INSERT INTO submissions (task_id, user_id, submitted_at, answer) VALUES (?, ?, ?, ?)", task.ID, user.ID, timestamp, req.Answer); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to insert submission: "+err.Error())
	}

	res := SubmitResponse{}

	answer := Answer{}
	err = tx.GetContext(c.Request().Context(), &answer, "SELECT * FROM answers WHERE task_id = ? AND answer = ?", task.ID, req.Answer)
	if err == sql.ErrNoRows {
		res.IsScored = false
		res.Score = 0
		res.SubtaskName = ""
		res.SubTaskDisplayName = ""
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get answer: "+err.Error())
	} else {
		res.IsScored = true
		res.Score = answer.Score
		subtask := Subtask{}
		if err := tx.GetContext(c.Request().Context(), &subtask, "SELECT * FROM subtasks WHERE id = ?", answer.SubtaskID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get subtask: "+err.Error())
		}
		res.SubtaskName = subtask.Name
		res.SubTaskDisplayName = subtask.DisplayName
	}

	return c.JSON(http.StatusOK, res)
}

// GET /api/submissions
func getSubmissionsHandler(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}
