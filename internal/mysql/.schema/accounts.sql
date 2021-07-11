CREATE TABLE `accounts` (
    `user_id` CHAR(64) NOT NULL COLLATE 'utf8mb4_general_ci',
    `item_id` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_general_ci',
    `account_id` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_general_ci',
    `balance_available` INT(11) NULL DEFAULT NULL,
    `balance_current` INT(11) NULL DEFAULT NULL,
    `balance_country_code` VARCHAR(16) NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `balance_limit` INT(11) NULL DEFAULT NULL,
    `iso_current_code` VARCHAR(128) NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `unofficial_currency_codes` VARCHAR(128) NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`user_id`, `item_id`, `account_id`) USING BTREE,
    CONSTRAINT `accounts_user_id_users_id_foreign` FOREIGN KEY (`user_id`) REFERENCES `ledger`.`users` (`id`) ON UPDATE CASCADE ON DELETE CASCADE
) COLLATE = 'utf8mb4_general_ci' ENGINE = InnoDB;