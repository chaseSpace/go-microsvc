package shumei

type Config struct {
	AccessKey    string `mapstructure:"access_key"`
	Appid        string `mapstructure:"appid"`
	UIDCryptoKey string `mapstructure:"uid_crypto_key"`
}
