package user

import (
	"errors"
	"fmt"
	"microsvc/bizcomm/comminfra"
	"microsvc/enums"
	"microsvc/infra/svccli/rpcext"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/userpb"
	deploy2 "microsvc/service/user/deploy"
	"microsvc/test/tbase"
	"microsvc/util/ucrypto"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/k0kubun/pp/v3"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
)

/*
测试RPC接口，需要先在本地启动待测试的微服务（可不用启动网关）
*/

func TestHealthCheck(t *testing.T) {
	tbase.GRPCHealthCheck(t, enums.SvcUser, deploy2.UserConf)
}

func TestBatchSignUpAll_WxMini(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	req := &userpb.SignUpAllReq{
		Base: tbase.TestBaseExtReq,
		Type: commonpb.SignInType_SIT_WX_MINI,
		Code: "xxx",
		Body: &userpb.SignUpBody{
			Nickname: fmt.Sprintf("user"), // 昵称限长10字符
			Sex:      enums.SexMale.ToPB(),
			Birthday: "2023-01-01",
			Extra:    &userpb.SignUpExt{Channel: "official"},
		},
	}
	_, err := rpcext.User().SignUpAll(tbase.TestCallCtx, req)
	if err != nil {
		assert.FailNow(t, "SignUpAll", err.Error())
	}
}

func TestBatchSignUpAll_Phone(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	phone := 18855557777
	for i := 0; i < 1; i++ {
		req := &userpb.SignUpAllReq{
			Base:       tbase.TestBaseExtReq,
			Type:       commonpb.SignInType_SIT_PHONE,
			AnyAccount: "86|" + cast.ToString(phone+i),
			VerifyCode: "1234",
			Body: &userpb.SignUpBody{
				Nickname: fmt.Sprintf("user%d", i), // 昵称限长10字符
				Sex:      enums.SexMale.ToPB(),
				Birthday: "2023-01-01",
				Extra:    &userpb.SignUpExt{Channel: "official"},
			},
		}
		if i%2 == 0 {
			req.Body.Sex = enums.SexFemale.ToPB()
		}
		_, err := rpcext.User().SignUpAll(tbase.TestCallCtx, req)
		if err != nil {
			assert.FailNow(t, "SignUpAll", err.Error())
		}
	}
}

func TestBatchSignUpAll_Email(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	start := 0
	for i := 0; i < 1; i++ {
		req := &userpb.SignUpAllReq{
			Base:       tbase.TestBaseExtReq,
			Type:       commonpb.SignInType_SIT_EMAIL,
			AnyAccount: cast.ToString(start+i) + "@qq.com",
			VerifyCode: "1234",
			Password:   ucrypto.MustSha1("1234"),
			Body: &userpb.SignUpBody{
				Nickname:  fmt.Sprintf("user%d", i), // 昵称限长10字符
				Firstname: "",
				Lastname:  "",
				Sex:       enums.SexNoSet.ToPB(),
				Birthday:  "2000-01-01",
				Extra:     &userpb.SignUpExt{Channel: "official"},
			},
		}
		if i%2 == 0 {
			req.Body.Sex = enums.SexFemale.ToPB()
		}
		_, err := rpcext.User().SignUpAll(tbase.TestCallCtx, req)
		if err != nil {
			assert.FailNow(t, "SignUpAll", err.Error())
		}
	}
}

