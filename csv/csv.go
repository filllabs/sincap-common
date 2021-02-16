package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"

	"strconv"
	"strings"
	"time"

	"gitlab.com/sincap/sincap-common/logging"
	"gitlab.com/sincap/sincap-common/reflection"
	"go.uber.org/zap"
)

var timeKind = reflect.TypeOf(time.Time{}).Kind()

const timeLayout = "02.01.2006"

// Tag holds basic properties of csv tag
type Tag struct {
	Name   string
	Ignore bool
	Index  int
}

// type OnItem func(i interface{}) error

// Read creates a csv reader from the given reader and returns a slice of the parsed rows
func Read(r io.Reader, t interface{}, hasTitleRow bool, delimiter rune, orderByTitles bool) ([]interface{}, error) {
	var recordArr []interface{}
	onItem := func(i interface{}) error {
		recordArr = append(recordArr, i)
		return nil
	}
	if err := ReadWithCallback(r, t, hasTitleRow, delimiter, orderByTitles, onItem); err != nil {
		return recordArr, err
	}
	return recordArr, nil
}

// ReadWithCallback creates a csv reader from the given reader and calls onItem function on evert row
func ReadWithCallback(r io.Reader, t interface{}, hasTitleRow bool, delimiter rune, orderByTitles bool, onItem func(i interface{}) error) error {
	typ := reflection.ExtractRealType(reflect.TypeOf(t))
	fieldLen := typ.NumField()
	logging.Logger.Debug("Type received", zap.String("type", typ.String()), zap.Int("fieldLen", fieldLen))
	reader := csv.NewReader(r)
	reader.Comma = delimiter
	reader.ReuseRecord = true
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	var tags []*Tag
	ignoredFieldCount := 0
	for i := 0; i < fieldLen; i++ {
		f := typ.Field(i)
		if tag, ok := f.Tag.Lookup("csv"); ok {
			parts := strings.Split(tag, ",")
			name := parts[0]
			if len(name) == 0 {
				name = f.Name
			}
			ct := &Tag{Name: name, Ignore: name == "-", Index: i}
			tags = append(tags, ct)
			if name == "-" {
				ignoredFieldCount++
			}
		} else {
			ct := &Tag{Name: f.Name, Ignore: false, Index: i}
			tags = append(tags, ct)
		}
	}
	logging.Logger.Debug("Tags read", zap.Any("type", typ.String()), zap.Int("tags", len(tags)))

	var columnIndexMatch []int
	if hasTitleRow {
		titles, err := reader.Read()
		if err == io.EOF {
			return err
		} else if err != nil {
			logging.Logger.Error("Can't title row", zap.Error(err))
			return err
		}
		logging.Logger.Debug("Titles read", zap.Any("titles", titles))
		if orderByTitles {
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
				if !tag.Ignore {
					return fmt.Errorf("Required field not found on CSV Title:%s Index:%d", tag.Name, tag.Index)
				}
				columnIndexMatch = append(columnIndexMatch, -1)
			}
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
			return err
		} else {
			row = line
		}

		// TODO: no need for this but check more, dig deeper
		// if len(row) > fieldLen-ignoredFieldCount {
		// 	return fmt.Errorf("Column count error at row %d Expected %d Received %d. Content: %s", rowIndex, fieldLen, len(row), strings.Join(row, ","))
		// }
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
			if cIndex == -1 {
				continue
			}
			if cIndex >= len(row) {
				continue
			}
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
					return fmt.Errorf(`Field type converting error.Coordinates: "%d:%d" Type: "%s" Field: "%s" Value: "%s". Error: "%v"`, rowIndex, fIndex, typ.Name(), field.String(), row[fIndex], e)
				}
			case reflect.Uint,
				reflect.Uint8,
				reflect.Uint16,
				reflect.Uint32,
				reflect.Uint64:
				if i, e := strconv.ParseUint(value, 10, 64); e == nil {
					field.SetUint(i)
				} else {
					return fmt.Errorf(`Field type converting error.Coordinates: "%d:%d" Type: "%s" Field: "%s" Value: "%s". Error: "%v"`, rowIndex, fIndex, typ.Name(), field.String(), row[fIndex], e)
				}
			case reflect.Float32,
				reflect.Float64:
				if i, e := strconv.ParseFloat(value, 64); e == nil {
					field.SetFloat(i)
				} else {
					return fmt.Errorf(`Field type converting error. Type: "%s" Field: "%s" Value: "%s". Error: "%v"`, typ.Name(), field.String(), row[fIndex], e)
				}
			case reflect.Bool:
				field.SetBool(value == "true")
			case timeKind:
				i, e := strconv.ParseInt(value, 10, 64)
				if e != nil {
					t, eT := time.Parse(timeLayout, value)
					if eT != nil {
						return fmt.Errorf(`Field type converting error.Coordinates: "%d:%d" Type: "%s" Field: "%s" Value: "%s". Error: "%v" \n "%v"`, rowIndex, fIndex, typ.Name(), field.String(), row[fIndex], e, eT)
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
				return fmt.Errorf(`Field type converting NOT FOUND error.Coordinates: "%d:%d" Type: "%s" Field: "%s" Value: "%s"`, rowIndex, fIndex, typ.Name(), field.String(), row[fIndex])
			}
		}
		if err := onItem(ins.Interface()); err != nil {
			return err
		}
	}
	return nil
}
