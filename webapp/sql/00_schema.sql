CREATE TABLE `users` (
    `id` INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(255) NOT NULL UNIQUE,
    `display_name` VARCHAR(255) NOT NULL,
    `description` TEXT NOT NULL,
    `passhash` VARCHAR(255) NOT NULL
) ENGINE=InnoDB CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;

CREATE TABLE `teams` (
    `id` INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(255) NOT NULL UNIQUE,
    `display_name` VARCHAR(255) NOT NULL,
    `leader_id` INT NOT NULL,
    `member1_id` INT DEFAULT NULL,
    `member2_id` INT DEFAULT NULL,
    `description` TEXT NOT NULL,
    `invitation_code` VARCHAR(255) NOT NULL
) ENGINE=InnoDB CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;

CREATE TABLE `problems` (
    `id` INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(255) NOT NULL UNIQUE,
    `display_name` VARCHAR(255) NOT NULL,
    `statement` TEXT NOT NULL,
    `score` INT NOT NULL
) ENGINE=InnoDB CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;

CREATE TABLE `answers` (
    `id` INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    `problem_id` INT NOT NULL,
    `answer` VARCHAR(255) NOT NULL,
    `score` INT NOT NULL
) ENGINE=InnoDB CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;

CREATE TABLE `submissions` (
    `id` INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
    `problem_id` INT NOT NULL,
    `user_id` INT NOT NULL,
    `submitted_at` DATETIME NOT NULL,
    `answer` VARCHAR(255) NOT NULL,
) ENGINE=InnoDB CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;

