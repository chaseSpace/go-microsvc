drop table if exists biz_core.sys_world_county;
create table if not exists biz_core.sys_world_county
(
    id          int primary key auto_increment,
    country     varchar(50) not null,
    city        varchar(50) not null,
    state_short varchar(10) not null comment '州或省简写，如NY代指NewYork',
    state_full  varchar(50) not null comment '州或省全称，如New York',
    county      varchar(50) not null comment '县',
    city_alias  varchar(50) not null comment '城市别名',
    zip         varchar(10) not null comment '邮编',
    key (city)
) default character set utf8mb4 comment '世界城市-县表';

drop table if exists biz_core.sys_world_city;
create table biz_core.sys_world_city
(
    id          int primary key auto_increment,
    country     varchar(50) not null,
    city        varchar(50) not null,
    state_short varchar(10) not null comment '州或省简写，如NY代指NewYork',
    state_full  varchar(50) not null comment '州或省全称，如New York',
    biz_support tinyint     not null default 0 comment '业务是否支持',
    unique key uk_city (country, state_short, city)
) default character set utf8mb4 comment '世界城市表';
