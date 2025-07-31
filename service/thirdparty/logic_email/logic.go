package logic_email

import (
	"context"
	"microsvc/protocol/svc/thirdpartypb"
)

type ctrl struct {
}

var Ext ctrl

func (c ctrl) VerifyEmailCode(ctx context.Context, req *thirdpartypb.VerifyEmailCodeReq) (*thirdpartypb.VerifyEmailCodeRes, error) {
	match, err := __verifyEmailCode(ctx, &__verifyEmailCodeReq{
		Email:             req.InputEmail,
		Code:              req.InputCode,
		Scene:             req.Scene,
		DeleteAfterVerify: req.DeleteAfterVerify,
	})
	return &thirdpartypb.VerifyEmailCodeRes{IsMatch: match}, err
}
