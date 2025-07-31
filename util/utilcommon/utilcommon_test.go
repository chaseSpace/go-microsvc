package utilcommon

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrintlnStackErr(t *testing.T) {
	PrintlnStackMsg("test")
}

type Temp int

func (Temp) Run() string {
	return CurrFuncName(1)
}

func TestCurrFuncName(t *testing.T) {
	assert.Equal(t, "TestCurrFuncName", CurrFuncName(1))
	assert.Equal(t, "Temp.Run", Temp(1).Run())
}
