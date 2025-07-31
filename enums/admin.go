package enums

type ReviewTextTyp uint8

const (
	ReviewTextTypeNickname    ReviewTextTyp = iota
	ReviewTextTypeDescription               // 签名
)

type ReviewTextStatus uint8

const (
	ReviewTextStatusWaiting           ReviewTextStatus = iota
	ReviewTextStatusRejected                           // 拒绝
	ReviewTextStatusApprovedByMachine                  // 机审通过
	ReviewTextStatusApprovedByAdmin                    // 人审通过
)

type ReviewImgTyp uint8

const (
	ReviewImgTypeIcon  ReviewTextTyp = iota // 头像
	ReviewImgTypeAlbum                      // 相册
)

type ReviewImgStatus uint8

const (
	ReviewImgStatusWaiting           ReviewTextStatus = iota
	ReviewImgStatusRejected                           // 拒绝
	ReviewImgStatusApprovedByMachine                  // 机审通过
	ReviewImgStatusApprovedByAdmin                    // 人审通过
)
