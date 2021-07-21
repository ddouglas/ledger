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

CREATE TABLE `institutions` (
    `id` VARCHAR(50) NOT NULL COLLATE 'utf8mb4_bin',
    `name` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_bin',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`) USING BTREE
) COLLATE = 'utf8mb4_bin' ENGINE = InnoDB;

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

CREATE TABLE `accounts` (
    `item_id` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_bin',
    `account_id` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_bin',
    `mask` VARCHAR(32) NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `name` VARCHAR(128) NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `official_name` VARCHAR(128) NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `balance_available` DOUBLE NOT NULL DEFAULT '0',
    `balance_current` DOUBLE NOT NULL DEFAULT '0',
    `balance_limit` DOUBLE NOT NULL DEFAULT '0',
    `balance_country_code` VARCHAR(16) NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `balance_last_updated` DATETIME NULL DEFAULT NULL,
    `iso_currency_code` VARCHAR(128) NOT NULL DEFAULT '' COLLATE 'utf8mb4_bin',
    `unofficial_currency_code` VARCHAR(128) NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `subtype` VARCHAR(128) NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `type` VARCHAR(128) NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `recalculate_balance` TINYINT(1) NOT NULL DEFAULT '0',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`item_id`, `account_id`) USING BTREE,
    INDEX `account_id` (`account_id`) USING BTREE,
    CONSTRAINT `FK_accounts_user_items` FOREIGN KEY (`item_id`) REFERENCES `ledger`.`user_items` (`item_id`) ON UPDATE CASCADE ON DELETE CASCADE
) COLLATE = 'utf8mb4_bin' ENGINE = InnoDB;

CREATE TABLE `transactions` (
    `item_id` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_bin',
    `account_id` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_bin',
    `transaction_id` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_bin',
    `pending_transaction_id` VARCHAR(64) NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `category_id` VARCHAR(64) NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `name` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_bin',
    `pending` TINYINT(1) NOT NULL,
    `payment_channel` ENUM('online', 'in store', 'other') NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `merchant_name` VARCHAR(255) NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `categories` VARCHAR(255) NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `unofficial_currency_code` VARCHAR(255) NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `iso_currency_code` VARCHAR(255) NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `amount` DOUBLE NOT NULL,
    `transaction_code` VARCHAR(64) NULL DEFAULT NULL COLLATE 'utf8mb4_bin',
    `authorized_date` DATE NULL DEFAULT NULL,
    `authorized_datetime` DATETIME NULL DEFAULT NULL,
    `date` DATE NOT NULL,
    `datetime` DATETIME NULL DEFAULT NULL,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`item_id`, `account_id`, `transaction_id`) USING BTREE,
    INDEX `transactions_account_id_idx` (`account_id`) USING BTREE,
    INDEX `date` (`date`) USING BTREE,
    INDEX `pending` (`pending`) USING BTREE,
    INDEX `transaction_id` (`transaction_id`) USING BTREE,
    CONSTRAINT `transactions_account_id_accounts_id` FOREIGN KEY (`account_id`) REFERENCES `ledger`.`accounts` (`account_id`) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT `transactions_item_id_items_id_foreign` FOREIGN KEY (`item_id`) REFERENCES `ledger`.`user_items` (`item_id`) ON UPDATE CASCADE ON DELETE CASCADE
) COLLATE = 'utf8mb4_bin' ENGINE = InnoDB;