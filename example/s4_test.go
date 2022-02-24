package example

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestS41(t *testing.T) {
	require.Equal(t, S41_1.CodeName(), "A b C")
	require.Equal(t, S41_2.CodeName(), "中 华")
	require.Equal(t, S41_3.CodeName(), "啊`啊")

	require.Equal(t, S41_1.Name2(), "d E f")
	require.Equal(t, S41_2.Name2(), "人 们")
	require.Equal(t, S41_3.Name2(), "i'm ok")

	require.Equal(t, S41FromCode("A b C", S41_1), S41_1)
	require.Equal(t, S41FromCode("中 华", S41_1), S41_2)
	require.Equal(t, S41FromCode("啊`啊", S41_1), S41_3)
}
