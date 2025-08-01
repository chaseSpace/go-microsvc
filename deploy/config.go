package deploy

import (
	"fmt"
	"io/fs"
	"microsvc/consts"
	"microsvc/enums"
	"microsvc/pkg/xerr"
	"microsvc/util"
	"path/filepath"
	"strings"

	"github.com/k0kubun/pp/v3"
	"github.com/spf13/viper"
)

// XConfig 是服务基础配置结构体
type XConfig struct {
	Svc                   enums.Svc         `mapstructure:"-"` // set by this.svcConf
	Env                   enums.Environment `mapstructure:"env"`
	Mysql                 map[string]*Mysql `mapstructure:"mysql"`
	Redis                 map[string]*Redis `mapstructure:"redis"`
	SimpleSdHttpPort      int               `mapstructure:"simplesd_http_port"`       // 本地简单注册中心的固定端口，可在配置修改
	SvcTokenSignKey       string            `mapstructure:"svc_token_sign_key"`       // 微服务鉴权token使用的key
	AdminTokenSignKey     string            `mapstructure:"admin_token_sign_key"`     // Admin鉴权token使用的key
	SensitiveInfoCryptKey string            `mapstructure:"sensitive_info_crypt_key"` // 敏感信息加密key（如手机号、身份证等）
	GRPC                  GRPCConfig        `mapstructure:"grpc"`
	// 动态端口
	dynamicGRPCPort int
	dynamicHTTPPort int

	// 固定端口
	GRPCPort int `mapstructure:"grpc_port"`
	HTTPPort int `mapstructure:"http_port"`

	ServiceDiscovery `mapstructure:"service_discovery"`

	// 开启api限流
	OpenAPIRateLimit bool `mapstructure:"open_api_rate_limit"`

	// 消息队列
	MqConfig `mapstructure:"mq_config"`

	// 外部消息通知
	ExternalNotify `mapstructure:"external_notify"`

	// 管理后台相关
	AdminSystem `mapstructure:"admin_system"`

	// 官网相关
	OfficialSite `mapstructure:"official_site"`

	// api网关相关
	ApiGateway `mapstructure:"api_gateway"`

	Stripe `mapstructure:"stripe"`

	// 默认文件路径
	DefaultAssets `mapstructure:"default_assets"`

	// 接管svc的配置
	svcConf SvcConfImpl
}

type ServiceDiscovery struct {
	FixedSvcIp string `mapstructure:"fixed_svc_ip"`
	Consul     struct {
		Address string `mapstructure:"address"`
	} `mapstructure:"consul"`
}

type GRPCConfig struct {
	LogPrintReqMaxLen  int `mapstructure:"log_print_req_max_len"`  // grpc 打印请求体最大长度
	LogPrintRespMaxLen int `mapstructure:"log_print_resp_max_len"` // 打印response body的最大长度
}

type MqConfig struct {
	Kafka struct {
		Brokers []string `mapstructure:"brokers"`
	} `mapstructure:"kafka"`
	Redis struct {
		Meta *Redis `mapstructure:"meta"`
	} `mapstructure:"redis"`
}

type ExternalNotify struct {
	Dingtalk map[string]*struct {
		Token  string `mapstructure:"token"`
		Secret string `mapstructure:"secret"`
	}
}

type AdminSystem struct {
	ImgBaseUrl string `mapstructure:"img_base_url"`
}

type OfficialSite struct {
	BaseUrl string `mapstructure:"base_url"`
}

type ApiGateway struct {
	BaseUrl string `mapstructure:"base_url"`
}

type Stripe struct {
	Key string `mapstructure:"key"`
}

func (s *XConfig) GetSvc() string {
	return s.Svc.Name()
}

func (s *XConfig) SetGRPC(port int) {
	s.dynamicGRPCPort = port
}

func (s *XConfig) SetHTTP(port int) {
	s.dynamicHTTPPort = port
}

func (s *XConfig) RegGRPCBase() (name string, addr string, port int) {
	return s.Svc.Name(), "", s.dynamicGRPCPort
}

func (s *XConfig) RegGRPCMeta() map[string]string {
	return nil
}

func (s *XConfig) GetSvcConf() SvcConfImpl {
	return s.svcConf
}

