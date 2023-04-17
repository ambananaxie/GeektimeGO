CREATE DATABASE IF NOT EXISTS `userapp`
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `userapp`.users
(
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` varchar(128) NOT NULL,
    `avatar` varchar(512) NOT NULL ,
    `email` varchar(128) NOT NULL,
    `password` CHAR(128) NOT NULL,
    `create_time` INT UNSIGNED NOT NULL,
    `update_time` INT UNSIGNED NOT NULL,
    `salt` varchar(128) NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE (email)
) CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

# 影子库
CREATE DATABASE IF NOT EXISTS `userapp_shadow`
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `userapp_shadow`.users
(
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` varchar(128) NOT NULL,
    `avatar` varchar(512) NOT NULL ,
    `email` varchar(128) NOT NULL,
    `password` CHAR(128) NOT NULL,
    `create_time` INT UNSIGNED NOT NULL,
    `update_time` INT UNSIGNED NOT NULL,
    `salt` varchar(128) NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE (email)
) CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;


# 影子表
CREATE TABLE IF NOT EXISTS `userapp`.users_shadow
(
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` varchar(128) NOT NULL,
    `avatar` varchar(512) NOT NULL ,
    `email` varchar(128) NOT NULL,
    `password` CHAR(128) NOT NULL,
    `create_time` INT UNSIGNED NOT NULL,
    `update_time` INT UNSIGNED NOT NULL,
    `salt` varchar(128) NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE (email)
) CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;