func TestConcurrencySignUp(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	x := sync.WaitGroup{}

	uidRepeatedErrCnt := atomic.Int32{}

	expectedErr := xerr.ErrInternal.New("太多人注册辣，隔几秒再试一下哦")
	// 并发注册

	// 业务代码中设置的号池容量为100，这里设置相同的并发参数100，则完全可以处理，不会出现重复ID
	// - 对于更高的并发，虽然也不会出现重复id，但会增加接口耗时，同时建议提高 号池容量 配置以提高并发性能
	total := 100
	phone := 18855560818
	for i := 0; i < total; i++ {
		req := &userpb.SignUpAllReq{
			Type:       commonpb.SignInType_SIT_PHONE,
			AnyAccount: "86|" + cast.ToString(phone+i),
			Body: &userpb.SignUpBody{
				Nickname: fmt.Sprintf("user%d", i), // 昵称限长10字符
				Sex:      enums.SexMale.ToPB(),
				Birthday: "2023-01-01",
			},
		}
		if i%2 == 0 {
			req.Body.Sex = enums.SexFemale.ToPB()
		}
		x.Add(1)
		go func() {
			_, err := rpcext.User().SignUpAll(tbase.TestCallCtxNoAuth, req)
			if err != nil {
				assert.Equal(t, expectedErr, err)
				uidRepeatedErrCnt.Add(1)
			}
			x.Done()
		}()
	}

	x.Wait()

	errCnt := uidRepeatedErrCnt.Load()
	if errCnt > 0 {
		t.Errorf("并发次数：%d, 失败次数:%d 超出预期\n", total, errCnt)
	} else {
		t.Logf("并发次数：%d, 失败次数:%d 符合预期\n", total, errCnt)
	}
}

func TestSignInAll(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	type item struct {
		title   string
		req     *userpb.SignInAllReq
		wantErr error
	}

	// todo 待添加用例
	var tt = []*item{
		//{
		//	title: "by phone",
		//	req: &userpb.SignInAllReq{
		//		Base:       tbase.TestBaseExtReq,
		//		Type:       commonpb.SignInType_SIT_PHONE,
		//		AnyAccount: "86|" + cast.ToString(18855557777),
		//		VerifyCode: "1234",
		//	},
		//	wantErr: nil,
		//},
		{
			title: "by email",
			req: &userpb.SignInAllReq{
				Base:       tbase.TestBaseExtReq,
				Type:       commonpb.SignInType_SIT_EMAIL,
				AnyAccount: "0@qq.com",
				VerifyCode: "1234",
				Password:   ucrypto.MustSha1("1234"),
			},
			wantErr: nil,
		},
	}

	for _, v := range tt {
		r, err := rpcext.User().SignInAll(tbase.TestCallCtx, v.req)
		if !errors.Is(err, v.wantErr) {
			t.Fatalf("title:%v err:%v wantErr:%v", v.title, err, v.wantErr)
		}
		if err == nil {
			assert.NotEmpty(t, r.Token)
		}
	}
}

func TestGetUserInfo(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	res, err := rpcext.User().GetUserInfo(tbase.TestCallCtx, &userpb.GetUserInfoReq{
		Base:      tbase.TestBaseExtReq,
		Uids:      []int64{100016},
		GetCaller: true,
		//PopulateNotFound: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	pp.Println(res.Umap)
	pp.Println("Caller:", res.Caller)
}

func TestAPIRateLimit(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	var wg sync.WaitGroup
	var nilErrs int
	var tooManyReqErrs int

	N := comminfra.DefaultSingleAPIMaxQPSByUID + 1
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := rpcext.User().GetUserInfo(tbase.TestCallCtx, &userpb.GetUserInfoReq{
				Base: tbase.TestBaseExtReq,
				Uids: []int64{1},
			})
			if err == nil {
				nilErrs++
			} else if errors.Is(err, xerr.ErrTooManyRequests) {
				tooManyReqErrs++
			}
		}(i)
	}
	wg.Wait()
	assert.Equal(t, N-1, nilErrs, "nilErrs count")
	assert.Equal(t, 1, tooManyReqErrs, "tooManyReqErrs count")
}

func TestResetPassword(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	rpcext.User().ResetPassword(tbase.TestCallCtx, &userpb.ResetPasswordReq{
		Base:        tbase.TestBaseExtReq,
		Email:       "0@qq.com",
		VerifyCode:  "271001",
		NewPassword: ucrypto.MustSha1("1234"),
	})
}

func Test_UpdateUserInfo_Passwd(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	rpcext.User().UpdateUserInfo(tbase.TestCallCtx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Password,
				AnyValue:  ucrypto.MustSha1("1234") + "|" + ucrypto.MustSha1("1235"),
			},
		},
	})
}

func Test_UpdateUserInfo_Avatar(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	rpcext.User().UpdateUserInfo(tbase.TestCallCtx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Avatar,
				AnyValue:  "x",
			},
		},
	})
}
