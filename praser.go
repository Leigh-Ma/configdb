package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type Parser struct {
	FieldsSplitter string
	LinesSplitter  string
	PanicOnError   bool
	Tag            string

	fieldsParser map[reflect.Type]reflect.Value
}

type IAfterParse interface {
	AfterParse()
}

func NewParser(fieldSplitter, lineSplitter, fieldTag string, panicOnError bool) *Parser {
	if fieldSplitter == "" || lineSplitter == "" || fieldTag == "" {
		panic("neither of the string paramters should be blank")
		return nil
	}

	p := &Parser{
		FieldsSplitter: fieldSplitter,
		LinesSplitter:  lineSplitter,
		PanicOnError:   panicOnError,
		Tag:            fieldTag,
		fieldsParser:   make(map[reflect.Type]reflect.Value, 0),
	}

	p.RegisterParser(asString)
	p.RegisterParser(asBool)
	p.RegisterParser(asInt)
	p.RegisterParser(asInt8)
	p.RegisterParser(asInt16)
	p.RegisterParser(asInt32)
	p.RegisterParser(asInt64)
	p.RegisterParser(asFloat32)
	p.RegisterParser(asFloat64)

	return p
}

func ReadFile(tableDataFile string) (string, error) {
	fi, err := os.Open(tableDataFile)
	if err != nil {
		return "", err
	}
	defer fi.Close()

	data, err := ioutil.ReadAll(fi)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p Parser) ParseFieldIndex(titleLine string) map[string]int {
	fieldIndex := make(map[string]int, 0)
	for index, field := range strings.Split(titleLine, p.FieldsSplitter) {
		fieldIndex[field] = index
	}
	return fieldIndex
}

func (p Parser) ParseRecord(line string, fieldIndex map[string]int, record interface{}) error {
	pointerType := reflect.TypeOf(record)

	if pointerType.Kind() != reflect.Ptr {
		err := errors.New("parse: record should be POINTOR to a table record struct")
		if p.PanicOnError {
			panic(err.Error())
		}
		return err
	}

	structType := pointerType.Elem()
	if structType.Kind() != reflect.Struct {
		err := errors.New("parse: record should be pointor to a table record STRUCT")
		if p.PanicOnError {
			panic(err.Error())
		}
		return err
	}

	fieldArrayData := strings.Split(line, p.FieldsSplitter)
	fieldArrayLen := len(fieldArrayData)

	structValue := reflect.ValueOf(record).Elem()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		tag := field.Tag.Get(p.Tag)
		if tag != "" {
			index, ok := fieldIndex[tag]
			if !ok || index >= fieldArrayLen {
				err := errors.New(fmt.Sprintf("parse: tag %s [%s] error or data error", tag, field.Name))
				if p.PanicOnError {
					panic(err.Error())
				}
				return err
			}
			out := p.parseField(fieldArrayData[index], field.Type)
			value, err := out[0], out[1]
			if !err.IsNil() {
				e := err.Interface().(error)
				if p.PanicOnError {
					panic(e.Error())
				}
				return e
			}
			structValue.Field(i).Set(value)
		}
	}

	if iAfter, ok := record.(IAfterParse); ok {
		iAfter.AfterParse()
	}

	return nil
}

func (p Parser) FormatRecord(record interface{}) string {
	s := "{ "
	structType := reflect.TypeOf(record).Elem()
	structValue := reflect.ValueOf(record).Elem()
	for i := 0; i < structValue.NumField(); i++ {
		s += structType.Field(i).Name + ": "
		s += fmt.Sprint(structValue.Field(i)) + ", "
	}
	return s + "}"
}

func (p *Parser) RegisterParser(fun interface{}) error {
	funcType := reflect.TypeOf(fun)

	if funcType.Kind() != reflect.Func {
		err := errors.New(fmt.Sprintf("parse: should not register other than function"))
		if p.PanicOnError {
			panic(err.Error())
		}
		return err
	}

	if funcType.NumIn() != 1 || funcType.In(0).Kind() != reflect.String {
		err := errors.New(fmt.Sprintf("parse: field parser function should have one IN param with type string"))
		if p.PanicOnError {
			panic(err.Error())
		}
		return err
	}

	if funcType.NumOut() != 2 || funcType.Out(1).Name() != "error" {
		err := errors.New(fmt.Sprintf("parse: field parser function should have 2 out param with second type error"))
		if p.PanicOnError {
			panic(err.Error())
		}
		return err
	}

	fieldType := funcType.Out(0)
	p.fieldsParser[fieldType] = reflect.ValueOf(fun)

	return nil
}

func (p Parser) parseField(field string, fieldType reflect.Type) []reflect.Value {
	fieldParser, ok := p.fieldsParser[fieldType]
	if !ok {
		err := errors.New(fmt.Sprintf("parse: can not find field parser for field %v", fieldType))
		if p.PanicOnError {
			panic(err.Error())
			return []reflect.Value{reflect.ValueOf(nil), reflect.ValueOf(err)}
		}
	}

	return fieldParser.Call([]reflect.Value{reflect.ValueOf(field)})
}

func asString(o string) (string, error) {
	return string(o), nil
}

func asBool(o string) (bool, error) {
	if o == "0" || o == "false" {
		return false, nil
	} else if o == "1" || o == "true" {
		return true, nil
	}
	return false, errors.New("can not parse to a bool value")
}

func asInt(o string) (int, error) {
	return strconv.Atoi(o)
}

func asInt8(o string) (int8, error) {
	value, err := strconv.Atoi(o)
	if err != nil {
		return 0, err
	}
	return int8(value), nil
}

func asInt16(o string) (int16, error) {
	value, err := strconv.Atoi(o)
	if err != nil {
		return 0, err
	}
	return int16(value), nil
}

func asInt32(o string) (int32, error) {
	value, err := strconv.ParseInt(o, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(value), nil
}

func asInt64(o string) (int64, error) {
	value, err := strconv.ParseInt(o, 10, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func asFloat32(o string) (float32, error) {
	value, err := strconv.ParseFloat(o, 32)
	if err != nil {
		return 0., err
	}
	return float32(value), nil
}

func asFloat64(o string) (float64, error) {
	value, err := strconv.ParseFloat(o, 32)
	if err != nil {
		return 0., err
	}
	return value, nil
}
