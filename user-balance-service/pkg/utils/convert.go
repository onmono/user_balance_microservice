package convert

import (
	"fmt"
	"github.com/google/uuid"
)

func GetUUIDFromMap(t interface{}) (uuid.UUID, error) {
	parse, err := uuid.Parse(t.(string))
	if err != nil {
		return [16]byte{}, err
	}
	return parse, nil
}

func GetFloatFromMap(t interface{}) (float64, error) {
	switch t := t.(type) {
	case float64:
		return t, nil
	case int:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case float32:
		return float64(t), nil
	default:
		return 0, fmt.Errorf("type %T not supported", t)
	}
}
