// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package dbs

func anyError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func MakeSlice[T any](capacity int) []T {
	if capacity <= 0 {
		return make([]T, 0)
	}
	switch capacity {
	case 1:
		return make([]T, 0, 1)
	case 2:
		return make([]T, 0, 2)
	case 3:
		return make([]T, 0, 3)
	case 4:
		return make([]T, 0, 4)
	case 5:
		return make([]T, 0, 5)
	case 6:
		return make([]T, 0, 6)
	case 7:
		return make([]T, 0, 7)
	case 8:
		return make([]T, 0, 8)
	case 9:
		return make([]T, 0, 9)
	case 10:
		return make([]T, 0, 10)
	default:
		return make([]T, 0, capacity)
	}
}
