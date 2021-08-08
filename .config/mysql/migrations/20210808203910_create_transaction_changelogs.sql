CREATE TABLE `transaction_changelogs` (
    `item_id` VARCHAR(64) NOT NULL COLLATE 'utf8_bin',
    `transaction_id` VARCHAR(64) NOT NULL COLLATE 'utf8_bin',
    `changelog_id` INT(11) NOT NULL AUTO_INCREMENT,
    `changelog` JSON NOT NULL,
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`item_id`, `transaction_id`, `changelog_id`) USING BTREE,
    INDEX `changelog_id` (`changelog_id`) USING BTREE
) COLLATE = 'utf8_bin' ENGINE = InnoDB;