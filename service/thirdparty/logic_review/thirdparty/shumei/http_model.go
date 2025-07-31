package shumei

import (
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/service/thirdparty/logic_review/thirdparty"
	"microsvc/util/ucrypto"

	"github.com/spf13/cast"
)

/* 文本审核model */

type TextPayload struct {
	AccessKey string `json:"accessKey"`
	AppID     string `json:"appId"`
	EventID   string `json:"eventId"` // 联系数美服务开通
	Type      string `json:"type"`
	Data      struct {
		Text    string `json:"text"`
		TokenID string `json:"tokenId"`
	} `json:"data"`
}

type TextResponse struct {
	Code            int        `json:"code"`            // 返回码 1100：成功
	Message         string     `json:"message"`         // 返回码描述
	RequestId       string     `json:"requestId"`       // 请求标识
	RiskLevel       string     `json:"riskLevel"`       // 处置建议
	RiskLabel1      string     `json:"riskLabel1"`      // 一级风险标签
	RiskLabel2      string     `json:"riskLabel2"`      // 二级风险标签
	RiskLabel3      string     `json:"riskLabel3"`      // 三级风险标签
	RiskDescription string     `json:"riskDescription"` // 风险原因
	RiskDetail      RiskDetail `json:"riskDetail"`      // 风险详情
	FinalResult     int        `json:"finalResult"`     // 是否最终结果  1:是，0：否(说明该结果为数美风控的过程结果，还经过数美人审再次check后回传贵司)
	ResultType      int        `json:"resultType"`      // 0:机审，1:人审
	_AIReviewStatus commonpb.AIReviewStatus
}

func (r TextResponse) isOK() bool {
	r._AIReviewStatus = map[string]commonpb.AIReviewStatus{
		"PASS":   commonpb.AIReviewStatus_ARS_Pass,
		"REVIEW": commonpb.AIReviewStatus_ARS_Review,
		"REJECT": commonpb.AIReviewStatus_ARS_Reject,
	}[r.RiskLevel]
	return r.Code == 1100 && r.FinalResult == 1 && r._AIReviewStatus > 0
}
func (r TextResponse) ToResult() (*thirdparty.ReviewResult, error) {
	if !r.isOK() {
		return nil, xerr.ErrThirdParty.New("shumei: %s", r.Message)
	}
	return &thirdparty.ReviewResult{
		ReqId:           r.RequestId,
		RiskDescription: r.RiskDescription,
		RiskLabel:       r.RiskLabel1,
		Status:          r._AIReviewStatus,
	}, nil
}

type RiskDetail struct {
	MatchedLists []MatchedList `json:"matchedLists"` // 命中的自定义名单列表
	RiskSegments []RiskSegment `json:"riskSegments"` // 高风险内容片段
}

type MatchedList struct {
	Name  string `json:"name"`  // 命中的名单名称
	Words []Word `json:"words"` // 命中的敏感词数组
}

type Word struct {
	Word     string `json:"word"`     // 命中的敏感词
	Position []int  `json:"position"` // 敏感词所在位置
}

type RiskSegment struct {
	Segment  string `json:"segment"`  // 高风险内容片段
	Position int    `json:"position"` // 	高风险内容片段所在位置
}

// 系统文本类型 对应 数美文本检测策略
var textTypPolicy = map[thirdpartypb.TextType]string{}

func (s *Shumei) buildTextPayload(uid int64, text string, typ thirdpartypb.TextType) *TextPayload {
	cryptUID, _ := ucrypto.NewCryptoAes(s.config.UIDCryptoKey).Encrypt([]byte(cast.ToString(uid)))
	policy := textTypPolicy[typ]
	if policy == "" {
		policy = "TEXTRISK_FRUAD_TEXTMINOR"
	}
	return &TextPayload{
		AccessKey: s.config.AccessKey,
		AppID:     s.config.Appid,
		EventID:   "", // 联系数美服务开通
		Type:      policy,
		Data: struct {
			Text    string `json:"text"`
			TokenID string `json:"tokenId"`
		}{
			Text:    text,
			TokenID: cryptUID, // UID 加密
		},
	}
}

