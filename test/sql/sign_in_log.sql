
DROP PROCEDURE IF EXISTS CreateSignInLogData;

DELIMITER $$

CREATE PROCEDURE CreateSignInLogData(IN num INT, IN base_uid INT)
BEGIN
    DECLARE i INT DEFAULT 1;
    DECLARE current_uid INT DEFAULT base_uid;

    WHILE i <= num DO
        INSERT INTO biz_core_log.sign_in_log (
            uid, platform, `system`, `type`, sign_in_at, ip
        ) VALUES (
            current_uid, -- 使用传入的base_uid作为基础UID
            FLOOR(1 + (RAND() * 4)), -- 假设platform有4个可能的值
            FLOOR(1 + (RAND() * 3)), -- 假设system有3个可能的值
            FLOOR(1 + (RAND() * 5)), -- 假设type有5个可能的值
            DATE_FORMAT(NOW() - INTERVAL FLOOR(RAND() * 10000) SECOND, '%Y-%m-%d %H:%i:%s.%f'), -- 随机的sign_in_at时间
            CONCAT(FLOOR(RAND() * 255), '.', FLOOR(RAND() * 255), '.', FLOOR(RAND() * 255), '.', FLOOR(RAND() * 255)) -- 随机的IP地址
        );

        SET current_uid = current_uid + 1;
        SET i = i + 1;
    END WHILE;
END$$

DELIMITER ;





