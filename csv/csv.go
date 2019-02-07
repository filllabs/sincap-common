package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"sincap-common/logging"

	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

var timeKind = reflect.TypeOf(time.Time{}).Kind()

const timeLayout = "02.01.2006"

type CsvTag struct {
	Name   string
	Ignore bool
	Index  int
}

// Read creates a csv reader from the given reader.
func Read(r io.Reader, t interface{}, hasTitleRow bool, delimiter rune, orderByTitles bool) ([]interface{}, error) {
	typ := reflect.TypeOf(t)
	var recordArr []interface{}
	fieldLen := typ.NumField()
	logging.Logger.Debug("Type received", zap.Any("type", typ.String()), zap.Int("fieldLen", fieldLen))
	reader := csv.NewReader(r)
	reader.Comma = delimiter
	reader.ReuseRecord = true
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	var tags []*CsvTag
	ignoredFieldCount := 0
	for i := 0; i < fieldLen; i++ {
		f := typ.Field(i)
		if tag, ok := f.Tag.Lookup("csv"); ok {
			parts := strings.Split(tag, ",")
			name := parts[0]
			if len(name) == 0 {
				name = f.Name
			}
			ct := &CsvTag{Name: name, Ignore: name == "-", Index: i}
			tags = append(tags, ct)
			if name == "-" {
				ignoredFieldCount++
			}
		} else {
			ct := &CsvTag{Name: f.Name, Ignore: false, Index: i}
			tags = append(tags, ct)
		}
	}
	logging.Logger.Debug("Tags read", zap.Any("type", typ.String()), zap.Int("tags", len(tags)))

	var columnIndexMatch []int
	if hasTitleRow {
		titles, err := reader.Read()
		if err == io.EOF {
			return recordArr, err
		} else if err != nil {
			logging.Logger.Error("Can't title row", zap.Error(err))
			return recordArr, err
		}
		logging.Logger.Debug("Titles read", zap.Any("titles", titles))
	outer:
		for i := 0; i < len(tags); i++ {
			tag := tags[i]
			for j := 0; j < len(titles); j++ {
				title := titles[j]
				if !tag.Ignore && strings.TrimSpace(title) == tag.Name {
					columnIndexMatch = append(columnIndexMatch, j)
					continue outer
				}
			}
			columnIndexMatch = append(columnIndexMatch, -1)
		}
		logging.Logger.Debug("Titles Matched", zap.Any("matches", columnIndexMatch))
	}
	logging.Logger.Debug("Reading Rows")
	for rowIndex := 0; true; rowIndex++ {
		var row []string
		if line, err := reader.Read(); err == io.EOF {
			break
		} else if err != nil {
			logging.Logger.Error("Can't read csv", zap.Error(err))
			return recordArr, err
		} else {
			row = line
		}

		if len(row) > fieldLen-ignoredFieldCount {
			return recordArr, fmt.Errorf("Column count error at row %d Expected %d Received %d. Content: %s", rowIndex, fieldLen, len(row), strings.Join(row, ","))
		}

		//TODO: support callback on
		//TODO: add pointer support
		ins := reflect.New(typ).Elem()
		for fIndex := 0; fIndex < fieldLen; fIndex++ {
			cIndex := fIndex
			if orderByTitles {
				tag := tags[fIndex]
				cIndex = columnIndexMatch[fIndex]
				if tag.Ignore {
					continue
				}
			}
			field := ins.Field(fIndex)
			value := row[cIndex]

			if value == "" {
				continue
			}
			isPtr := false
			kind := field.Type().Kind()
			if kind == reflect.Ptr {
				kind = field.Type().Elem().Kind()
				isPtr = true
			}
			switch kind {
			case reflect.String:
				field.SetString(value)
			case reflect.Int,
				reflect.Int8,
				reflect.Int16,
				reflect.Int32,
				reflect.Int64:
				if i, e := strconv.Atoi(value); e == nil {
					field.SetInt(int64(i))
				} else {
					//TODO: field.String() is not printing good
					return recordArr, fmt.Errorf(`Field type converting error.Coordinates: "%d:%d" Type: "%s" Field: "%s" Value: "%s". Error: "%v"`, rowIndex, fIndex, typ.Name(), field.String(), row[fIndex], e)
				}
			case reflect.Uint,
				reflect.Uint8,
				reflect.Uint16,
				reflect.Uint32,
				reflect.Uint64:
				if i, e := strconv.ParseUint(value, 10, 64); e == nil {
					field.SetUint(i)
				} else {
					return recordArr, fmt.Errorf(`Field type converting error.Coordinates: "%d:%d" Type: "%s" Field: "%s" Value: "%s". Error: "%v"`, rowIndex, fIndex, typ.Name(), field.String(), row[fIndex], e)
				}
			case reflect.Float32,
				reflect.Float64:
				if i, e := strconv.ParseFloat(value, 64); e == nil {
					field.SetFloat(i)
				} else {
					return recordArr, fmt.Errorf(`Field type converting error. Type: "%s" Field: "%s" Value: "%s". Error: "%v"`, typ.Name(), field.String(), row[fIndex], e)
				}
			case reflect.Bool:
				field.SetBool(value == "true")
			case timeKind:
				i, e := strconv.ParseInt(value, 10, 64)
				if e != nil {
					t, eT := time.Parse(timeLayout, value)
					if eT != nil {
						return recordArr, fmt.Errorf(`Field type converting error.Coordinates: "%d:%d" Type: "%s" Field: "%s" Value: "%s". Error: "%v" \n "%v"`, rowIndex, fIndex, typ.Name(), field.String(), row[fIndex], e, eT)
					}
					if isPtr {
						field.Set(reflect.ValueOf(&t))
					} else {
						field.Set(reflect.ValueOf(t))
					}
				} else {
					field.Set(reflect.ValueOf(time.Unix(0, i*int64(time.Millisecond))))
				}
			default:
				return recordArr, fmt.Errorf(`Field type converting NOT FOUND error.Coordinates: "%d:%d" Type: "%s" Field: "%s" Value: "%s"`, rowIndex, fIndex, typ.Name(), field.String(), row[fIndex])
			}
		}
		recordArr = append(recordArr, ins.Interface())
	}
	return recordArr, nil
}
