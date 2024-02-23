package main

import (
	"fmt"
	"net/http"

	"github.com/isucon/isucandar/failure"
)

func validateLoginUser(loginRes *LoginResponse, user *User, team *Team) ResponseValidator {
	return func(r *http.Response) error {

		// HTTP status codeのチェックで失敗したら続行しない
		if err := IsuAssertStatus(200, r.StatusCode, Hint(PostLogin, "")); err != nil {
			return err
		}

		if err := parseJsonBody(r, loginRes); err != nil {
			AdminLogger.Printf("loginRes:%#v\n", loginRes)
			return failure.NewError(ValidationErrInvalidResponseBody, fmt.Errorf("%s: %v ", PostLogin+" のBodyの Json decodeに失敗しました", err))
		}

		// レスポンスの整合性チェック
		// Name
		if err := IsuAssert(user.Name, loginRes.Name, Hint(PostLogin, "Name")); err != nil {
			return err
		}
		// DisplayName
		if err := IsuAssert(user.DisplayName, loginRes.DisplayName, Hint(PostLogin, "DisplayName")); err != nil {
			return err
		}
		// TeamName
		if team != nil {
			if err := IsuAssert(team.Name, loginRes.TeamName, Hint(PostLogin, "TeamName")); err != nil {
				return err
			}
			if err := IsuAssert(team.DisplayName, loginRes.TeamDisplayName, Hint(PostLogin, "TeamDisplayName")); err != nil {
				return err
			}
		} else {
			if err := IsuAssert("", loginRes.TeamName, Hint(PostLogin, "TeamName")); err != nil {
				return err
			}
			if err := IsuAssert("", loginRes.TeamDisplayName, Hint(PostLogin, "TeamDisplayName")); err != nil {
				return err
			}
		}

		return nil
	}
}

func validategetUser(getuserRes *UserResponse, user *User, team *Team) ResponseValidator {
	return func(r *http.Response) error {

		// HTTP status codeのチェックで失敗したら続行しない
		if err := IsuAssertStatus(200, r.StatusCode, Hint(GetUser, "")); err != nil {
			return err
		}

		if err := parseJsonBody(r, getuserRes); err != nil {
			AdminLogger.Printf("getuserRes:%#v\n", getuserRes)
			return failure.NewError(ValidationErrInvalidResponseBody, fmt.Errorf("%s: %v ", GetUser+" のBodyの Json decodeに失敗しました", err))
		}

		// レスポンスの整合性チェック
		// Name
		if err := IsuAssert(user.Name, getuserRes.Name, Hint(GetUser, "Name")); err != nil {
			return err
		}
		// DisplayName
		if err := IsuAssert(user.DisplayName, getuserRes.DisplayName, Hint(GetUser, "DisplayName")); err != nil {
			return err
		}
		// Description
		if err := IsuAssert(user.Description, getuserRes.Description, Hint(GetUser, "Description")); err != nil {
			return err
		}
		// SubmissionCount
		if err := IsuAssert(len(user.SubmissionIDs), getuserRes.SubmissionCount, Hint(GetUser, "SubmissionCount")); err != nil {
			return err
		}

		if team != nil {
			// TeamName
			if err := IsuAssert(team.Name, getuserRes.TeamName, Hint(GetUser, "TeamName")); err != nil {
				return err
			}
			// TeamDisplayName
			if err := IsuAssert(team.DisplayName, getuserRes.TeamDisplayName, Hint(GetUser, "TeamDisplayName")); err != nil {
				return err
			}
		} else {
			// TeamName
			if err := IsuAssert("", getuserRes.TeamName, Hint(GetUser, "TeamName")); err != nil {
				return err
			}
			// TeamDisplayName
			if err := IsuAssert("", getuserRes.TeamDisplayName, Hint(GetUser, "TeamDisplayName")); err != nil {
				return err
			}
		}

		return nil
	}
}
