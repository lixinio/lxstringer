package example

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestS31(t *testing.T) {
	require.Equal(t, S31_1.Code(), "A b C")
	require.Equal(t, S31_2.Code(), "中 华")
	require.Equal(t, S31_3.Code(), "啊`啊")

	require.Equal(t, S31_1.Name(), "d E f")
	require.Equal(t, S31_2.Name(), "人 们")
	require.Equal(t, S31_3.Name(), "i'm ok")

	require.Equal(t, CodeToS31("A b C", S31_1), S31_1)
	require.Equal(t, CodeToS31("中 华", S31_1), S31_2)
	require.Equal(t, CodeToS31("啊`啊", S31_1), S31_3)
}

func TestS32(t *testing.T) {
	require.Equal(t, S32_1.Code(), "A b C")
	require.Equal(t, S32_2.Code(), "中 华")
	require.Equal(t, S32_3.Code(), "啊`啊")

	require.Equal(t, S32_1.Name(), "d E f")
	require.Equal(t, S32_2.Name(), "人 们")
	require.Equal(t, S32_3.Name(), "i'm ok")

	require.Equal(t, CodeToS32("A b C", S32_1), S32_1)
	require.Equal(t, CodeToS32("中 华", S32_1), S32_2)
	require.Equal(t, CodeToS32("啊`啊", S32_1), S32_3)
}

func TestS33(t *testing.T) {
	require.Equal(t, S33_1.Code(), "A b C1")
	require.Equal(t, S33_2.Code(), "中 华1")
	require.Equal(t, S33_3.Code(), "啊`啊1")
	require.Equal(t, S33_4.Code(), "A b C2")
	require.Equal(t, S33_5.Code(), "中 华2")
	require.Equal(t, S33_6.Code(), "啊`啊2")
	require.Equal(t, S33_7.Code(), "A b C3")
	require.Equal(t, S33_8.Code(), "中 华3")
	require.Equal(t, S33_9.Code(), "啊`啊3")
	require.Equal(t, S33_10.Code(), "A b C4")
	require.Equal(t, S33_11.Code(), "中 华4")
	require.Equal(t, S33_12.Code(), "啊`啊4")

	require.Equal(t, S33_1.Name(), "d E f")
	require.Equal(t, S33_2.Name(), "人 们")
	require.Equal(t, S33_3.Name(), "i'm ok")
	require.Equal(t, S33_4.Name(), "d E f")
	require.Equal(t, S33_5.Name(), "人 们")
	require.Equal(t, S33_6.Name(), "i'm ok")
	require.Equal(t, S33_7.Name(), "d E f")
	require.Equal(t, S33_8.Name(), "人 们")
	require.Equal(t, S33_9.Name(), "i'm ok")
	require.Equal(t, S33_10.Name(), "d E f")
	require.Equal(t, S33_11.Name(), "人 们")
	require.Equal(t, S33_12.Name(), "i'm ok")

	require.Equal(t, CodeToS33("A b C1", S33_1), S33_1)
	require.Equal(t, CodeToS33("中 华1", S33_1), S33_2)
	require.Equal(t, CodeToS33("啊`啊1", S33_1), S33_3)
	require.Equal(t, CodeToS33("A b C2", S33_1), S33_4)
	require.Equal(t, CodeToS33("中 华2", S33_1), S33_5)
	require.Equal(t, CodeToS33("啊`啊2", S33_1), S33_6)
	require.Equal(t, CodeToS33("A b C3", S33_1), S33_7)
	require.Equal(t, CodeToS33("中 华3", S33_1), S33_8)
	require.Equal(t, CodeToS33("啊`啊3", S33_1), S33_9)
	require.Equal(t, CodeToS33("A b C4", S33_1), S33_10)
	require.Equal(t, CodeToS33("中 华4", S33_1), S33_11)
	require.Equal(t, CodeToS33("啊`啊4", S33_1), S33_12)
}
