package utils

// CloneInt64Ptr возвращает копию указателя на int64
func CloneInt64Ptr(value *int64) *int64 {
	if value == nil {
		return nil
	}

	cloned := *value
	return &cloned
}

// ToIntPtr конвертирует указатель int64 в int и возвращает его
func ToIntPtr(value *int64) *int {
	if value == nil {
		return nil
	}

	converted := int(*value)
	return &converted
}

// ToInt64Ptr конвертирует указатель int в int64 и возвращает его
func ToInt64Ptr(value *int) *int64 {
	if value == nil {
		return nil
	}

	converted := int64(*value)
	return &converted
}
