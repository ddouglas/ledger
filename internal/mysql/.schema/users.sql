CREATE TABLE `users` (
    `id` CHAR(64) NOT NULL COLLATE 'utf8mb4_general_ci',
    `name` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_general_ci',
    `email` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_general_ci',
    `auth0_subject` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_general_ci',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE INDEX `email` (`email`) USING BTREE
) COLLATE = 'utf8mb4_general_ci' ENGINE = InnoDB;