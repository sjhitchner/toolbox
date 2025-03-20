package fileutils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
)

func LoadJSONFile[T any](filename string) ([]T, error) {
	list := make([]T, 0, 50)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for i := 1; ; i++ {
		var obj T

		err := decoder.Decode(&obj)
		if err == io.EOF {
			break
		}

		if err != nil {
			switch ee := err.(type) {
			case *json.SyntaxError:
				return nil, fmt.Errorf("Decode err o:%d, e:%+v", ee.Offset, ee)

			case *json.InvalidUTF8Error:
				return nil, fmt.Errorf("Decode err s:%s, e:%+v", ee.S, ee)

			case *json.InvalidUnmarshalError:
				return nil, fmt.Errorf("Decode err t:%v, e:%+v", ee.Type, ee)

			case *json.UnmarshalFieldError:
				return nil, fmt.Errorf("Decode err k:%s, t:%s, f:%v, e:%+v", ee.Key, ee.Type, ee.Field, ee)

			case *json.UnmarshalTypeError:
				return nil, fmt.Errorf("Decode err v:%s, t:%v, o:%d, s:%s, f:%s, e:%+v", ee.Value, ee.Type, ee.Offset, ee.Struct, ee.Field, ee)

			case *json.UnsupportedTypeError:
				return nil, fmt.Errorf("Decode err t:%v, e:%+v", ee.Type, ee)

			case *json.UnsupportedValueError:
				return nil, fmt.Errorf("Decode err v:%s, s:%s, e:%+v", ee.Value, ee.Str, ee)

			default:
				return nil, fmt.Errorf("Decode err (%d) %v", i, err)
			}
		}

		list = append(list, obj)
	}

	return list, nil
}

func StreamJSONFile[T any](filename string) (<-chan T, <-chan error) {
	outCh := make(chan T)
	errCh := make(chan error)

	go func() {
		defer close(outCh)
		defer close(errCh)

		file, err := os.Open(filename)
		if err != nil {
			errCh <- err
			return
		}
		defer file.Close()

		dec := json.NewDecoder(file)
		for {
			var obj T
			if err := dec.Decode(&obj); err != nil {
				if err == io.EOF {
					break
				}

				errCh <- errors.Wrapf(err, "StreamJSONFile.Decode")
				continue
			}

			outCh <- obj
		}
	}()
	return outCh, errCh
}

/*

	list := make([]T, 0, 50)
	decoder := json.NewDecoder(file)
	for i := 1; ; i++ {
		var obj T

		err := decoder.Decode(&obj)
		if err == io.EOF {
			break
		}

		if err != nil {
			switch ee := err.(type) {
			case *json.SyntaxError:
				return nil, fmt.Errorf("Decode err o:%d, e:%+v", ee.Offset, ee)

			case *json.InvalidUTF8Error:
				return nil, fmt.Errorf("Decode err s:%s, e:%+v", ee.S, ee)

			case *json.InvalidUnmarshalError:
				return nil, fmt.Errorf("Decode err t:%v, e:%+v", ee.Type, ee)

			case *json.UnmarshalFieldError:
				return nil, fmt.Errorf("Decode err k:%s, t:%s, f:%v, e:%+v", ee.Key, ee.Type, ee.Field, ee)

			case *json.UnmarshalTypeError:
				return nil, fmt.Errorf("Decode err v:%s, t:%v, o:%d, s:%s, f:%s, e:%+v", ee.Value, ee.Type, ee.Offset, ee.Struct, ee.Field, ee)

			case *json.UnsupportedTypeError:
				return nil, fmt.Errorf("Decode err t:%v, e:%+v", ee.Type, ee)

			case *json.UnsupportedValueError:
				return nil, fmt.Errorf("Decode err v:%s, s:%s, e:%+v", ee.Value, ee.Str, ee)

			default:
				return nil, fmt.Errorf("Decode err (%d) %v", i, err)
			}
		}

		list = append(list, obj)
	}

	return list, nil
}
*/
