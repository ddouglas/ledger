CREATE TABLE `merchant_aliases` (
    `alias_id` VARCHAR(255) NOT NULL COLLATE 'utf8_bin',
    `merchant_id` VARCHAR(255) NOT NULL COLLATE 'utf8_bin',
    `alias` VARCHAR(255) NOT NULL COLLATE 'utf8_bin',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    INDEX `merchant_alias_merchant_id` (`merchant_id`) USING BTREE,
    CONSTRAINT `merchant_alias_merchant_id` FOREIGN KEY (`merchant_id`) REFERENCES `ledger`.`merchants` (`id`) ON UPDATE CASCADE ON DELETE CASCADE
) COLLATE = 'utf8_bin' ENGINE = InnoDB;