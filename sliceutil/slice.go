package commonutil

import (
	"math/rand"
	"reflect"
	"strings"
	"time"
)

// InSlice checks given string in string slice or not.
func InSlice(v string, sl []string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

// InSliceIface checks given interface in interface slice.
func InSliceIface(v interface{}, sl []interface{}) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

// ContainsSliceItem contain slice
func ContainsSliceItem(v string, sl []string) bool {
	for _, vv := range sl {
		if strings.Contains(v, vv) {
			return true
		}
	}
	return false
}

// InSliceCommon in slice common.
func InSliceCommon(raw interface{}, rawSlice interface{}) bool {
	sliceValue := reflect.ValueOf(rawSlice)
	if sliceValue.Kind() != reflect.Slice || sliceValue.Len() == 0 {
		return false
	}

	rawValue := reflect.ValueOf(raw)

	for i := 0; i < sliceValue.Len(); i++ {
		element := sliceValue.Index(i)
		if element.Kind() == reflect.Interface {
			element = element.Elem()
		}
		if !rawValue.IsValid() && !element.IsValid() {
			return true
		} else if !rawValue.IsValid() || !element.IsValid() {
			continue
		}
		if rawValue.Kind() != element.Kind() {
			continue
		}
		if reflect.DeepEqual(element.Interface(), rawValue.Interface()) {
			return true
		}
	}

	return false
}

// SliceRandList generate an int slice from min to max.
func SliceRandList(min, max int) []int {
	if max < min {
		min, max = max, min
	}
	length := max - min + 1
	t0 := time.Now()
	rand.Seed(int64(t0.Nanosecond()))
	list := rand.Perm(length)
	for index := range list {
		list[index] += min
	}
	return list
}

// SliceMerge merges interface slices to one slice.
func SliceMerge(slice1, slice2 []interface{}) (c []interface{}) {
	c = append(slice1, slice2...)
	return
}

// SliceDiff returns diff slice of slice1 - slice2.
func SliceDiff(slice1, slice2 []interface{}) (diffslice []interface{}) {
	for _, v := range slice1 {
		if !InSliceIface(v, slice2) {
			diffslice = append(diffslice, v)
		}
	}
	return
}

// SliceChunk separates one slice to some sized slice.
func SliceChunk(slice []interface{}, size int) (chunkslice [][]interface{}) {
	if size >= len(slice) {
		chunkslice = append(chunkslice, slice)
		return
	}
	end := size
	for i := 0; i <= (len(slice) - size); i += size {
		chunkslice = append(chunkslice, slice[i:end])
		end += size
	}
	return
}

// SliceUnique cleans repeated values in slice.
func SliceUnique(slice []interface{}) (uniqueslice []interface{}) {
	for _, v := range slice {
		if !InSliceIface(v, uniqueslice) {
			uniqueslice = append(uniqueslice, v)
		}
	}
	return
}

// SliceShuffle shuffles a slice.
func SliceShuffle(slice []interface{}) []interface{} {
	for i := 0; i < len(slice); i++ {
		a := rand.Intn(len(slice))
		b := rand.Intn(len(slice))
		slice[a], slice[b] = slice[b], slice[a]
	}
	return slice
}
