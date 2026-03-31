package utils

func CloneInt64Ptr(value *int64) *int64 {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}
