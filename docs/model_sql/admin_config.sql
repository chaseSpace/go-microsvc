drop table if exists biz_admin.admin_config_center;
create table biz_admin.admin_config_center
(
    id                   int primary key auto_increment,
    `key`                varchar(100) not null comment '配置键，不含空格，唯一',
    name                 varchar(50)  not null comment '配置中文名',
    value                text         not null comment '配置值',
    is_lock              bool         not null comment '是否锁定，true表示value默认在前端不可见，仅创建人可见&可改；且此选项优先级最高！' default false,
    allow_program_update bool         not null comment '允许程序修改'                                                                 default false,
    created_by           int          not null comment '添加人'                                                                       default 0,
    updated_by           int          not null comment '修改人: 0表示程序'                                                            default 0,
    created_at           datetime     not null                                                                                        default current_timestamp,
    updated_at           datetime     not null                                                                                        default current_timestamp on update current_timestamp,
    deleted_ts           int          not null                                                                                        default 0,
    unique key uk_ckey (`key`, deleted_ts)
) comment '配置中心表';

insert into biz_admin.admin_config_center (`key`, name, value, created_by, updated_by)
values ('spider:ins.scrape_bar_activity.account', 'account', '{"account": "leigege_ya", "pass":"1213is.."}', 0, 0);


drop table if exists biz_admin.admin_switch_center;
create table biz_admin.admin_switch_center
(
    id         int primary key auto_increment,
    `key`      varchar(100) not null comment '开关键，不含空格，唯一',
    name       varchar(50)  not null comment '开关中文名',
    value      tinyint      not null comment '开关值',
    value_ext  json         not null comment '扩展值，json格式',
    is_lock    bool         not null comment '是否锁定，true表示仅创建人可改！' default false,
    created_by int          not null comment '添加人'                         default 0,
    updated_by int          not null comment '修改人'                         default 0,
    created_at datetime     not null                                          default current_timestamp,
    updated_at datetime     not null                                          default current_timestamp on update current_timestamp,
    deleted_ts int          not null                                          default 0,
    unique key uk_skey (`key`, deleted_ts)
) comment '配置中心表';
