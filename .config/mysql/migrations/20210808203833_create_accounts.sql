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