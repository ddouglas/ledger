CREATE TABLE `institutions` (
    `id` VARCHAR(50) NOT NULL COLLATE 'utf8mb4_general_ci',
    `name` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_general_ci',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`) USING BTREE
) COLLATE = 'utf8mb4_general_ci' ENGINE = InnoDB;