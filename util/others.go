package util

import (
	"math"
	"unicode"
)

func SplitRule(c rune) bool {
	if c == '\t' {
		return true
	} else {
		return false
	}
}

func SplitRuleSpace(c rune) bool {
	if c == '\t' || c == ' ' {
		return true
	} else {
		return false
	}
}

func SplitRuneAt(c rune) bool {
	if c == '@' {
		return true
	} else {
		return false
	}
}

func SplitRuneUnderline(c rune) bool {
	if c == '_' {
		return true
	} else {
		return false
	}
}

type funcSplitRule func(c rune) bool

func GetSpliteRule(sc rune) funcSplitRule {
	return func(c rune) bool {
		if c == sc {
			return true
		} else {
			return false
		}
	}
}

func Max(x, y float32) float32 {
	return float32(math.Max(float64(x), float64(y)))
}

func Min(x, y float32) float32 {
	return float32(math.Min(float64(x), float64(y)))
}

func MinUInt32(x, y uint32) uint32 {
	if x > y {
		return y
	}
	return x
}

func GetSequence() func() int {
	i := 0
	return func() int {
		i += 1
		return i
	}
}

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

func MaxNumber[T Number](i, j T) T {
	if i > j {
		return i
	}
	return j
}

func MinNumber[T Number](i, j T) T {
	if i < j {
		return i
	}
	return j
}

func Title(src string) string {
	dest := []rune{}
	c := true
	for _, v := range src {
		if v == '_' {
			c = true
			continue
		}
		if c {
			v = unicode.ToUpper(v)
			c = false
		}
		dest = append(dest, v)
	}
	return string(dest)
}
