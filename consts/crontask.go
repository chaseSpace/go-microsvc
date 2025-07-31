package consts

type BarActPlatform string

const (
	BarActPlatformInstagram BarActPlatform = "instagram"
)

func (v BarActPlatform) ToStr() string {
	return string(v)
}
