CREATE TABLE `transaction_changelog` (
    `transaction_id` VARCHAR(64) NOT NULL COLLATE 'utf8mb4_general_ci',
    `changelog` JSON NOT NULL,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    INDEX `FK_transaction_changelog_transactions` (`transaction_id`) USING BTREE,
    CONSTRAINT `FK_transaction_changelog_transactions` FOREIGN KEY (`transaction_id`) REFERENCES `ledger`.`transactions` (`transaction_id`) ON UPDATE CASCADE ON DELETE CASCADE
) COLLATE = 'utf8mb4_general_ci' ENGINE = InnoDB;