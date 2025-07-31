drop table if exists biz_admin.review_text;
create table if not exists biz_admin.review_text
(
    id         bigint primary key auto_increment,
    uid        bigint           not null,
    text       varchar(500)     not null,
    biz_type   tinyint unsigned not null comment '枚举见PB: commonpb.BizType',
    status     tinyint unsigned not null comment '枚举见PB: commonpb.ReviewStatus',
    updated_by bigint           not null comment '上次操作管理员id',
    created_at datetime(3)      not null default current_timestamp(3),
    updated_at datetime(3)      not null default current_timestamp(3) on update current_timestamp(3),
    key idx_uid (uid),
    key idx_biz_type (biz_type),
    key idx_status (status),
    key idx_updated_by (updated_by),
    key idx_created_at (created_at)
) default charset utf8mb4;

drop table if exists biz_admin.review_image;
create table if not exists biz_admin.review_image
(
    id         bigint primary key auto_increment,
    uid        bigint           not null,
    text       varchar(500)     not null comment '图片可以伴随文本（如动态），同时审核',
    urls       json             not null comment 'json数组',
    biz_type   tinyint unsigned not null comment '枚举见PB: commonpb.BizType',
    status     tinyint unsigned not null comment '枚举见PB: commonpb.ReviewStatus',
    updated_by bigint           not null comment '上次操作管理员id',
    created_at datetime(3)      not null default current_timestamp(3),
    updated_at datetime(3)      not null default current_timestamp(3) on update current_timestamp(3),
    key idx_uid (uid),
    key idx_biz_type (biz_type),
    key idx_status (status),
    key idx_updated_by (updated_by),
    key idx_created_at (created_at)
) default charset utf8mb4;


drop table if exists biz_admin.review_video;
create table if not exists biz_admin.review_video
(
    id              bigint primary key auto_increment,
    uid             bigint           not null,
    text            varchar(500)     not null comment '视频可以伴随文本（如动态），同时审核',
    url             varchar(200)     not null,
    biz_type        tinyint unsigned not null comment '枚举见PB: commonpb.BizType',
    status          tinyint unsigned not null comment '枚举见PB: commonpb.ReviewStatus',
    biz_uniq_id     int              not null comment '源业务主键id',
    th_task_id      varchar(100)     not null comment '第三方审核任务id',
    th_name         varchar(20)      not null comment '使用哪个第三方审核服务',
    query_ret_fails int              not null comment '查询第三方审核结果失败次数',
    updated_by      bigint           not null comment '上次操作管理员id',
    created_at      datetime(3)      not null default current_timestamp(3),
    updated_at      datetime(3)      not null default current_timestamp(3) on update current_timestamp(3),
    key idx_uid (uid),
    key idx_biz_type (biz_type),
    key idx_status (status),
    key idx_updated_by (updated_by),
    key idx_th_task_id (th_task_id),
    key idx_created_at (created_at)
) default charset utf8mb4;


drop table if exists biz_admin.review_audio;
create table if not exists biz_admin.review_audio
(
    id              bigint primary key auto_increment,
    uid             bigint           not null,
    text            varchar(500)     not null comment '音频可以伴随文本（如动态），同时审核',
    url             varchar(200)     not null,
    biz_type        tinyint unsigned not null comment '枚举见PB: commonpb.BizType',
    status          tinyint unsigned not null comment '枚举见PB: commonpb.ReviewStatus',
    biz_uniq_id     int              not null comment '源业务主键id',
    th_task_id      varchar(100)     not null comment '第三方审核任务id',
    th_name         varchar(20)      not null comment '使用哪个第三方审核服务',
    query_ret_fails int              not null comment '查询第三方审核结果失败次数',
    updated_by      bigint           not null comment '上次操作管理员id',
    created_at      datetime(3)      not null default current_timestamp(3),
    updated_at      datetime(3)      not null default current_timestamp(3) on update current_timestamp(3),
    key idx_uid (uid),
    key idx_biz_type (biz_type),
    key idx_status (status),
    key idx_updated_by (updated_by),
    key idx_th_task_id (th_task_id),
    key idx_created_at (created_at)
) default charset utf8mb4;
