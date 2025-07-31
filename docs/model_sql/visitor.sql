drop table if exists biz_core.visitor;
create table biz_core.visitor
(
    id                 int primary key auto_increment,
    uid                int         not null,
    vid                int         not null,
    day_visit_times    int         not null comment '当日访问次数',
    day_visit_duration int         not null comment '当日访问时长',
    date               int         not null comment 'yyyymmdd，访问日期',
    created_at         datetime(3) not null default current_timestamp(3),
    updated_at         datetime(3) not null default current_timestamp(3) on update current_timestamp(3),
    unique key uk_main (uid, vid, date),
    key idx_vid (vid),
    key idx_ut (updated_at)
);

# INSERT INTO biz_core.visitor (uid, vid, day_visit_times, day_visit_duration, date, created_at, updated_at)
# VALUES (?, ?, 1, ?, REPLACE(DATE(NOW()), '-', ''), NOW(), NOW())
# ON DUPLICATE KEY UPDATE day_visit_times    = day_visit_times + 1,
#                         day_visit_duration = day_visit_duration + ?,
#                         updated_at         = NOW();