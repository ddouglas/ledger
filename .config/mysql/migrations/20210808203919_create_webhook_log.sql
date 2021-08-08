CREATE TABLE `webhook_log` (
    `id` INT(11) NOT NULL AUTO_INCREMENT,
    `payload` JSON NOT NULL,
    `created_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`) USING BTREE
) COLLATE = 'latin1_swedish_ci' ENGINE = InnoDB;