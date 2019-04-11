package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortParse(t *testing.T) {
	sort := Sort{}
	err := sort.Parse("")
	assert.Error(t, err, "Empty param must throw error")

	err2 := sort.Parse("   ")
	assert.Error(t, err2, "Empty spaced param must throw error")

	err3 := sort.Parse("1")
	assert.Error(t, err3, "Short spaced param must throw error")

	err4 := sort.Parse("1   ")
	assert.Error(t, err4, "Short param must throw error")

	err5 := sort.Parse("_sample")
	assert.Error(t, err5, "Wrong started param must throw error")

	err6 := sort.Parse("-sample")
	assert.NoError(t, err6, "- started param does not throw error")
	assert.Equal(t, sort.Direction, DSC, "- must be DSC")
	assert.Equal(t, sort.Name, "sample", "name must be sample")

	err7 := sort.Parse("+sample")
	assert.NoError(t, err7, "+ started param does not throw error")
	assert.Equal(t, sort.Direction, ASC, "+ must be ASC")
	assert.Equal(t, sort.Name, "sample", "name must be sample")
	err8 := sort.Parse(" sample")
	assert.NoError(t, err8, "' ' started param does not throw error")
	assert.Equal(t, sort.Direction, ASC, "' ' must be ASC")
	assert.Equal(t, sort.Name, "sample", "name must be sample")

}
