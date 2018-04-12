/*
 * Copyright (c) 2018, Jeremy Bingham (<jeremy@goiardi.gl>)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

func RemoveDupeSliceString(strSlice []string) []string {
	for i, u := range strSlice {
		if i > len(strSlice) {
			break
		}
		j := 1
		s := 0
		for {
			if i+j >= len(strSlice) {
				break
			}
			if u == strSlice[i+j] {
				j++
				s++
			} else {
				break
			}
		}
		if s == 0 {
			continue
		}
		strSlice = append(strSlice[:i+1], strSlice[i+1+s:]...)
	}
	return strSlice
}
