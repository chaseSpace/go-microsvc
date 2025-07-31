package commadmin

import "microsvc/protocol/svc/commonpb"

type SwitchKey string

const (
	SwitchKeyReviewUserNickname    SwitchKey = "review_user_nickname"
	SwitchKeyReviewUserFirstname   SwitchKey = "review_user_firstname"
	SwitchKeyReviewUserLastname    SwitchKey = "review_user_lastname"
	SwitchKeyReviewUserIcon        SwitchKey = "review_user_icon"
	SwitchKeyReviewUserDescription SwitchKey = "review_user_description"
	SwitchKeyAIReviewUserNewMoment SwitchKey = "ai_review_new_moment"
)

// ---------------0-------------------

// 扩展的开关值一定大于1

const (
	SwitchValExt_XXX commonpb.SwitchValue = 2
)

func DefaultSwitchValue(skey SwitchKey) *SwitchItem {
	return defaultSwitchValues[skey]
}

var defaultSwitchValues = map[SwitchKey]*SwitchItem{
	SwitchKeyReviewUserNickname: {
		&commonpb.SwitchItem{
			Core: &commonpb.SwitchItemCore{
				Key:      string(SwitchKeyReviewUserNickname),
				Name:     "用户昵称审核开关",
				Value:    commonpb.SwitchValue_ST_Off,
				ValueExt: nil,
				IsLock:   false,
			},
		},
	},
	SwitchKeyAIReviewUserNewMoment: {
		&commonpb.SwitchItem{
			Core: &commonpb.SwitchItemCore{
				Key:      string(SwitchKeyAIReviewUserNewMoment),
				Name:     "用户发布动态AI审核开关",
				Value:    commonpb.SwitchValue_ST_Off,
				ValueExt: nil,
				IsLock:   false,
			},
		},
	},
}

// ---------------------

type SwitchItem struct {
	*commonpb.SwitchItem
}

// IsOpen 对于开关选项只有开或关的情况，可以调用此方法
func (s *SwitchItem) IsOpen() bool {
	return s != nil && s.Core.Value == commonpb.SwitchValue_ST_On // 所有开关必须遵循0关1开的原则
}

// IsClose 对于开关选项只有开或关的情况，可以调用此方法
func (s *SwitchItem) IsClose() bool {
	return s.Core.Value == commonpb.SwitchValue_ST_Off // 所有开关必须遵循0关1开的原则
}

// Equal 判断开关是否等于某个值，用于复杂开关
func (s *SwitchItem) Equal(value commonpb.SwitchValue) bool {
	return s.Core.Value == value
}
