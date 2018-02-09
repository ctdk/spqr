package util

// common utility functions

func SliceEqual(oldSlice []string, NewSlice []string) bool {
	if len(oldSlice) != len(NewSlice) {
		return false
	}

	for i := 0; i < len(oldSlice); i++ {
		if oldSlice[i] != NewSlice[i] {
			return false
		}
	}

	return true
}
