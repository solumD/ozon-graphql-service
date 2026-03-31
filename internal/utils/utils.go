package utils

func CloneInt64Ptr(value *int64) *int64 {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}

func ToIntPtr(value *int64) *int {
	if value == nil {
		return nil
	}

	converted := int(*value)
	return &converted
}

func ToInt64Ptr(value *int) *int64 {
	if value == nil {
		return nil
	}

	converted := int64(*value)
	return &converted
}
