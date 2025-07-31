package enums

type Environment string

const (
	EnvDev  Environment = "dev"
	EnvBeta Environment = "beta"
	EnvProd Environment = "prod"
)

func (e Environment) S() string {
	return string(e)
}

func (e Environment) IsDev() bool {
	return e == EnvDev
}

func (e Environment) IsProd() bool {
	return e == EnvProd
}

func (e Environment) AsciiGraphic() string {
	switch e {
	case EnvDev:
		return `
ooooooooo   ooooooooooo ooooo  oooo
 888    88o  888    88   888    88
 888    888  888ooo8      888  88
 888    888  888    oo     88888
o888ooo88   o888ooo8888     888

`
	case EnvBeta:
		return `
oooooooooo  ooooooooooo ooooooooooo      o
 888    888  888    88  88  888  88     888
 888oooo88   888ooo8        888        8  88
 888    888  888    oo      888       8oooo88
o888ooo888  o888ooo8888    o888o    o88o  o888o

`
	case EnvProd:
		return `
oooooooooo  oooooooooo    ooooooo   ooooooooo
 888    888  888    888 o888   888o  888    88o
 888oooo88   888oooo88  888     888  888    888
 888         888  88o   888o   o888  888    888
o888o       o888o  88o8   88ooo88   o888ooo88

`
	}
	return `
    _  _  _
 _ (_)(_)(_)_
(_)        (_)
         _ (_)
      _ (_)
     (_)
      _
     (_)
`
}

type Svc string

// Name 请不要随意修改这个方法的逻辑，因为微服务客户端证书中的subjectAltName包含了这个Name
// 一旦修改，客户端证书需要重新申请，否则RPC请求将会失败：tls: bad certificate
// （一种不安全的做法是取消RPC的双向证书验证）
// 参阅：generate_cert_for_svc.md
func (s Svc) Name() string {
	if s == "" {
		return "unknown-svc"
	}
	return string(s)
}

const (
	SvcTemplate   Svc = "template"
	SvcGateway    Svc = "gateway"
	SvcAdmin      Svc = "admin"
	SvcUser       Svc = "user"
	SvcCurrency   Svc = "currency"
	SvcThirdparty Svc = "thirdparty"
	SvcFriend     Svc = "friend"
	SvcMoment     Svc = "moment" // 动态
	SvcGift       Svc = "gift"
	SvcMqConsumer Svc = "mqconsumer"
	SvcCrontask   Svc = "crontask"
	SvcCommander  Svc = "commander"
)