/* 图片审核model */

type ImagePayload struct {
	AccessKey string `json:"accessKey"`
	AppID     string `json:"appId"`
	EventID   string `json:"eventId"` // 联系数美服务开通
	Type      string `json:"type"`
	Data      struct {
		Img     string `json:"img"`
		TokenID string `json:"tokenId"`
		Lang    string `json:"lang"`
	} `json:"data"`
}

type ImgRiskDetail struct {
	Faces      []Face  `json:"faces"`
	OcrText    OcrText `json:"ocrText"`
	RiskSource int     `json:"riskSource"`
}

type Face struct {
	FaceRatio   float64 `json:"face_ratio"`
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Probability float64 `json:"probability"`
}

type OcrText struct {
	Text string `json:"text"`
}

type Label struct {
	Probability     float64       `json:"probability"`
	RiskDescription string        `json:"riskDescription"`
	RiskDetail      ImgRiskDetail `json:"riskDetail"`
	RiskLabel1      string        `json:"riskLabel1"`
	RiskLabel2      string        `json:"riskLabel2"`
	RiskLabel3      string        `json:"riskLabel3"`
	RiskLevel       string        `json:"riskLevel"`
}

type ImageResponse struct {
	AllLabels       []Label       `json:"allLabels"`
	Code            int           `json:"code"`
	FinalResult     int           `json:"finalResult"`
	Message         string        `json:"message"`
	RequestId       string        `json:"requestId"`
	ResultType      int           `json:"resultType"`
	RiskDescription string        `json:"riskDescription"`
	RiskDetail      ImgRiskDetail `json:"riskDetail"`
	RiskLabel1      string        `json:"riskLabel1"`
	RiskLabel2      string        `json:"riskLabel2"`
	RiskLabel3      string        `json:"riskLabel3"`
	RiskLevel       string        `json:"riskLevel"`
	_AIReviewStatus commonpb.AIReviewStatus
}

func (r ImageResponse) isOK() bool {
	r._AIReviewStatus = map[string]commonpb.AIReviewStatus{
		"PASS":   commonpb.AIReviewStatus_ARS_Pass,
		"REVIEW": commonpb.AIReviewStatus_ARS_Review,
		"REJECT": commonpb.AIReviewStatus_ARS_Reject,
	}[r.RiskLevel]
	return r.Code == 1100 && r.FinalResult == 1 && r._AIReviewStatus > 0
}
func (r ImageResponse) ToResult() (*thirdparty.ReviewResult, error) {
	if !r.isOK() {
		return nil, xerr.ErrThirdParty.New("shumei: %s", r.Message)
	}
	return &thirdparty.ReviewResult{
		ReqId:           r.RequestId,
		RiskDescription: r.RiskDescription,
		RiskLabel:       r.RiskLabel1,
		Status:          r._AIReviewStatus,
	}, nil
}

// 系统图片类型 对应 数美图片检测策略
var imageTypPolicy = map[thirdpartypb.ImageType]string{}

func (s *Shumei) buildImagePayload(uid int64, uri string, typ thirdpartypb.ImageType) *ImagePayload {
	cryptUID, _ := ucrypto.NewCryptoAes(s.config.UIDCryptoKey).Encrypt([]byte(cast.ToString(uid)))
	policy := imageTypPolicy[typ]
	if policy == "" {
		policy = "POLITY_EROTIC_VIOLENT_QRCODE_ADVERT_IMGTEXTRISK"
	}
	return &ImagePayload{
		AccessKey: s.config.AccessKey,
		AppID:     s.config.Appid,
		EventID:   "", // 联系数美服务开通
		Type:      policy,
		Data: struct {
			Img     string `json:"img"`
			TokenID string `json:"tokenId"`
			Lang    string `json:"lang"`
		}{
			Img:     uri,
			TokenID: cryptUID,
			Lang:    "zh", // 当type中包含`IMGTEXTRISK`时，检测图片文字；zh-中文，en-英文
		},
	}
}

/* 视频审核model */

