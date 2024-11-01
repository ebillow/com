package util

func AbsInt32(a int32) int32 {
	if a < 0 {
		a = -a
	}
	return a
}
