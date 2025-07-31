package commthirdparty

import (
	"microsvc/protocol/svc/commonpb"
)

type ReviewResult interface {
	GetStatus() commonpb.AIReviewStatus
	GetMessage() string
}
