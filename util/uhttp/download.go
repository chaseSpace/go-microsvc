package uhttp

import (
	"io"
	"microsvc/pkg/xerr"
	"net/http"
)

func DownFile(url string) (buf []byte, err error) {
	r, err := http.Get(url)
	if err != nil {
		return
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		return nil, xerr.ErrThirdParty.New("DownFile fail, code(%d)", r.StatusCode)
	}
	return io.ReadAll(r.Body)
}
