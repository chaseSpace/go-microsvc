create
    database if not exists biz_core DEFAULT CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci; -- 业务核心库
create
    database if not exists biz_core_log DEFAULT CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci; -- 业务核心库日志库
create
    database if not exists biz_admin DEFAULT CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci; -- 后台管理
create
    database if not exists micro_gateway DEFAULT CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci; -- 网关专用
create
    database if not exists micro_svc DEFAULT CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci; -- 微服务基础设施专用
create
    database if not exists gva DEFAULT CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci; -- 开源admin
