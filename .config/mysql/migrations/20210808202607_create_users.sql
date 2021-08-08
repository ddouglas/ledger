CREATE TABLE `users` (
    `id` CHAR(64) NOT NULL COLLATE 'utf8mb4_bin',
    `name` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_bin',
    `email` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_bin',
    `auth0_subject` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_bin',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE INDEX `email` (`email`) USING BTREE
) COLLATE = 'utf8mb4_bin' ENGINE = InnoDB;