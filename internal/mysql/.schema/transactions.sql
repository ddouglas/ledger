CREATE TABLE `transactions` (
    `item_id` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_general_ci',
    `account_id` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_general_ci',
    `transaction_id` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_general_ci',
    `interator` INT(11) NOT NULL AUTO_INCREMENT,
    `pending_transaction_id` VARCHAR(64) NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `category_id` VARCHAR(64) NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `name` VARCHAR(255) NOT NULL COLLATE 'utf8mb4_general_ci',
    `pending` TINYINT(1) NOT NULL,
    `payment_channel` ENUM('online', 'in store', 'other') NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `merchant_name` VARCHAR(255) NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `categories` VARCHAR(255) NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `unofficial_currency_code` VARCHAR(255) NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `iso_currency_code` VARCHAR(255) NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `amount` DOUBLE NOT NULL,
    `transaction_code` VARCHAR(64) NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `authorized_date` DATE NULL DEFAULT NULL,
    `authorized_datetime` DATETIME NULL DEFAULT NULL,
    `date` DATE NOT NULL,
    `datetime` DATETIME NULL DEFAULT NULL,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`item_id`, `account_id`, `transaction_id`) USING BTREE,
    INDEX `transactions_account_id_idx` (`account_id`) USING BTREE,
    INDEX `date` (`date`) USING BTREE,
    INDEX `interator` (`interator`) USING BTREE,
    INDEX `pending` (`pending`) USING BTREE,
    CONSTRAINT `transactions_account_id_accounts_id` FOREIGN KEY (`account_id`) REFERENCES `ledger`.`accounts` (`account_id`) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT `transactions_item_id_items_id_foreign` FOREIGN KEY (`item_id`) REFERENCES `ledger`.`user_items` (`item_id`) ON UPDATE CASCADE ON DELETE CASCADE
) COLLATE = 'utf8mb4_general_ci' ENGINE = InnoDB AUTO_INCREMENT = 8776;