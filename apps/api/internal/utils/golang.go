package utils

import (
	"fmt"
	"strconv"
)

func ConvertAnyToInt64(t any) (int64, error) {
	switch t := t.(type) {
	case int64:
		return t, nil
	case int:
		return int64(t), nil
	case string:
		return strconv.ParseInt(t, 10, 64)
	default:
		return 0, fmt.Errorf("type %T not supported", t)
	}
}
