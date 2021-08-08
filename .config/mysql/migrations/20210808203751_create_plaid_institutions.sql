CREATE TABLE `plaid_institutions` (
    `id` VARCHAR(50) NOT NULL COLLATE 'utf8mb4_bin',
    `name` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_bin',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`) USING BTREE
) COLLATE = 'utf8mb4_bin' ENGINE = InnoDB;