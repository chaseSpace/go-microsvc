drop table if exists biz_core.moment;
create table if not exists biz_core.moment
(
    id             bigint primary key auto_increment,
    uid            int                          not null,
    text           varchar(500) charset utf8mb4 not null,
    type           tinyint                      not null comment '枚举见PB: MomentType',
    review_status  tinyint                      not null comment '枚举见PB: momentpb.ReviewStatus',
    media_urls     JSON                         not null,
    likes          JSON                         not null comment '点赞UID数组',
    forwards       int                          not null comment '转发数',
    review_pass_at bigint                       not null comment '审核通过毫秒时间戳',
    created_at     datetime(3)                  not null default current_timestamp(3),
    updated_at     datetime(3)                  not null default current_timestamp(3) on update current_timestamp(3),
    deleted_at     datetime(3)                  null,
    key idx_uid (uid),
    key idx_type (type),
    key idx_forwards (forwards),
    key idx_created_at (created_at),
    key idx_deleted_at (deleted_at)
) comment '用户动态表';

-- test
# insert into biz_core.moment (uid, text, type, review_status, media_urls, likes, forwards)
#     value (1, 'test', 1, 1, '[
#   "http://test.com/1.jpg",
#   "http://test.com/2.jpg"
# ]', '{
#   "1": 1
# }', 1);
#
# # select json_set(likes, '$."2"', 2) from moment;
# UPDATE biz_core.moment
# SET likes = JSON_SET(likes, '$."1"', '')
# WHERE uid = 1;

drop table if exists biz_core.moment_comment;
create table if not exists biz_core.moment_comment
(
    id         bigint primary key auto_increment,
    mid        bigint                       not null,
    uid        int                          not null,
    reply_uid  bigint                       not null,
    content    varchar(500) charset utf8mb4 not null,
    created_at datetime(3)                  not null default current_timestamp(3),
    deleted_at datetime(3)                  null
) comment '动态评论';
