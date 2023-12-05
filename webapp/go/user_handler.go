package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const (
	defaultSessionIDKey       = "SESSIONID"
	defaultSessionUserNameKey = "username"
)

type User struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	DisplayName string `db:"display_name"`
	Description string `db:"description"`
	Passhash    string `db:"passhash"`
}

type Team struct {
	ID             int    `db:"id"`
	Name           string `db:"name"`
	DisplayName    string `db:"display_name"`
	LeaderID       int    `db:"leader_id"`
	Member1ID      int    `db:"member1_id"`
	Member2ID      int    `db:"member2_id"`
	Description    string `db:"description"`
	InvitationCode string `db:"invitation_code"`
}

type RegisterRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Password    string `json:"password"` // ハッシュ化されていない
}

// POST /api/register
func registerHandler(c echo.Context) error {
	ctx := c.Request().Context()
	defer c.Request().Body.Close()

	req := RegisterRequest{}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to bind request: "+err.Error())
	}

	if req.Name == "" || req.DisplayName == "" || req.Description == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	// 同じ name のユーザーがいないか確認
	usr := User{}

	err = tx.GetContext(ctx, &usr, "SELECT * FROM users WHERE name = ?", req.Name)

	if err == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "user already exists")
	} else if err != sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user: "+err.Error())
	}

	pashhash := fmt.Sprintf("%x", sha256.Sum256([]byte(req.Password)))

	if _, err = tx.ExecContext(ctx, "INSERT INTO users (name, display_name, description, passhash) VALUES (?, ?, ?, ?)", req.Name, req.DisplayName, req.Description, pashhash); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to insert user: "+err.Error())
	}

	if err = tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit transaction: "+err.Error())
	}

	return c.NoContent(http.StatusCreated)
}

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"` // ハッシュ化されていない
}

// POST /api/login
func loginHandler(c echo.Context) error {
	ctx := c.Request().Context()
	defer c.Request().Body.Close()

	req := LoginRequest{}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to bind request: "+err.Error())
	}

	tx, err := dbConn.BeginTxx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	usr := User{}

	err = tx.GetContext(ctx, &usr, "SELECT * FROM users WHERE name = ?", req.Name)

	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user: "+err.Error())
	}

	pashhash := fmt.Sprintf("%x", sha256.Sum256([]byte(req.Password)))

	if usr.Passhash != pashhash {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication failed")
	}

	if err = tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit transaction: "+err.Error())
	}

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session: "+err.Error())
	}
	sess.Options = &sessions.Options{
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
	sess.Values[defaultSessionUserNameKey] = usr.Name
	if err = sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save session: "+err.Error())
	}

	return c.NoContent(http.StatusOK)
}

// POST /api/logout
func logoutHandler(c echo.Context) error {
	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session: "+err.Error())
	}
	sess.Options = &sessions.Options{
		MaxAge:   -1,
		HttpOnly: true,
	}
	sess.Values[defaultSessionUserNameKey] = ""
	if err = sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save session: "+err.Error())
	}

	return c.NoContent(http.StatusOK)
}

type UserResponse struct {
	Name            string `json:"name"`
	DisplayName     string `json:"display_name"`
	Description     string `json:"description"`
	SubmissionCount int    `json:"submission_count"`
	Teamname        string `json:"teamname"`
	Teamdisplayname string `json:"teamdisplayname"`
}

// GET /api/user/:username
func getUserHandler(c echo.Context) error {
	username := c.Param("username")

	tx, err := dbConn.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to begin transaction: "+err.Error())
	}
	defer tx.Rollback()

	usr := User{}

	err = tx.GetContext(c.Request().Context(), &usr, "SELECT * FROM users WHERE name = ?", username)

	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user: "+err.Error())
	}

	res := UserResponse{
		Name:        usr.Name,
		DisplayName: usr.DisplayName,
		Description: usr.Description,
	}

	err = tx.GetContext(c.Request().Context(), &res.SubmissionCount, "SELECT COUNT(*) FROM submissions JOIN users ON user_id = users.id WHERE name = ?", username)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get submission count: "+err.Error())
	}

	team := Team{}

	err = tx.GetContext(c.Request().Context(), &team, "SELECT teams.* FROM teams JOIN users ON leader_id = users.id OR member1_id = users.id OR member2_id = users.id WHERE users.name = ?", username)
	if err == sql.ErrNoRows {
		// 空文字列を返せばフロントエンドがいい感じに処理してくれる
		res.Teamname = ""
		res.Teamdisplayname = ""
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get team info: "+err.Error())
	} else {
		res.Teamname = team.Name
		res.DisplayName = team.DisplayName
	}
	return c.JSON(http.StatusOK, res)
}
