package example

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestS21(t *testing.T) {
	require.Equal(t, S21_1.Code(), "A A")
	require.Equal(t, S21_2.Code(), "FD SAF")
	require.Equal(t, S21_3.Code(), "F发 生")

	require.Equal(t, S21_1.Name(), "aaa")
	require.Equal(t, S21_2.Name(), "bbb")
	require.Equal(t, S21_3.Name(), "ccc")

	require.Equal(t, CodeToS21("A A", S21_1), S21_1)
	require.Equal(t, CodeToS21("FD SAF", S21_1), S21_2)
	require.Equal(t, CodeToS21("F发 生", S21_1), S21_3)
}

func TestS22(t *testing.T) {
	require.Equal(t, S22_1.Code(), "A b C")
	require.Equal(t, S22_2.Code(), "中 华")
	require.Equal(t, S22_3.Code(), "啊`啊")

	require.Equal(t, S22_1.Name(), "d E f")
	require.Equal(t, S22_2.Name(), "人 们")
	require.Equal(t, S22_3.Name(), "i'm ok")

	require.Equal(t, CodeToS22("A b C", S22_1), S22_1)
	require.Equal(t, CodeToS22("中 华", S22_1), S22_2)
	require.Equal(t, CodeToS22("啊`啊", S22_1), S22_3)
}
