package main

import (
	"fmt"
	"net/http"

	"github.com/isucon/isucandar/failure"
)

func validategetTeam(getteamRes *TeamResponse, team *Team, leader, member1, member2 *User, isleader bool) ResponseValidator {
	return func(r *http.Response) error {

		// HTTP status codeのチェックで失敗したら続行しない
		if err := IsuAssertStatus(200, r.StatusCode, Hint(GetTeam, "")); err != nil {
			return err
		}

		if err := parseJsonBody(r, getteamRes); err != nil {
			AdminLogger.Printf("getteamRes:%#v\n", getteamRes)
			return failure.NewError(ValidationErrInvalidResponseBody, fmt.Errorf("%s: %v ", GetTeam+" のBodyの Json decodeに失敗しました", err))
		}

		// レスポンスの整合性チェック
		// Name
		if err := IsuAssert(team.Name, getteamRes.Name, Hint(GetTeam, "Name")); err != nil {
			return err
		}
		// DisplayName
		if err := IsuAssert(team.DisplayName, getteamRes.DisplayName, Hint(GetTeam, "DisplayName")); err != nil {
			return err
		}
		// Leader
		if err := IsuAssert(leader.Name, getteamRes.LeaderName, Hint(GetTeam, "Leader")); err != nil {
			return err
		}
		if err := IsuAssert(leader.DisplayName, getteamRes.LeaderDisplayName, Hint(GetTeam, "Leader")); err != nil {
			return err
		}
		// Member1
		if member1 != nil {
			if err := IsuAssert(member1.Name, getteamRes.Member1Name, Hint(GetTeam, "Member1")); err != nil {
				return err
			}
			if err := IsuAssert(member1.DisplayName, getteamRes.Member1DisplayName, Hint(GetTeam, "Member1")); err != nil {
				return err
			}
		} else {
			if err := IsuAssert("", getteamRes.Member1Name, Hint(GetTeam, "Member1")); err != nil {
				return err
			}
			if err := IsuAssert("", getteamRes.Member1DisplayName, Hint(GetTeam, "Member1")); err != nil {
				return err
			}
		}

		// Member2
		if member2 != nil {
			if err := IsuAssert(member2.Name, getteamRes.Member2Name, Hint(GetTeam, "Member2")); err != nil {
				return err
			}
			if err := IsuAssert(member2.DisplayName, getteamRes.Member2DisplayName, Hint(GetTeam, "Member2")); err != nil {
				return err
			}
		} else {
			if err := IsuAssert("", getteamRes.Member2Name, Hint(GetTeam, "Member2")); err != nil {
				return err
			}
			if err := IsuAssert("", getteamRes.Member2DisplayName, Hint(GetTeam, "Member2")); err != nil {
				return err
			}
		}

		// Description
		if err := IsuAssert(team.Description, getteamRes.Description, Hint(GetTeam, "Description")); err != nil {
			return err
		}

		// SubmissionCount
		if err := IsuAssert(len(team.SubmissionIDs), getteamRes.SubmissionCount, Hint(GetTeam, "SubmissionCount")); err != nil {
			return err
		}
		
		// InvitationCode
		if isleader {
			if err := IsuAssert(team.InvitationCode, getteamRes.InvitationCode, Hint(GetTeam, "InvitationCode")); err != nil {
				return err
			}
		} else {
			if err := IsuAssert("", getteamRes.InvitationCode, Hint(GetTeam, "InvitationCode")); err != nil {
				return err
			}
		}

		return nil
	}
}