type VideoPayload struct {
	AccessKey string `json:"accessKey"`
	AppID     string `json:"appId"`
	//AudioBusinessType string `json:"audioBusinessType"`
	AudioType string `json:"audioType"`
	Callback  string `json:"callback"`
	Data      Data   `json:"data"`
	EventId   string `json:"eventId"`
	//ImgBusinessType   string `json:"imgBusinessType"`
	ImgType string `json:"imgType"`
}

type AdvancedFrequency struct {
	DurationPoints []int `json:"durationPoints"`
	Frequencies    []int `json:"frequencies"`
}

type Data struct {
	AdvancedFrequency AdvancedFrequency `json:"advancedFrequency"` // 高级截帧策略
	BtId              string            `json:"btId"`              // 客户侧请求唯一标识
	DetectFrequency   int               `json:"detectFrequency"`
	Ip                string            `json:"ip"`
	ReturnAllAudio    int               `json:"returnAllAudio"`
	ReturnAllImg      int               `json:"returnAllImg"`
	TokenId           string            `json:"tokenId"`    // UID 加密
	Url               string            `json:"url"`        // 视频URL
	DataId            string            `json:"dataId"`     // 客户自定义数据Id。可以用于数美saas后台检索
	Lang              string            `json:"lang"`       // 语言类型 zh | en
	VideoTitle        string            `json:"videoTitle"` // 视频标题，用于数美后台界面展示
}

