drop table if exists biz_core_log.sign_in_log;
create table biz_core_log.sign_in_log
(
    id         int primary key auto_increment,
    uid        int         not null,
    platform   tinyint     not null comment '枚举见PB: SignInPlatform',
    `system`   tinyint     not null comment '枚举见PB: SignInSystem',
    type       tinyint     not null comment '枚举见PB: SignInType',
    sign_in_at datetime(3) not null,
    ip         varchar(20) not null,
    created_at datetime(3) not null default current_timestamp(3),
    key idx_uid (uid),
    key idx_platform (platform),
    key idx_system (`system`),
    key idx_ip (ip),
    key idx_created_at (created_at)
) comment '登录日志';