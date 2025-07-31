drop table if exists biz_core.user;
create table biz_core.user
(
    id            int primary key auto_increment,
    uid           int                          not null comment '用户id',
    nid           int                          null comment '靓号id',
    avatar        varchar(200)                 not null comment '头像',
    nickname      varchar(20) charset utf8mb4  not null,
    firstname     varchar(20) charset utf8mb4  not null comment '外国人，名',
    lastname      varchar(20) charset utf8mb4  not null comment '外国人，姓',
    description   varchar(200) charset utf8mb4 not null comment '用户签名',
    birthday      date                         not null,
    sex           tinyint unsigned             not null comment '0-未知 1男2女',
    password      varchar(40)                  not null,
    password_salt varchar(10)                  not null,
    phone         varchar(30)                  null,
    reg_channel   varchar(20)                  not null comment '注册渠道',
    reg_type      tinyint                      not null comment '注册类型：见PB：commonpb.SignInType',
    email         varchar(150)                 not null comment '邮箱',
    created_at    datetime(3)                  not null default current_timestamp(3),
    updated_at    datetime(3)                  not null default current_timestamp(3) on update current_timestamp(3),
    unique key idx_uid (uid),
    unique key idx_phone (phone),
    unique key idx_nid (nid),
    key idx_ct (created_at)
) default character set utf8mb4 comment '用户核心表';

# alter table biz_core.user add column email varchar(150) not null comment '邮箱';

drop table if exists biz_core.user_ext;
create table biz_core.user_ext
(
    uid                     int key          not null comment '用户id，关联主表user',
    voice_url               varchar(200)     not null comment '语音签名',
    education               tinyint unsigned not null comment '学历：初中 高中 大专 大学等',
    height                  tinyint unsigned not null comment '身高:cm, >=140',
    Weight                  tinyint unsigned not null comment '体重:kg, >=30',
    emotional               tinyint unsigned not null comment '情感状态：枚举：见PB: commonpb.EmotionalType',
    year_income             tinyint unsigned not null comment '年收入w范围，枚举：见PB: commonpb.YearIncomeType',
    occupation              varchar(20)      not null comment '职业：自填',
    hometown                varchar(20)      not null comment '籍贯/家乡',
    living_house            tinyint unsigned not null comment '居住方式：见PB: commonpb.LivingHouseType',
    house_buying            tinyint unsigned not null comment '购房情况：见PB: commonpb.HouseBuyingType',
    car_buying              tinyint unsigned not null comment '购车情况：见PB: commonpb.CarBuyingType',
    university              varchar(20)      not null comment '毕业院校',
    tags                    json             not null comment '其他标签：[x1, x2]',
    is_realperson_certified tinyint unsigned not null comment '是否真人认证：0-否 1-是（头像和人脸识别匹配）',
    is_realname_certified   tinyint unsigned not null comment '是否实名认证：0-否 1-是（头像和身份证匹配）'
) default charset utf8mb4 comment '用户扩展信息';

drop table if exists biz_core.user_register_weixin;
create table biz_core.user_register_weixin
(
    id         int primary key auto_increment,
    uid        int                         not null comment '用户id',
    account    varchar(100)                not null comment 'openid',
    nickname   varchar(20) charset utf8mb4 not null comment '微信昵称',
    union_id   varchar(100)                not null comment 'union_id',
    type       tinyint unsigned            not null comment '0-app 1-小程序 2-公众号',
    created_at datetime(3)                 not null default current_timestamp(3),
    updated_at datetime(3)                 not null default current_timestamp(3) on update current_timestamp(3),
    unique key uk_uid (uid),
    unique key uk_account (account),
    key idx_ct (created_at)
) default charset utf8mb4 comment '用户表-微信平台注册（uid关联用户核心表）';

drop table if exists biz_core.user_register_th;
create table biz_core.user_register_th
(
    id         int primary key auto_increment,
    uid        int          not null comment '用户id',
    account    varchar(100) not null comment 'email/三方ID',
    th_type    tinyint      not null comment '第三方类型：见PB: commonpb.SignInType（限三方登录类型）',
    created_at datetime(3)  not null default current_timestamp(3),
    updated_at datetime(3)  not null default current_timestamp(3) on update current_timestamp(3),
    unique key uk_uid (uid),
    unique key uk_account_thtype (account, th_type),
    key idx_ct (created_at)
) default charset utf8mb4 comment '用户表-第三方注册（如谷歌，uid关联用户核心表）';

