CREATE TABLE `transaction_location` (
    `item_id` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_general_ci',
    `transaction_id` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_general_ci',
    `address` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_general_ci',
    `city` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_general_ci',
    `lat` DOUBLE NOT NULL DEFAULT '0',
    `lon` DOUBLE NOT NULL DEFAULT '0',
    `region` VARCHAR(64) NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `store_number` VARCHAR(64) NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `postal_code` VARCHAR(64) NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `country` VARCHAR(64) NULL DEFAULT NULL COLLATE 'utf8mb4_general_ci',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`item_id`, `transaction_id`) USING BTREE,
    INDEX `transaction_id` (`transaction_id`) USING BTREE,
    CONSTRAINT `transaction_location_transaction_id_transactions_id_foregin` FOREIGN KEY (`transaction_id`) REFERENCES `ledger`.`transactions` (`item_id`) ON UPDATE CASCADE ON DELETE CASCADE
) COLLATE = 'utf8mb4_general_ci' ENGINE = InnoDB;