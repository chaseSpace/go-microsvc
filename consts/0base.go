package consts

const (
	EnvVar         = "MICRO_SVC_ENV"
	EnvNoPrintCfg  = "MICRO_SVC_NO_PRINT_CFG"
	EnvVarLogLevel = "MICRO_SVC_LOG_LEVEL"
)

type CtxKey struct{}

type CtxValue struct {
}

const YearMonth = "$yyyymm"

const (
	AdminFileUploadReqPath    = "uploads/"        // 管理后台上传文件 对应的http请求路径
	MicroSvcFileUploadReqPath = "micro_uploads/"  // 微服务接口上传文件 对应的http请求路径
	ManualUploadReqPath       = "manual_uploads/" // 手动上传文件 对应的http请求路径
)
