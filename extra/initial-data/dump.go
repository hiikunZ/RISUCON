package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	DumpDir     = "dump"
	SQLdumpfile = "dump.sql"
)

func LoadDeafultTasks(jsonFile string) []Task {
	file, err := os.Open(jsonFile)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	tasks := []Task{}

	decoder := json.NewDecoder(file)

	if err = decoder.Decode(&tasks); err != nil {
		fmt.Println(err)
	}

	return tasks
}

func DumpUserstoJSON(users []User) {
	file, err := os.Create(DumpDir + "/users.json")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	if err = encoder.Encode(users); err != nil {
		fmt.Println(err)
	}
}

func DumpTeamstoJSON(teams []Team) {
	file, err := os.Create(DumpDir + "/teams.json")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	if err = encoder.Encode(teams); err != nil {
		fmt.Println(err)
	}
}

func DumpSubmissionstoJSON(submissions []Submission) {
	file, err := os.Create(DumpDir + "/submissions.json")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	if err = encoder.Encode(submissions); err != nil {
		fmt.Println(err)
	}
}

func DumpDatatoSQL(users []User, teams []Team, submissions []Submission, tasks []Task) {
	file, err := os.Create(DumpDir + "/" + SQLdumpfile)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	fmt.Fprintf(file, "TRUNCATE TABLE `users`;\n")
	fmt.Fprintf(file, "ALTER TABLE `users` AUTO_INCREMENT = 1;\n")
	fmt.Fprintf(file, "INSERT INTO `users` (`id`, `name`, `display_name`, `description`, `passhash`) VALUES\n")
	for i, user := range users {
		pashhash := fmt.Sprintf("%x", sha256.Sum256([]byte(user.Password)))
		fmt.Fprintf(file, "(%d, '%s', '%s', '%s', '%s')", user.ID, user.Name, user.DisplayName, user.Description, pashhash)
		if i != len(users)-1 {
			fmt.Fprintf(file, ",\n")
		} else {
			fmt.Fprintf(file, ";\n")
		}
	}

	fmt.Fprintf(file, "TRUNCATE TABLE `teams`;\n")
	fmt.Fprintf(file, "ALTER TABLE `teams` AUTO_INCREMENT = 1;\n")
	fmt.Fprintf(file, "INSERT INTO `teams` (`id`, `name`, `display_name`, `leader_id`, `member1_id`, `member2_id`, `description`, `invitation_code`) VALUES\n")
	for i, team := range teams {
		fmt.Fprintf(file, "(%d, '%s', '%s', %d, %d, %d, '%s', '%s')", team.ID, team.Name, team.DisplayName, team.LeaderID, team.Member1ID, team.Member2ID, team.Description, team.InvitationCode)
		if i != len(teams)-1 {
			fmt.Fprintf(file, ",\n")
		} else {
			fmt.Fprintf(file, ";\n")
		}
	}
	fmt.Fprintf(file, "TRUNCATE TABLE `tasks`;\n")
	fmt.Fprintf(file, "ALTER TABLE `tasks` AUTO_INCREMENT = 1;\n")
	fmt.Fprintf(file, "INSERT INTO `tasks` (`id`, `name`, `display_name`, `statement`, `submission_limit`) VALUES\n")
	for i, task := range tasks {
		fmt.Fprintf(file, "(%d, '%s', '%s', '%s', %d)", task.ID, task.Name, task.DisplayName, strings.Replace(task.Statement, "\\", "\\\\", -1), task.SubmissionLimit)
		if i != len(tasks)-1 {
			fmt.Fprintf(file, ",\n")
		} else {
			fmt.Fprintf(file, ";\n")
		}
	}
	fmt.Fprintf(file, "TRUNCATE TABLE `subtasks`;\n")
	fmt.Fprintf(file, "ALTER TABLE `subtasks` AUTO_INCREMENT = 1;\n")
	fmt.Fprintf(file, "INSERT INTO `subtasks` (`id`, `name`, `display_name`, `task_id`, `statement`) VALUES\n")
	for i, task := range tasks {
		for j, subtask := range task.SubTasks {
			fmt.Fprintf(file, "(%d, '%s', '%s', %d, '%s')", subtask.ID, subtask.Name, subtask.DisplayName, subtask.TaskID, strings.Replace(subtask.Statement, "\\", "\\\\", -1))
			if i != len(tasks)-1 || j != len(task.SubTasks)-1 {
				fmt.Fprintf(file, ",\n")
			} else {
				fmt.Fprintf(file, ";\n")
			}
		}
	}
	fmt.Fprintf(file, "TRUNCATE TABLE `answers`;\n")
	fmt.Fprintf(file, "ALTER TABLE `answers` AUTO_INCREMENT = 1;\n")
	fmt.Fprintf(file, "INSERT INTO `answers` (`id`, `task_id`, `subtask_id`, `answer`, `score`) VALUES\n")
	for i, task := range tasks {
		for j, subtask := range task.SubTasks {
			for k, answer := range subtask.Answers {
				fmt.Fprintf(file, "(%d, %d, %d, '%s', %d)", answer.ID, answer.TaskID, answer.SubtaskID, answer.Answer, answer.Score)
				if i != len(tasks)-1 || j != len(task.SubTasks)-1 || k != len(subtask.Answers)-1 {
					fmt.Fprintf(file, ",\n")
				} else {
					fmt.Fprintf(file, ";\n")
				}
			}
		}
	}
	fmt.Fprintf(file, "TRUNCATE TABLE `submissions`;\n")
	fmt.Fprintf(file, "ALTER TABLE `submissions` AUTO_INCREMENT = 1;\n")
	fmt.Fprintf(file, "INSERT INTO `submissions` (`id`, `task_id`, `user_id`, `submitted_at`, `answer`) VALUES\n")
	for i, submission := range submissions {
		submittedat := submission.SubmittedAt.Format("2006-01-02 15:04:05")
		fmt.Fprintf(file, "(%d, %d, %d, '%s', '%s')", submission.ID, submission.TaskID, submission.UserID, submittedat, submission.Answer)
		if i != len(submissions)-1 {
			fmt.Fprintf(file, ",\n")
		} else {
			fmt.Fprintf(file, ";\n")
		}
	}
}
