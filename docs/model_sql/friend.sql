drop table if exists biz_core.friend;
create table biz_core.friend
(
    id         int primary key auto_increment,
    uid        int         not null,
    fid        int         not null,
    intimacy   int         not null comment '亲密度',
    created_at datetime(3) not null default current_timestamp(3),
    updated_at datetime(3) not null default current_timestamp(3) on update current_timestamp(3),
    unique key uk_main (uid, fid),
    key idx1 (fid),
    key idx2 (intimacy),
    key idx3 (created_at)
) comment '好友表(一对好友存在两条数据)';

drop table if exists biz_core.block;
create table biz_core.block
(
    id         int primary key auto_increment,
    uid        int         not null,
    bid        int         not null,
    created_at datetime(3) not null default current_timestamp(3),
    updated_at datetime(3) not null default current_timestamp(3) on update current_timestamp(3),
    unique key uk_main (uid, bid),
    key idx1 (bid),
    key idx2 (created_at)
)