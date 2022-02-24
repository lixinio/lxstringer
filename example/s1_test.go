package example

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestS11(t *testing.T) {
	require.Equal(t, S11_1.Code(), "A A")
	require.Equal(t, S11_2.Code(), "FD SAF")
	require.Equal(t, S11_3.Code(), "F发 生")
	require.Equal(t, S11_4.Code(), "D")
	require.Equal(t, S11_5.Code(), "D")

	require.Equal(t, S11_1.Name(), "aaa")
	require.Equal(t, S11_2.Name(), "bbb")
	require.Equal(t, S11_3.Name(), "ccc")
	require.Equal(t, S11_4.Name(), "DD")
	require.Equal(t, S11_5.Name(), "DD")

	require.Equal(t, CodeToS11("A A", S11_1), S11_1)
	require.Equal(t, CodeToS11("FD SAF", S11_1), S11_2)
	require.Equal(t, CodeToS11("F发 生", S11_1), S11_3)
	require.Equal(t, CodeToS11("D", S11_1), S11_4)
}
