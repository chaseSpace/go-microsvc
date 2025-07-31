package modelsql

const APICallLogMonthTable = `create table if not exists micro_svc.api_call_log_$yyyymm
(
    id         bigint primary key auto_increment,
    uid        int                          not null comment '用户id，可能为0，表示无需登录的api，如注册、登录',
    api_name   varchar(20)                  not null comment 'api名称：例如 GetGiftInfo，取自 svc/*pb/ext或int的rpc接口名称',
    api_ctrl   varchar(100)                 not null comment 'api的控制器名称：例如 giftExt, 取自 svc/*pb/service 名称',
    req_ip     varchar(20)                  not null comment '请求ip',
    dur_ms     int                          not null comment '接口调用耗时，单位：毫秒',
    success    bool                         not null comment '是否调用成功',
    svc          varchar(20)                not null comment '服务名',
    from_gateway bool                       not null comment '是否来自网关直接调用',
    panic      bool                         not null comment '是否发生panic',
    err_msg    varchar(200) charset utf8mb4 not null comment '错误信息',
    created_at datetime(3)                  not null default current_timestamp(3),
    key idx_api_name (api_name),
    key idx_api_ctrl (api_ctrl),
    key idx_uid (uid),
    key idx_req_ip (req_ip),
    key idx_ct (created_at),
    key idx_dur_ms (dur_ms)
) comment '用户api调用日志';`
