drop table if exists biz_core.gold_account;
create table biz_core.gold_account
(
    id             int primary key auto_increment,
    uid            int         not null,
    balance        int         not null default 0,
    recharge_total int         not null default 0 comment '累计充值总额',
    created_at     datetime(3) not null default current_timestamp(3),
    updated_at     datetime(3) not null default current_timestamp(3) on update current_timestamp(3),
    unique key uk_uid (uid),
    key idx_balance (balance),
    key idx_ut (updated_at)
) comment '核心表-用户虚拟货币：(金币)账户';

# 用户的金币使用记录（月表）
drop table if exists biz_core.gold_tx_log_$yyyymm;
create table if not exists biz_core.gold_tx_log_$yyyymm
(
    id         int primary key auto_increment,
    tx_id      varchar(40)                  not null COLLATE utf8mb4_bin comment '唯一交易id，大小写敏感',
    uid        int                          not null comment '消费用户id',
    delta      int                          not null comment '消费虚拟货币数额: 可正负',
    balance    int                          not null comment '消费后用户余额',
    tx_type    int                          not null comment '单人交易类型: 枚举见PB：GoldSingleTxType',
    remark     varchar(200) charset utf8mb4 not null comment '交易备注',
    created_at datetime(3)                  not null default current_timestamp(3),
    unique key uk_tx_id (tx_id),
    key idx_uid (uid),
    key idx_delta (delta),
    key idx_tx_type (tx_type),
    key idx_ct (created_at)
) comment '核心表-用户虚拟货币：(金币)交易记录(单用户)';
