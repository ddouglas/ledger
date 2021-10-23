ALTER TABLE
    `transactions` CHANGE COLUMN `merchant_name` `merchant_id` VARCHAR(255) NULL DEFAULT NULL COLLATE 'utf8mb4_bin'
AFTER
    `payment_channel`;