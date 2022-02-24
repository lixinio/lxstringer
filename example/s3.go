package example

type S31 int
type S32 int
type S33 int

const (
	S31_1 S31 = 2 * iota // "A b C" "d E f" "G h I"
	S31_2                // "中 华" "人 们"
	S31_3                // "啊`啊" "i'm ok"
)

const (
	S32_1 S32 = 2*iota + 100 // "A b C" "d E f" "G h I"
	S32_2                    // "中 华" "人 们"
	S32_3                    // "啊`啊" "i'm ok"
)

const (
	S33_1  S33 = 1<<iota + iota // "A b C1" "d E f" "G h I"
	S33_2                       // "中 华1" "人 们"
	S33_3                       // "啊`啊1" "i'm ok"
	S33_4                       // "A b C2" "d E f" "G h I"
	S33_5                       // "中 华2" "人 们"
	S33_6                       // "啊`啊2" "i'm ok"
	S33_7                       // "A b C3" "d E f" "G h I"
	S33_8                       // "中 华3" "人 们"
	S33_9                       // "啊`啊3" "i'm ok"
	S33_10                      // "A b C4" "d E f" "G h I"
	S33_11                      // "中 华4" "人 们"
	S33_12                      // "啊`啊4" "i'm ok"
)
