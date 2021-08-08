CREATE TABLE `plaid_categories` (
    `id` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_bin',
    `name` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_bin',
    `group` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_bin',
    `hierarchy` JSON NOT NULL,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`) USING BTREE
) COLLATE = 'utf8mb4_bin' ENGINE = InnoDB;