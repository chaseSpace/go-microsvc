package ufile

import (
	"testing"
)

func Test_CompressImageFile(t *testing.T) {
	buf := MustRead("../../test/testdata/baidu.png")
	buf2, compressed, err := CompressImageFile(buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(compressed, len(buf2), len(buf))
}
