drop table if exists biz_core.punish;
create table biz_core.punish
(
    id         int primary key auto_increment,
    uid        int          not null comment '用户id',
    type       tinyint      not null comment '枚举见PB: PunishType',
    reason     varchar(200) not null collate utf8mb4_bin comment '惩罚原因',
    duration   int          not null comment '惩罚时长，单位秒',
    State      tinyint      not null comment '生效状态：枚举见PB: PunishState',
    created_by int          not null comment '创建人uid（admin）',
    updated_by int          not null comment '更新人uid（admin）',
    created_at datetime(3)  not null default current_timestamp(3),
    updated_at datetime(3)  not null default current_timestamp(3) on update current_timestamp(3),
    key idx_uid_type (uid, type),
    key idx_created_by (created_by),
    key idx_updated_by (updated_by),
    key idx_duration (duration),
    key idx_ct (created_at)
) default character set utf8mb4 comment '用户惩罚表(只含正在惩罚的用户记录)';

drop table if exists biz_core.punish_log;
create table biz_core.punish_log
(
    id          int primary key auto_increment,
    uid         int          not null comment '用户id',
    op_type     tinyint      not null comment '枚举见PB: PunishOpType（解除时为0）',
    punish_type tinyint      not null comment '枚举见PB: PunishType',
    reason      varchar(200) not null collate utf8mb4_bin comment '惩罚/解除原因',
    duration    int          not null comment '惩罚时长，单位秒（解除时为0）',
    created_by  int          not null comment '创建人uid（admin）',
    created_at  datetime(3)  not null default current_timestamp(3),
    key idx_uid (uid),
    key idx_created_by (created_by),
    key idx_ct (created_at)
) default character set utf8mb4 comment '用户惩罚操作日志';