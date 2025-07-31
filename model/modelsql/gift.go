package modelsql

const GiftTxLogPersonalMonthTable = `create table if not exists gift_tx_log_personal_$yyyymm
(
    id                   int primary key auto_increment,
    tx_id                varchar(40)  not null COLLATE utf8mb4_bin comment '唯一交易id，大小写敏感',
    uid                  int          not null comment '交易用户id',
    gift_id              int          not null comment '交易礼物id',
    gift_name            varchar(20)  not null comment '交易礼物名称',
    price                int          not null comment '交易礼物价格(金币)',
    delta                int          not null comment '交易礼物数量: 可正负',
    balance              int          not null comment '交易后用户剩余数量',
    total_value          int          not null comment '交易礼物总价值(正数): delta*price',
    first_person_tx_type int          not null comment '第一人称交易类型: 枚举见PB：GiftFirstPersonTxType',
    gift_scene           int          not null comment '交易礼物场景: 枚举见PB：GiftTxScene',
    gift_type            int          not null comment '交易礼物类型: 枚举见PB：GiftType',
    related_uid          int          not null comment '交易相关用户id（受赠人或赠送人，若无是自己，根据交易类型决定）',
    remark               varchar(200) not null comment '交易备注',
    created_at           datetime(3)  not null default current_timestamp(3),
    updated_at           datetime(3)  not null default current_timestamp(3),
    key tx_id (tx_id) comment '关联主交易表的tx_id',
    key idx_uid (uid),
    key idx_delta (delta),
    key idx_total_value (total_value),
    key idx_gift_scene (gift_scene),
    key idx_fp_tx_type (first_person_tx_type),
    key idx_ct (created_at)
) default character set utf8mb4 comment '礼物交易记录表(指定用户)';
`
