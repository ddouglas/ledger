CREATE TABLE `user_items` (
    `user_id` CHAR(64) NOT NULL COLLATE 'utf8mb4_bin',
    `item_id` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_bin',
    `access_token` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_bin',
    `institution_id` VARCHAR(32) NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `webhook` VARCHAR(255) NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `error` VARCHAR(255) NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `available_products` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_bin',
    `billed_products` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_bin',
    `consent_expiration_time` DATETIME NULL DEFAULT NULL,
    `update_type` ENUM('background', 'user_present_required') NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `item_status` JSON NULL DEFAULT NULL,
    `is_refreshing` TINYINT(4) NOT NULL DEFAULT '0',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`user_id`, `item_id`) USING BTREE,
    INDEX `item_id` (`item_id`) USING BTREE,
    CONSTRAINT `user_items_user_id_users_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `ledger`.`users` (`id`) ON UPDATE CASCADE ON DELETE CASCADE
) COLLATE = 'utf8mb4_bin' ENGINE = InnoDB;