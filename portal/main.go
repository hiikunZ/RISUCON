package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	echolog "github.com/labstack/gommon/log"

	_ "github.com/mattn/go-sqlite3"
)

const (
	listenPort                = 8080
	frontendContentsPath      = "public"
	DBFilepath                = "portal.db"
	sqliteDriverName          = "sqlite3"
	defaultSessionIDKey       = "SESSIONID"
	defaultSessionTeamNameKey = "teamname"
	benchmarklimit            = 4
)
const initsql = `CREATE TABLE IF NOT EXISTS teams(
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    passhash TEXT NOT NULL,
	server_ip TEXT NOT NULL,
	is_benchmarking BOOLEAN DEFAULT FALSE
);
CREATE TABLE IF NOT EXISTS score_data(
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    team_id INTEGER NOT NULL,
	team_name TEXT,
	is_passed BOOLEAN,
    score INTEGER,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    contestant_log TEXT,
    admin_log TEXT
);`

var (
	db     *sqlx.DB
	secret = []byte("risuconportal_session_cookiestore_defaultsecret")
)

func getDB() (*sqlx.DB, error) {
	db, err := sqlx.Connect(sqliteDriverName, DBFilepath)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	dbconn, err := getDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	db = dbconn
	db.SetMaxOpenConns(10)
	defer db.Close()

	_, err = db.Exec(initsql)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	e := echo.New()
	e.Debug = false
	e.Logger.SetLevel(echolog.ERROR)
	e.Use(middleware.Logger())
	cookiestore := sessions.NewCookieStore(secret)
	e.Use(session.Middleware(cookiestore))

	e.POST("/api/login", loginHandler)
	e.POST("/api/logout", logoutHandler)
	e.POST("/api/benchmark", benchmarkHandler)

	e.GET("/api/history", historyHandler)
	e.GET("/api/isbenchmarking", isBenchmarkingHandler)
	e.GET("/api/scoreboard", scoreboardHandler)

	e.Static("/assets", frontendContentsPath+"/assets")
	e.GET("/*", getIndexHandler)

	listenAddr := net.JoinHostPort("", strconv.Itoa(listenPort))
	if err := e.Start(listenAddr); err != nil {
		e.Logger.Errorf("failed to start HTTP server: %v", err)
		os.Exit(1)
	}
}

func getIndexHandler(c echo.Context) error {
	return c.File(frontendContentsPath + "/index.html")
}

type Team struct {
	ID             int    `db:"id" json:"-"`
	Name           string `db:"name" json:"name"`
	DisplayName    string `db:"display_name" json:"display_name"`
	Passhash       string `db:"passhash" json:"-"`
	ServerIP       string `db:"server_ip" json:"-"`
	IsBenchmarking bool   `db:"is_benchmarking" json:"-"`
}

type ScoreData struct {
	ID            int       `db:"id" json:"-"`
	TeamID        int       `db:"team_id" json:"-"`
	TeamName      string    `db:"team_name" json:"team_name"`
	IsPassed      bool      `db:"is_passed" json:"is_passed"`
	Score         int       `db:"score" json:"score"`
	Timestamp     time.Time `db:"timestamp" json:"timestamp"`
	ContestantLog string    `db:"contestant_log" json:"contestant_log"`
	AdminLog      string    `db:"admin_log" json:"-"`
}

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"` // ハッシュ化されていない
}

type LoginResponse struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

func loginHandler(c echo.Context) error {
	ctx := c.Request().Context()
	defer c.Request().Body.Close()

	req := LoginRequest{}

	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	t := Team{}

	err := db.GetContext(ctx, &t, "SELECT * FROM teams WHERE name = ?", req.Name)

	if err == sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user: "+err.Error())
	}

	pashhash := fmt.Sprintf("%x", sha256.Sum256([]byte(req.Password)))

	if t.Passhash != pashhash {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication failed")
	}

	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session: "+err.Error())
	}
	sess.Options = &sessions.Options{
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
	sess.Values[defaultSessionTeamNameKey] = t.Name
	if err = sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save session: "+err.Error())
	}

	return c.JSON(http.StatusOK, LoginResponse{
		Name:        t.Name,
		DisplayName: t.DisplayName,
	})
}

func verifyUserSession(c echo.Context) error {
	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session")
	}
	if sess.Values[defaultSessionTeamNameKey] == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not logged in")
	}
	return nil
}

