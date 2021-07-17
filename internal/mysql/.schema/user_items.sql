CREATE TABLE `user_items` (
    `user_id` CHAR(64) NOT NULL COLLATE 'utf8mb4_general_ci',
    `item_id` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_general_ci',
    `access_token` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_general_ci',
    `institution_id` VARCHAR(32) NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `webhook` VARCHAR(255) NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `error` VARCHAR(255) NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `available_products` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_general_ci',
    `billed_products` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_general_ci',
    `consent_expiration_time` DATETIME NULL DEFAULT NULL,
    `update_type` ENUM('background', 'user_present_required') NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `item_status` JSON NULL DEFAULT NULL,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`user_id`, `item_id`) USING BTREE,
    INDEX `item_id` (`item_id`) USING BTREE,
    CONSTRAINT `user_items_user_id_users_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `ledger`.`users` (`id`) ON UPDATE CASCADE ON DELETE CASCADE
) COLLATE = 'utf8mb4_general_ci' ENGINE = InnoDB;