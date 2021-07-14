package eventengine

import (
	"testing"

	gutils "github.com/Laisky/go-utils"
	"github.com/stretchr/testify/require"
)

func TestGetFuncAddress(t *testing.T) {
	var foo1 = func() {}
	var foo2 = func() {}

	a1 := GetFuncAddress(foo1)
	a12 := GetFuncAddress(foo1)
	require.Equal(t, a1, a12)

	a2 := GetFuncAddress(foo2)
	a22 := GetFuncAddress(foo2)
	require.Equal(t, a2, a22)

	require.NotEqual(t, a1, a2)
	require.NotEqual(t, a1, a22)
	require.NotEqual(t, a12, a2)
	require.NotEqual(t, a12, a22)

	ok := gutils.IsPanic(func() {
		GetFuncAddress(123)
	})
	require.True(t, ok)
}
