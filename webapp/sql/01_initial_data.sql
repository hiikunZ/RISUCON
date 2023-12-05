TRUNCATE TABLE `users`;
ALTER TABLE `users` AUTO_INCREMENT = 1;
INSERT INTO `users` (`id`, `name`, `display_name`, `description`, `passhash`) VALUES
    (1, 'admin', '管理者', 'admin です', '8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918'),
    (2, 'isucon1', 'isucon1', 'テスト用ユーザー1', 'b88d6a9f342ee06400fa2b16664184b83cbd3406a6bba22d919c8f48a4af2cd3'),
    (3, 'isucon2', 'isucon2', 'テスト用ユーザー2', '35359832f0a86c2e80cc583177a299f6aefc1c8d71e81daf88d3f6bc993e3339'),
    (4, 'isucon3', 'isucon3', 'テスト用ユーザー3', 'eb4143a58bbec1fb0aed79bc9235558a559bd7cba13744809d9f6923f6492c96'),
    (5, 'isucon4', 'isucon4', 'テスト用ユーザー4', '087520a867907817d76563bcd7626f2c21ca72b7808f9cf684c3c99c88fc541d');

TRUNCATE TABLE `teams`;
ALTER TABLE `teams` AUTO_INCREMENT = 1;
INSERT INTO `teams` (`id`, `name`, `display_name`, `leader_id`, `member1_id`, `member2_id`, `description`, `invitation_code`) VALUES
    (1, 'team1', 'チーム1', 2, 3, 4, 'チーム1 です', 'fcvDhE6jj5bmoERRHrNi'),
    (2, 'team2', 'チーム2', 5, NULL, NULL, 'チーム2 です', 'GoGbP011M6o6UEi2pdOh');

TRUNCATE TABLE `tasks`;
ALTER TABLE `tasks` AUTO_INCREMENT = 1;
INSERT INTO `tasks` (`id`, `name`, `display_name`, `statement`) VALUES
    (1, 'A', '足し算', '足し算をしてください。'),
    (2, 'B', '引き算', '引き算をしてください。');

TRUNCATE TABLE `questions`;
ALTER TABLE `questions` AUTO_INCREMENT = 1;
INSERT INTO `questions` (`id`, `name`, `display_name`, `task_id`, `statement`) VALUES
    (1, 'A_1', '(1)', 1, '1+1=?'),
    (2, 'A_2', '(2)', 1, '1+2=?'),
    (3, 'B_1', '(1)', 2, '1-1=?'),
    (4, 'B_2', '(2)', 2, '1-2=? (符号のみが間違っていた場合、部分点が与えられる。)');

TRUNCATE TABLE `answers`;
ALTER TABLE `answers` AUTO_INCREMENT = 1;
INSERT INTO `answers` (`id`, `task_id`, `question_id`, `answer`, `score`) VALUES
    (1, 1, 1, '2', 10),
    (2, 1, 2, '3', 10),
    (3, 2, 3, '0', 10),
    (4, 2, 4, '-1', 10),
    (4, 2, 4, '1', 5);

TRUNCATE TABLE `submissions`;
ALTER TABLE `submissions` AUTO_INCREMENT = 1;
INSERT INTO `submissions` (`id`, `question_id`, `user_id`, `submitted_at`, `answer`) VALUES
    (1, 1, 2, '2012-06-20 00:00:00', '2'),
    (2, 2, 3, '2012-06-20 00:00:01', '3'),
    (3, 3, 4, '2012-06-20 00:00:02', '0'),
    (4, 4, 2, '2012-06-20 00:00:03', '-1'),
    (5, 4, 3, '2012-06-20 00:00:04', '1'),
    (6, 4, 5, '2012-06-20 00:00:04', '-1');