type VideoResponse struct {
	BtId      string `json:"btId"` // 我方的唯一请求ID
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"requestId"`
}

func (r *VideoResponse) ToResult(name string) (*thirdparty.AsyncReviewResult, error) {
	if r.Code != 1100 || r.BtId == "" {
		return nil, xerr.ErrThirdParty.New("shumei: %s", r.Message)
	}
	return &thirdparty.AsyncReviewResult{
		ReqId:             r.BtId,
		ThirdPartySvcName: name,
	}, nil
}

// VideoResultResp --
type VideoResultResp struct {
	AudioDetail     []AudioDetail `json:"audioDetail"`
	AuxInfo         AuxInfo       `json:"auxInfo"`
	BtId            string        `json:"btId"`
	Code            int           `json:"code"`
	FrameDetail     []FrameDetail `json:"frameDetail"`
	Message         string        `json:"message"`
	RequestId       string        `json:"requestId"`
	RiskLevel       string        `json:"riskLevel"`
	_AIReviewStatus commonpb.AIReviewStatus
}

func (r VideoResultResp) isOK() bool {
	r._AIReviewStatus = map[string]commonpb.AIReviewStatus{
		"":       commonpb.AIReviewStatus_ARS_Pending,
		"PASS":   commonpb.AIReviewStatus_ARS_Pass,
		"REVIEW": commonpb.AIReviewStatus_ARS_Review,
		"REJECT": commonpb.AIReviewStatus_ARS_Reject,
	}[r.RiskLevel]
	return r.Code == 1100 || r.Code == 1101
}
func (r VideoResultResp) ToResult() (*thirdparty.ReviewResult, error) {
	if !r.isOK() {
		return nil, xerr.ErrThirdParty.New("shumei: %s", r.Message)
	}
	rr := &thirdparty.ReviewResult{
		ReqId:     r.RequestId,
		Status:    r._AIReviewStatus,
		RiskLabel: r.RiskLevel,
	}
	// 从违规的视频帧 或 音频中 提取 RiskDescription
	if len(r.FrameDetail) > 0 && len(r.FrameDetail[0].AllLabels) > 0 {
		rr.RiskDescription = r.FrameDetail[0].AllLabels[0].RiskDescription
	} else if len(r.AudioDetail) > 0 && len(r.AudioDetail[0].AllLabels) > 0 {
		rr.RiskDescription = r.AudioDetail[0].AllLabels[0].RiskDescription
	} else {
		return nil, xerr.ErrThirdParty.New("shumei-video-callback: no labels found")
	}
	return rr, nil
}

type VideoRiskDetail struct {
	AudioText  string `json:"audioText"`
	RiskSource int    `json:"riskSource"`
	OcrText    struct {
		Text string `json:"text"`
	} `json:"ocrText"`
}

type VideoLabel struct {
	Probability     float64         `json:"probability"`
	RiskDescription string          `json:"riskDescription"`
	RiskDetail      VideoRiskDetail `json:"riskDetail"`
	RiskLabel1      string          `json:"riskLabel1"`
	RiskLabel2      string          `json:"riskLabel2"`
	RiskLabel3      string          `json:"riskLabel3"`
	RiskLevel       string          `json:"riskLevel"`
}

type AudioDetail struct {
	AllLabels      []VideoLabel    `json:"allLabels"`
	AudioEndtime   int             `json:"audioEndtime"`
	AudioStarttime int             `json:"audioStarttime"`
	AudioText      string          `json:"audioText"`
	AudioUrl       string          `json:"audioUrl"`
	BusinessLabels []VideoLabel    `json:"businessLabels"`
	RequestId      string          `json:"requestId"`
	RiskDetail     VideoRiskDetail `json:"riskDetail"`
	RiskLabel1     string          `json:"riskLabel1"`
	RiskLabel2     string          `json:"riskLabel2"`
	RiskLabel3     string          `json:"riskLabel3"`
	RiskLevel      string          `json:"riskLevel"`
}

type AuxInfo struct {
	BillingAudioDuration int `json:"billingAudioDuration"`
	BillingImgNum        int `json:"billingImgNum"`
	FrameCount           int `json:"frameCount"`
	Time                 int `json:"time"`
}

type FrameDetail struct {
	AllLabels  []VideoLabel    `json:"allLabels"`
	AuxInfo    AuxInfo         `json:"auxInfo"`
	ImgText    string          `json:"imgText"`
	ImgUrl     string          `json:"imgUrl"`
	RequestId  string          `json:"requestId"`
	RiskDetail VideoRiskDetail `json:"riskDetail"`
	RiskLabel1 string          `json:"riskLabel1"`
	RiskLabel2 string          `json:"riskLabel2"`
	RiskLabel3 string          `json:"riskLabel3"`
	RiskLevel  string          `json:"riskLevel"`
	Time       int             `json:"time"`
}

// 系统图片类型 对应 数美图片检测策略
// - Value: [imgTyp, audioTyp]
var videoTypPolicy = map[thirdpartypb.VideoType][]string{}

func (s *Shumei) buildVideoPayload(uid int64, uri string, typ thirdpartypb.VideoType, reqID string) *VideoPayload {
	cryptUID, _ := ucrypto.NewCryptoAes(s.config.UIDCryptoKey).Encrypt([]byte(cast.ToString(uid)))
	policy := videoTypPolicy[typ]
	if policy == nil {
		policy = []string{"POLITY_EROTIC_VIOLENT_QRCODE_ADVERT_IMGTEXTRISK", "POLITY_EROTIC_ADVERT_DIRTY_ADLAW_MOAN_AUDIOPOLITICAL_ANTHEN"}
	}
	return &VideoPayload{
		AccessKey: s.config.AccessKey,
		AppID:     s.config.Appid,
		ImgType:   policy[0],
		AudioType: policy[1],
		//Callback:  "", // 不走回调，而是定时任务主动轮询
		Data: Data{
			AdvancedFrequency: AdvancedFrequency{
				DurationPoints: []int{},
				Frequencies:    []int{},
			},
			BtId: reqID, // 唯一请求ID，暂时无用
			//DetectFrequency: 0,
			//Ip:              "",
			ReturnAllAudio: 0, // 返回非pass的音频
			ReturnAllImg:   0, // 返回非pass的图片
			TokenId:        cryptUID,
			Url:            uri,
			DataId:         reqID,
			Lang:           "zh",
			VideoTitle:     typ.String(),
		},
		EventId: "", // 联系数美获取
	}
}

/* 视频审核结果查询 model */

type VideoQueryPayload struct {
	AccessKey string `json:"accessKey"`
	BtId      string `json:"btId"`
}
