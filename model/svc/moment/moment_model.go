package moment

import (
	"microsvc/model"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/momentpb"
)

type likeMap map[int64]*struct{}

type Moment struct {
	model.FieldID
	model.FieldUID
	model.FieldAt
	model.FieldDeletedAt
	Text         string                `gorm:"column:text" json:"text"`
	Type         momentpb.MomentType   `gorm:"column:type" json:"type"`
	ReviewStatus momentpb.ReviewStatus `gorm:"column:review_status" json:"review_status"`
	MediaUrls    []string              `gorm:"column:media_urls; serializer:json" json:"media_urls"`
	Likes        likeMap               `gorm:"column:likes; serializer:json" json:"likes"`  // 点赞数组
	Forwards     int                   `gorm:"column:forwards" json:"forwards"`             // 转发数
	ReviewPassAt int64                 `gorm:"column:review_pass_at" json:"review_pass_at"` // 审核通过毫秒时间戳

	*commonpb.User `gorm:"-" json:"-"`
}

func (*Moment) TableName() string {
	return "moment"
}

func (t *Moment) ToPB(comments int64) *momentpb.Moment {
	pb := &momentpb.Moment{
		Mid:          t.Id,
		Uid:          t.UID,
		Text:         t.Text,
		Type:         t.Type,
		ReviewStatus: t.ReviewStatus,
		MediaUrls:    t.MediaUrls,
		CreatedAt:    t.CreatedAt.Unix(),
		Likes:        int64(len(t.Likes)),
		Comments:     comments,
		Forwards:     int64(t.Forwards),
	}
	return pb
}

func (t *Moment) GetUIDs() []int64 {
	return []int64{t.UID}
}
func (t *Moment) SetUser(user ...*commonpb.User) {
	t.User = user[0]
}

type MomentComment struct {
	model.FieldID
	model.FieldUID
	model.FieldCt
	Mid      int64  `gorm:"column:mid" json:"mid"`
	ReplyUID int64  `gorm:"column:reply_uid" json:"reply_uid"`
	Content  string `gorm:"column:content" json:"content"`

	User      *commonpb.User `gorm:"-" json:"-"`
	ReplyUser *commonpb.User `gorm:"-" json:"-"`
}

func (*MomentComment) TableName() string {
	return "moment_comment"
}
func (t *MomentComment) ToPB() *momentpb.Comment {
	return &momentpb.Comment{
		ReplyUid:  t.ReplyUID,
		Mid:       t.Mid,
		Uid:       t.UID,
		Content:   t.Content,
		CreatedAt: t.CreatedAt.Unix(),
	}
}

func (t *MomentComment) GetUIDs() []int64 {
	return []int64{t.UID, t.ReplyUID}
}
func (t *MomentComment) SetUser(user ...*commonpb.User) {
	t.User = user[0]
	t.ReplyUser = user[1]
}
