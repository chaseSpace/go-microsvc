drop table if exists biz_core.gift_conf;
create table if not exists biz_core.gift_conf
(
    id               int primary key auto_increment,
    name             varchar(20)  not null collate utf8mb4_general_ci comment '礼物名称(忽略大小写)',
    icon             varchar(200) not null comment '礼物图标',
    price            int          not null comment '礼物价格(金币)',
    type             tinyint      not null comment '礼物类型: 枚举见PB: GiftType',
    state            tinyint      not null comment '礼物状态: 枚举见PB: GiftState',
    supported_scenes JSON         not null comment '礼物支持的场景(JSON数组): 包含的枚举见PB: GiftScene',
    created_at       datetime(3)  not null default current_timestamp(3),
    updated_at       datetime(3)  not null default current_timestamp(3) on update current_timestamp(3),
    unique key uk_name (name),
    key idx_price (price),
    key idx_type (type),
    key idx_ct (created_at),
    key idx_ut (updated_at)
) default character set utf8mb4 comment '礼物列表';

drop table if exists biz_core.gift_account;
create table if not exists biz_core.gift_account
(
    uid        int         not null comment '用户id',
    gift_id    int         not null comment '礼物id',
    amount     int         not null comment '礼物数量',
    created_at datetime(3) not null default current_timestamp(3),
    updated_at datetime(3) not null default current_timestamp(3) on update current_timestamp(3),
    unique key uk_uid_giftId (uid, gift_id),
    key idx_count (amount)
) default character set utf8mb4 comment '礼物账户';

# 礼物交易记录（一笔交易一条记录，包含单人、双人交易）
# -- 单表，方便查询统计
drop table if exists biz_core.gift_tx_log;
create table if not exists biz_core.gift_tx_log
(
    id          int primary key auto_increment,
    tx_id       varchar(40)  not null COLLATE utf8mb4_bin comment '唯一交易id，大小写敏感',
    from_uid    int          not null comment '交易来源用户id',
    to_uid      int          not null comment '交易目标用户id',
    gift_id     int          not null comment '交易礼物id',
    gift_name   varchar(20)  not null comment '交易礼物名称',
    price       int          not null comment '交易礼物价格(金币)',
    amount      int          not null comment '交易礼物数量: 正数',
    total_value int          not null comment '交易礼物总价值(正数): amount*price',
    tx_type     int          not null comment '交易类型: 枚举见PB：GiftTxType',
    gift_scene  int          not null comment '交易礼物场景: 枚举见PB：GiftTxScene',
    gift_type   int          not null comment '交易礼物类型: 枚举见PB：GiftType',
    remark      varchar(200) not null comment '交易备注',
    created_at  datetime(3)  not null default current_timestamp(3),
    updated_at  datetime(3)  not null default current_timestamp(3) on update current_timestamp(3),
    unique key uk_tx_id (tx_id),
    key idx_from_uid (from_uid),
    key idx_to_uid (to_uid),
    key idx_amount (amount),
    key idx_total_value (total_value),
    key idx_tx_type (tx_type),
    key idx_gift_scene (gift_scene),
    key idx_ct (created_at)
) default character set utf8mb4 comment '礼物双人交易记录表(全)';

# 单个用户的交易记录月表（一笔双人交易产生两条记录）
# -- 月表减少单表数据量
drop table if exists biz_core.gift_tx_log_personal_$yyyymm;
create table if not exists biz_core.gift_tx_log_personal_$yyyymm
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
