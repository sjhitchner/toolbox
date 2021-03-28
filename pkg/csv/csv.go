/*

const CSV = "foo,bar,10,01/02/2006"

type Row struct {
	Foo string
	Bar string
	Amount float44
	Date time.Time `csv:"01/02/2006"`
}


reader := csv.NewReader(strings.NewReader(CSV))

var rows []Row
err := NewDecoder(reader).Decode(&rows)

*/
package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	defaultDateFormat = time.RFC3339
)

type Decoder struct {
	reader     *csv.Reader
	hasHeader  bool
	dateFormat string
}

func NewDecoder(reader *csv.Reader) *Decoder {
	return &Decoder{
		reader:     reader,
		hasHeader:  false,
		dateFormat: defaultDateFormat,
	}
}

func (t *Decoder) HasHeader() {
	t.hasHeader = true
}

func (t *Decoder) SetDateFormat(format string) {
	t.dateFormat = format
}

func (t *Decoder) Decode(rows interface{}) error {

	for i := 1; ; {
		row, err := t.reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if t.hasHeader && i == 1 {
			continue
		}

		if err := t.parseRow(rows, row); err != nil {
			return errors.Wrapf(err, "error parsing row %d", i)
		}
	}

	return nil
}

func (t *Decoder) parseRow(arrPtr interface{}, row []string) error {

	typ := reflect.TypeOf(arrPtr).Elem().Elem()
	v := reflect.New(typ).Elem()

	arr := reflect.ValueOf(arrPtr).Elem()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		switch field.Type.Name() {

		case "bool":
			b, err := strconv.ParseBool(row[i])
			if err != nil {
				return errors.Wrapf(err, "invalid string float %v", v)
			}

			v.Field(i).SetBool(b)

		case "float64":
			f, err := strconv.ParseFloat(row[i], 64)
			if err != nil {
				return errors.Wrapf(err, "invalid string float %v", v)
			}

			v.Field(i).SetFloat(f)

		case "string":
			v.Field(i).SetString(row[i])

		case "int64":
			n, err := strconv.ParseInt(row[i], 10, 64)
			if err != nil {
				return errors.Wrapf(err, "invalid string int %v", v)
			}

			v.Field(i).SetInt(n)

		case "Time":
			format := field.Tag.Get("csv")
			if format == "" {
				format = t.dateFormat
			}

			date, err := time.Parse(format, row[i])
			if err != nil {
				return errors.Wrapf(err, "invalid string date %v", v)
			}

			v.Field(i).Set(reflect.ValueOf(date))

		default:
			fmt.Println("None", field, field.Type.Name())

		}
	}

	arr.Set(reflect.Append(arr, v))

	return nil
}
