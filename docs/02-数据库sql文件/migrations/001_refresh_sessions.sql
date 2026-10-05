-- 切换到短期 Access JWT + 轮换 Refresh Token 会话。
-- 此迁移会删除旧 user_token 表中的令牌；用户需要重新登录。
CREATE TABLE IF NOT EXISTS `refresh_sessions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` INT UNSIGNED NOT NULL,
  `family_id` CHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `token_hash` CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
  `expires_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `family_expires_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `revoked_at` TIMESTAMP NULL DEFAULT NULL,
  `replaced_by_hash` CHAR(64) CHARACTER SET ascii COLLATE ascii_bin NULL DEFAULT NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uidx_refresh_sessions_token_hash` (`token_hash`),
  INDEX `idx_refresh_sessions_user` (`user_id`),
  INDEX `idx_refresh_sessions_family` (`family_id`),
  INDEX `idx_refresh_sessions_expiry` (`expires_at`),
  INDEX `idx_refresh_sessions_family_expiry` (`family_expires_at`),
  CONSTRAINT `fk_refresh_sessions_user`
    FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='刷新令牌会话与轮换记录';

DROP TABLE IF EXISTS `user_token`;
