package main

import (
	"database/sql"
	"net/http"
	"os/exec"
	"strings"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type CreateTeamRequest struct {
	Name			string `json:"name"`
	DisplayName		string `json:"display_name"`
	leader_id		int	   
	Description		string `json:"description"`
	InvitationCode	string 
}

func generateInvitationCode() string {
	out, err := exec.Command("/bin/bash","-c","openssl rand -hex 8").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSuffix(string(out), "\n")
}

// POST /api/team/create
func createTeamHandler(c echo.Context) error {
	ctx := c.Request().Context()
	defer c.Request().Body.Close()

	req := CreateTeamRequest{}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to bind request: "+err.Error())
	}

	if req.Name == "" || req.DisplayName == "" || req.Description == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	req.InvitationCode = generateInvitationCode()

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session: "+err.Error())
	}
	username, ok := sess.Values[defaultSessionUserNameKey].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "not logged in")
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	usr := User{}
	err = tx.GetContext(ctx, &usr, "SELECT * FROM users WHERE name = ?", username)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user: "+err.Error())
	}
	req.leader_id = usr.ID

	team := Team{}
	err = tx.GetContext(ctx, &team, "SELECT * FROM teams WHERE name = ?", req.Name)
	if err == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "team name already exists")
	} else if err != sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get team: "+err.Error())
	}

	if _,err = tx.ExecContext(ctx, "INSERT INTO teams (name, display_name, leader_id, description, invitation_code) VALUES (?, ?, ?, ?, ?)", req.Name, req.DisplayName, req.leader_id, req.Description, req.InvitationCode); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to insert team: "+err.Error())
	}

	if err = tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit transaction: "+err.Error())
	}

	return c.NoContent(http.StatusOK)
}

type JoinTeamRequest struct {
	TeamName		string `json:"team_name"`
	InvitationCode	string `json:"invitation_code"`
}

// POST /api/team/join
func joinTeamHandler(c echo.Context) error {
	ctx := c.Request().Context()
	defer c.Request().Body.Close()

	req := JoinTeamRequest{}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to bind request: "+err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	team := Team{}
	err = tx.GetContext(ctx, &team, "SELECT * FROM teams WHERE name = ?", req.TeamName)
	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusBadRequest, "team not found")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get team: "+err.Error())
	}

	if team.InvitationCode != req.InvitationCode {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid invitation code")
	}

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session: "+err.Error())
	}
	username, ok := sess.Values[defaultSessionUserNameKey].(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "not logged in")
	}

	usr := User{}
	err = tx.GetContext(ctx, &usr, "SELECT * FROM users WHERE name = ?", username)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user: "+err.Error())
	}

	if team.Member1ID == nil {
		if _, err := tx.ExecContext(ctx, "UPDATE teams SET member1_id = ? WHERE name = ?", usr.ID, req.TeamName); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update team: "+err.Error())
		}
	} else if team.Member2ID == nil {
		if _, err := tx.ExecContext(ctx, "UPDATE teams SET member2_id = ? WHERE name = ?", usr.ID, req.TeamName); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update team: "+err.Error())
		}
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, "team is full")
	}

	if err = tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit transaction: "+err.Error())
	}

	return c.NoContent(http.StatusOK)
}

type TeamResponse struct {
	Name			string `json:"name"`
	DisplayName		string `json:"display_name"`
	Leader			string `json:"leader"`
	Member1			string `json:"member1"`
	Member2			string `json:"member2"`
	Description		string `json:"description"`
	InvitationCode	string `json:"invitation_code"`
}

// GET /api/team/:teamname
func getTeamHandler(c echo.Context) error {
	teamname := c.Param("teamname")

	tx, err := dbConn.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	team := Team{}

	err = tx.GetContext(c.Request().Context(), &team, "SELECT * FROM teams WHERE name = ?", teamname)
	
	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusBadRequest, "team not found")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get team: "+err.Error())
	}

	res := TeamResponse{
		Name:			team.Name,
		DisplayName:	team.DisplayName,
		Description:	team.Description,
	}

	leader := User{}
	err = tx.GetContext(c.Request().Context(), &leader, "SELECT * FROM users WHERE id = ?", team.LeaderID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get leader: "+err.Error())
	}
	res.Leader = leader.Name

	if team.Member1ID != nil {
		member1 := User{}
		err = tx.GetContext(c.Request().Context(), &member1, "SELECT * FROM users WHERE id = ?", team.Member1ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get member1: "+err.Error())
		}
		res.Member1 = member1.Name
	} else {
		res.Member1 = ""
	}

	if team.Member2ID != nil {
		member2 := User{}
		err = tx.GetContext(c.Request().Context(), &member2, "SELECT * FROM users WHERE id = ?", team.Member2ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get member2: "+err.Error())
		}
		res.Member2 = member2.Name
	} else{
		res.Member2 = ""
	}

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session: "+err.Error())
	}
	username, ok := sess.Values[defaultSessionUserNameKey].(string)
	if !ok || username != res.Leader {
		res.InvitationCode = ""
	} else {
		res.InvitationCode = team.InvitationCode
	}

	return c.JSON(http.StatusOK, res)
}