func (s *XConfig) GetConfDir(subPath ...string) string {
	return filepath.Join(append([]string{"deploy", s.Env.S()}, subPath...)...)
}

func (s *XConfig) IsDevEnv() bool {
	return s.Env == enums.EnvDev
}

func (s *XConfig) RequiredFieldCheck() {
	if s.Stripe.Key == "" {
		panic(xerr.ErrParams.New("stripe key is empty"))
	}
}

type Initializer func(cc *XConfig)

var XConf = &XConfig{}

var _ SvcListenPortSetter = new(XConfig)
var _ RegisterSvc = new(XConfig)

type DBname string

type Mysql struct {
	DBname
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	GormArgs string `mapstructure:"gorm_args"`
}

func (m Mysql) Dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s&timeout=3s", m.User, m.Password, m.Host, m.Port, m.DBname, m.GormArgs)
}

type Redis struct {
	DBname
	DB       int    `mapstructure:"db"`
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
}

// -------------------- 分割线 ----------------------------

func init() {
	pp.Default.SetExportedOnly(true)
}

func InitCfg(name enums.Svc, svcConfVar SvcConfImpl) {
	XConf.Svc = name
	XConf.Env = readEnv()

	// ------------- 先读取公共配置 -------------------

	// 设置配置文件所在的路径（可选，默认为当前目录）
	viper.AddConfigPath(XConf.GetConfDir())
	viper.SetConfigType("yaml")
	// load other yaml
	err := filepath.Walk(XConf.GetConfDir(), func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".yaml") {
			_, f := filepath.Split(path)

			_, _ = pp.Printf("...loading public-config file [%s]\n", f)

			viper.SetConfigName(f)
			err = viper.MergeInConfig()
		}
		return err
	})
	util.AssertNilErr(err)
	util.AssertNilErr(viper.Unmarshal(XConf))

	for dbname, v := range XConf.Mysql {
		v.DBname = DBname(dbname)
	}
	for dbname, v := range XConf.Redis {
		v.DBname = DBname(dbname)
	}

	__isPrintCfg := !util.ReadEnvBool(consts.EnvNoPrintCfg) // 默认是要打印配置

	_, _ = pp.Printf("************* Init public-config OK *************\n")
	if __isPrintCfg {
		_, _ = pp.Printf("%+v\n\n", XConf)
	}

	// 自检
	XConf.RequiredFieldCheck()

	if svcConfVar != nil {
		__loadSvcCfg(name, svcConfVar, __isPrintCfg)
	}

	_, _ = pp.Printf("************* %s *************", "ENV="+XConf.Env)
	fmt.Println(XConf.Env.AsciiGraphic())
}

func __loadSvcCfg(name enums.Svc, svcConfVar SvcConfImpl, __isPrintCfg bool) {
	//wd, _ := os.Getwd()
	//fmt.Println("getwd", wd)
	// ------------- 下面读取svc专有配置 -------------------
	fmt.Println() // \n

	viper.Reset()

	cfgDir := fmt.Sprintf("service/%s/deploy/%s", name.Name(), XConf.Env)

	viper.SetConfigType("yaml")
	viper.AddConfigPath(cfgDir)

	err := filepath.Walk(cfgDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".yaml") {
			_, f := filepath.Split(path)

			_, _ = pp.Printf("...loading service-config file [%s]\n", f)

			viper.SetConfigName(f)
			util.AssertNilErr(viper.MergeInConfig())
		}
		return nil
	})

	util.AssertNilErr(err)
	util.AssertNilErr(viper.Unmarshal(svcConfVar))

	logLv := readLogLevelFromEnvVar()
	if logLv != "" {
		svcConfVar.OverrideLogLevel(logLv)

		_, _ = pp.Printf("************* read log level from env: %s\n", logLv)
	}

	if name != svcConfVar.Name() {
		panic(fmt.Sprintf("%s not match name name:%s in config file", name, svcConfVar.Name()))
	}
	_, _ = pp.Printf("************* Init service-config OK *************\n")
	if __isPrintCfg {
		_, _ = pp.Printf("%+v\n\n", svcConfVar)
	} else {
		fmt.Println() //\n
	}

	// self check
	util.AssertNilErr(svcConfVar.SelfCheck())

	//  Service Conf 嵌入主配置对象
	XConf.svcConf = svcConfVar
}
