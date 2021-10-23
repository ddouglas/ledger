CREATE TABLE `merchants` (
    `id` VARCHAR(255) NOT NULL COLLATE 'utf8_bin',
    `name` VARCHAR(255) NOT NULL COLLATE 'utf8_bin',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`) USING BTREE
) COLLATE = 'utf8_bin' ENGINE = InnoDB;