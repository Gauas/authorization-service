package bitmap

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func Offset[T Integer](id T) (int64, bool) {
	v := int64(id)
	if v <= 0 {
		return 0, false
	}
	return v, true
}
