package csv

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Sample struct {
	ID       uint    `csv:"id"`
	Name     string  `csv:"Name"`
	Owner    string  `csv:"Owner"`
	Count    uint    `csv:"Count"`
	Account  float32 `csv:"Account"`
	Height   float32 `csv:"Height"`
	Width    float32 `csv:"Width"`
	Material string  `csv:"Material"`
	Class    string  `csv:"Class"`
	Rate     float32 `csv:"Rate"`
}
type SampleIgnored struct {
	ID       uint    `csv:"ID"`
	Name     string  `csv:"Name"`
	Owner    string  `csv:"Owner"`
	Count    uint    `csv:"Count"`
	Account  float32 `csv:"-"`
	Height   float32 `csv:"Height"`
	Width    float32 `csv:"Width"`
	Material string  `csv:"Material"`
	Class    string  `csv:"-"`
	Rate     float32 `csv:"Rate"`
}

func TestRead(t *testing.T) {
	typ := Sample{}
	path, pErr := filepath.Abs("../../test/sampleCsv5000NoTitle.csv")
	assert.NoError(t, pErr, "Cant't find path")
	csvFile, fErr := os.Open(path)
	assert.NoError(t, fErr, "Cant't find file")
	records, err := Read(bufio.NewReader(csvFile), typ, false, ',', false)
	assert.NoError(t, err, "Can't read CSV file")
	assert.Equal(t, 5000, len(records), "CSV length is wrong!")
	assert.Equal(t, uint(1), records[0].(Sample).ID, "CSV first Elemnent is wrong!")
}

func TestReadTitle(t *testing.T) {
	typ := Sample{}
	path, pErr := filepath.Abs("../../test/sampleCsv5000Title.csv")
	assert.NoError(t, pErr, "Cant't find path")
	csvFile, fErr := os.Open(path)
	assert.NoError(t, fErr, "Cant't find file")
	records, err := Read(bufio.NewReader(csvFile), typ, true, ',', false)
	assert.NoError(t, err, "Can't read CSV file ")
	assert.Equal(t, 5000, len(records), "CSV length is wrong!")
	assert.Equal(t, uint(1), records[0].(Sample).ID, "CSV first Elemnent is wrong!")
}

func TestReadMixedTitle(t *testing.T) {
	typ := Sample{}
	path, pErr := filepath.Abs("../../test/sampleCsv5000MixedTitle.csv")
	assert.NoError(t, pErr, "Cant't find path")
	csvFile, fErr := os.Open(path)
	assert.NoError(t, fErr, "Cant't find file")
	records, err := Read(bufio.NewReader(csvFile), typ, true, ';', true)
	assert.NoError(t, err, "Can't read CSV file ")
	assert.Equal(t, 5000, len(records), "CSV length is wrong!")
	assert.Equal(t, "Nunavut", records[0].(Sample).Material, "CSV first Elemnent is wrong!")
}

func TestReadIgnoredTitle(t *testing.T) {
	typ := SampleIgnored{}
	path, pErr := filepath.Abs("../../test/sampleCsv5000IgnoredTitle.csv")
	assert.NoError(t, pErr, "Cant't find path")
	csvFile, fErr := os.Open(path)
	assert.NoError(t, fErr, "Cant't find file")
	records, err := Read(bufio.NewReader(csvFile), typ, true, ';', true)
	assert.NoError(t, err, "Can't read CSV file ")
	assert.Equal(t, 5000, len(records), "CSV length is wrong!")
	assert.Equal(t, "Nunavut", records[0].(SampleIgnored).Material, "CSV first Elemnent is wrong!")
}
