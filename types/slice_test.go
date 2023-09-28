package types_test

import (
	"testing"

	"github.com/filllabs/sincap-common/types"
	"github.com/stretchr/testify/assert"
)

func TestContainsSlice(t *testing.T) {
	slice := []string{"init", "server"}
	slicePtr := &[]string{"init", "server"}
	ptrSlice := []*string{&slice[0], &slice[1]}
	ptrSlicePtr := &[]*string{&slice[0], &slice[1]}
	target := "init"
	targetPtr := &target

	assert.Equal(t, true, types.SliceContains(slice, target), "Can't find  target in slice")
	assert.Equal(t, true, types.SliceContains(slicePtr, target), "Can't find target in slicePtr")
	assert.Equal(t, true, types.SliceContains(ptrSlice, target), "Can't find target in ptrSlice")
	assert.Equal(t, true, types.SliceContains(ptrSlicePtr, target), "Can't find target in ptrSlicePtr")

	assert.Equal(t, true, types.SliceContains(slice, targetPtr), "Can't find  targetPtr in slice")
	assert.Equal(t, true, types.SliceContains(slicePtr, targetPtr), "Can't find targetPtr in slicePtr")
	assert.Equal(t, true, types.SliceContains(ptrSlice, targetPtr), "Can't find targetPtr in ptrSlice")
	assert.Equal(t, true, types.SliceContains(ptrSlicePtr, targetPtr), "Can't find targetPtr in ptrSlicePtr")
}

func TestSliceOfString(t *testing.T) {
	assert.Equal(t, []string{"a", "a", "a"}, types.SliceOfString("a", 3))
}
