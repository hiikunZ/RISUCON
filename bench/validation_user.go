package main

import (
	"fmt"
	"net/http"

	"github.com/isucon/isucandar/failure"
)

func validateLoginUser(loginRes *LoginResponse, user *User) ResponseValidator {
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

		return nil
	}
}
