package modelsql

const MqLogMonthTable = `create table if not exists biz_core_log.mq_log_$yyyymm
(
    id         bigint primary key auto_increment,
	topic           varchar(100) not null,
    topic_unique_id varchar(100) not null comment 'topic消息自身的唯一标识（非表唯一键，仅用于索引）',
    data            json         not null,
    created_at      datetime(3)  not null,
    key idx_topic (topic),
    key idx_unique_id (topic_unique_id),
    key idx_ct (created_at)
) comment '消息队列日志';`
