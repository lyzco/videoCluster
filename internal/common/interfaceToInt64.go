package common

import (
	"fmt"
	"strconv"
)

type IntLike interface {
	~int | ~int32 | ~int64 | ~float32 | ~float64 | ~string
}

func ConvertToInt64[T IntLike](val T) (int64, error) {
	switch v := any(val).(type) {
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported type: %T", v)
	}
}