func benchmarkHandler(c echo.Context) error {
	if err := verifyUserSession(c); err != nil {
		return err
	}

	sess, _ := session.Get(defaultSessionIDKey, c)
	teamname, _ := sess.Values[defaultSessionTeamNameKey].(string)

	tx := db.MustBegin()
	defer tx.Rollback()

	cnt := 0

	err := tx.Get(&cnt, "SELECT COUNT(*) FROM teams WHERE is_benchmarking = TRUE")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get benchmarking count: "+err.Error())
	}

	if cnt >= benchmarklimit {
		return echo.NewHTTPError(http.StatusForbidden, "キューがいっぱいです。少し待ってから再度お試しください。")
	}

	t := Team{}
	err = tx.Get(&t, "SELECT * FROM teams WHERE name = ?", teamname)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get team: "+err.Error())
	}

	if t.IsBenchmarking {
		return echo.NewHTTPError(http.StatusForbidden, "既にベンチマーク中です。")
	}

	_, err = tx.Exec("UPDATE teams SET is_benchmarking = TRUE WHERE name = ?", teamname)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update team: "+err.Error())
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit transaction: "+err.Error())
	}

	c.JSON(http.StatusOK, nil)

	go func() {
		// ベンチマーク処理
		cmd := exec.Command("../bench/benchmarker", "--target-host="+t.ServerIP, "--stage=prod")

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()

		tx := db.MustBegin()
		defer tx.Rollback()

		if err != nil {
			log.Printf("failed to run benchmarker: %v", err)
		} else {
			contestantlog := stdout.String()
			adminlog := stderr.String()

			adminloglines := bytes.Split([]byte(adminlog), []byte("\n"))
			lastline := adminloglines[len(adminloglines)-2]

			passed := false
			score := 0
			// [ADMIN] xx:xx:xx [PASSED]: %v,[SCORE]: %d
			fmt.Fscan(bytes.NewReader(lastline), "[ADMIN] %*s [PASSED]: %t,[SCORE]: %d", &passed, &score)

			_, err = tx.Exec("INSERT INTO score_data(team_id, team_name, is_passed, score, contestant_log, admin_log) VALUES(?, ?, ?, ?, ?, ?)", t.ID, t.Name, passed, score, contestantlog, adminlog)
			if err != nil {
				log.Printf("failed to insert score data: %v", err)
			}
		}

		// ベンチマーク完了後

		_, err = tx.Exec("UPDATE teams SET is_benchmarking = FALSE WHERE name = ?", teamname)
		if err != nil {
			log.Printf("failed to update team: %v", err)
		}

		if err := tx.Commit(); err != nil {
			log.Printf("failed to commit transaction: %v", err)
		}
	}()

	return nil
}

func historyHandler(c echo.Context) error {
	if err := verifyUserSession(c); err != nil {
		return err
	}

	sess, _ := session.Get(defaultSessionIDKey, c)
	teamname, _ := sess.Values[defaultSessionTeamNameKey].(string)

	t := Team{}
	err := db.Get(&t, "SELECT * FROM teams WHERE name = ?", teamname)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get team: "+err.Error())
	}

	scores := []ScoreData{}
	err = db.Select(&scores, "SELECT * FROM score_data WHERE team_id = ?", t.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get score data: "+err.Error())
	}

	return c.JSON(http.StatusOK, scores)
}

func isBenchmarkingHandler(c echo.Context) error {
	if err := verifyUserSession(c); err != nil {
		return err
	}

	sess, _ := session.Get(defaultSessionIDKey, c)
	teamname, _ := sess.Values[defaultSessionTeamNameKey].(string)

	t := Team{}
	err := db.Get(&t, "SELECT * FROM teams WHERE name = ?", teamname)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get team: "+err.Error())
	}

	return c.JSON(http.StatusOK, t.IsBenchmarking)
}

type ScoreboardData struct {
	TeamName  string    `db:"team_name" json:"team_name"`
	Score     int       `db:"score" json:"score"`
	Timestamp time.Time `db:"timestamp" json:"timestamp"`
}

func scoreboardHandler(c echo.Context) error {
	if err := verifyUserSession(c); err != nil {
		return err
	}

	scores := []ScoreData{}
	err := db.Select(&scores, "SELECT * FROM score_data")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get score data: "+err.Error())
	}

	res := []ScoreboardData{}
	for _, score := range scores {
		res = append(res, ScoreboardData{
			TeamName:  score.TeamName,
			Score:     score.Score,
			Timestamp: score.Timestamp,
		})
	}

	return c.JSON(http.StatusOK, res)
}

func logoutHandler(c echo.Context) error {
	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get session: "+err.Error())
	}
	sess.Options = &sessions.Options{
		MaxAge:   -1,
		HttpOnly: true,
	}
	sess.Values[defaultSessionTeamNameKey] = ""
	if err = sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save session: "+err.Error())
	}

	return c.NoContent(http.StatusOK)
}
