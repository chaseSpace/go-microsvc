DROP PROCEDURE IF EXISTS CreateUserData;
DELIMITER $$

CREATE PROCEDURE CreateUserData(IN num INT)
BEGIN
    DECLARE i INT DEFAULT 1;
    DECLARE base_uid INT DEFAULT num+RAND(1000); -- 以num作为UID的基础值
    DECLARE base_nid INT DEFAULT num+RAND(1000); -- 以num作为NID的基础值
    TRUNCATE TABLE user;
    WHILE i <= num DO
        INSERT INTO user (
            uid, nid, avatar, nickname, description, birthday, sex, password, password_salt, phone, reg_channel, reg_type, email, created_at, updated_at
        ) VALUES (
            base_uid + i, -- 确保UID是唯一的，基于num和循环计数
            base_nid + i, -- 确保NID是唯一的，基于num和循环计数
            CONCAT('http://example.com/avatar', i, '.jpg'), -- 随机头像URL
            CONCAT('User', i), -- 随机昵称
            CONCAT('Description ', i), -- 随机用户签名
            DATE_FORMAT(DATE_ADD('1970-01-01', INTERVAL FLOOR(RAND() * 20000) DAY), '%Y-%m-%d'), -- 随机生日
            FLOOR(RAND() * 3), -- 随机性别
            CONCAT('password', i), -- 随机密码
            CONCAT(FLOOR(RAND() * 10), FLOOR(RAND() * 10)), -- 随机密码盐
            CONCAT('+86', FLOOR(RAND() * 100000000) + 100000000), -- 随机手机号
            'Web', -- 随机注册渠道
            FLOOR(RAND() * 5), -- 随机注册类型
            CONCAT('user', i, '@example.com'), -- 随机邮箱
            CURRENT_TIMESTAMP(3), -- 当前时间戳
            CURRENT_TIMESTAMP(3) -- 当前时间戳
        );

        SET i = i + 1;
    END WHILE;
END$$

DELIMITER ;

CALL CreateUserData(1000); -- 这将生成10条随机数据