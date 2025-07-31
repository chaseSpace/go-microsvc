drop table if exists biz_core.review_text;
create table biz_core.review_text
(
    id         int primary key auto_increment,
    uid        int              not null,
    text       varchar(500)     not null,
    text_type  tinyint unsigned not null comment '0-用户昵称 2-用户签名',
    status     tinyint unsigned not null comment '0-待审核 1-违规 2-机审通过 3-人审通过',
    admin_uid  int,
    admin_note varchar(100),
    created_at datetime(3)      not null default current_timestamp(3),
    updated_at datetime(3)      not null default current_timestamp(3) on update current_timestamp(3),
    key idx_uid (uid),
    key idx_ct (created_at),
    key idx_typ_status (text_type, status),
    key idx_admin_uid (admin_uid)
) comment '审核文本表';

drop table if exists biz_core.review_image;
create table biz_core.review_image
(
    id         int primary key auto_increment,
    uid        int              not null,
    img_url    varchar(500)     not null,
    img_type   tinyint unsigned not null comment '0-头像 1-相册图',
    status     tinyint unsigned not null comment '0-待审核 1-违规 2-机审通过 3-人审通过',
    admin_uid  int,
    admin_note varchar(100),
    created_at datetime(3)      not null default current_timestamp(3),
    updated_at datetime(3)      not null default current_timestamp(3) on update current_timestamp(3),
    key idx_uid (uid),
    key idx_ct (created_at),
    key idx_typ_status (img_type, status),
    key idx_admin_uid (admin_uid)
) comment '审核图片表';