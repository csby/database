package mssql

import (
	"errors"
	"fmt"
	"github.com/csby/database/sqldb"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

const (
	sqlFieldTagName              = "sql"
	sqlFieldFilterTagName        = "filter"
	sqlFieldOrderTagName         = "order"
	sqlFieldAutoIncrementTagName = "auto"
	sqlFieldPrimaryKeyTagName    = "primary"
	sqlFieldIndexTagName         = "index"

	sqlFunSchemaTagName = "SchemaName"
	sqlFunTableTagName  = "TableName"
)

type entity struct {
	name   string
	fields fieldCollection
}

// parse the name and fields of database table
// entity: address of the struct
func (s *entity) Parse(entity interface{}) error {
	s.name = ""
	s.fields = make([]*field, 0)

	// check kind of entity
	if entity == nil {
		return newError("invalid entity: nil")
	}
	if reflect.TypeOf(entity).Kind() != reflect.Ptr {
		return newError("invalid entity: not address")
	}
	v := reflect.ValueOf(entity).Elem()
	if v.Kind() != reflect.Struct {
		return newError("invalid entity (", v.Type().Name(), "): not struct")
	}

	err := s.parseName(v)
	if err != nil {
		return err
	}

	fields := make(map[string]*field)
	s.parseFields(v, fields)
	if len(fields) < 1 {
		return newError("invalid entity (", v.Type().Name(), "): field empty")
	}

	for _, field := range fields {
		s.fields = append(s.fields, field)
	}

	sort.Stable(s.fields)

	return nil
}

func (s *entity) ParseFilter(entity interface{}) error {
	s.name = ""
	s.fields = make([]*field, 0)

	// check kind of entity
	if entity == nil {
		return newError("invalid entity: nil")
	}
	if reflect.TypeOf(entity).Kind() != reflect.Ptr {
		return newError("invalid entity: not address")
	}
	v := reflect.ValueOf(entity).Elem()
	if v.Kind() != reflect.Struct {
		return newError("invalid entity (", v.Type().Name(), "): not struct")
	}

	s.parseFilterFields(v)
	if len(s.fields) < 1 {
		return newError("invalid entity (", v.Type().Name(), "): field empty")
	}

	return nil
}

func (s *entity) parseSchema(v reflect.Value) (string, error) {
	msgNotDefine := fmt.Sprintf("'func (s %s) %s() string' not define in struct", v.Type().Name(), sqlFunSchemaTagName)
	method := v.MethodByName(sqlFunSchemaTagName)
	if !method.IsValid() {
		return "", errors.New(msgNotDefine)
	}

	methodType := method.Type()
	if methodType.NumIn() != 0 {
		return "", errors.New(msgNotDefine)
	}
	if methodType.NumOut() != 1 {
		return "", errors.New(msgNotDefine)
	}
	if methodType.Out(0).Kind() != reflect.String {
		return "", errors.New(msgNotDefine)
	}

	result := method.Call([]reflect.Value{})
	if len(result) != 1 {
		return "", newError("get table name of '", v.Type().Name(), "' fail")
	}
	name := fmt.Sprintf("[%s]", result[0].String())
	if name == "[]" {
		return "", newError("invalid entity (", v.Type().Name(), "): table name is empty")
	}

	return name, nil
}

func (s *entity) parseName(v reflect.Value) error {
	msgNotDefine := fmt.Sprintf("'func (s %s) %s() string' not define in struct", v.Type().Name(), sqlFunTableTagName)
	method := v.MethodByName(sqlFunTableTagName)
	if !method.IsValid() {
		return errors.New(msgNotDefine)
	}

	methodType := method.Type()
	if methodType.NumIn() != 0 {
		return errors.New(msgNotDefine)
	}
	if methodType.NumOut() != 1 {
		return errors.New(msgNotDefine)
	}
	if methodType.Out(0).Kind() != reflect.String {
		return errors.New(msgNotDefine)
	}

	result := method.Call([]reflect.Value{})
	if len(result) != 1 {
		return newError("get table name of '", v.Type().Name(), "' fail")
	}
	name := fmt.Sprintf("[%s]", result[0].String())
	if name == "[]" {
		return newError("invalid entity (", v.Type().Name(), "): table name is empty")
	}

	schema, err := s.parseSchema(v)
	if err == nil && len(schema) > 0 {
		s.name = fmt.Sprintf("%s.%s", schema, name)
	} else {
		s.name = name
	}

	return nil
}

func (s *entity) parseFields(v reflect.Value, fields map[string]*field) {
	if v.Kind() != reflect.Struct {
		return
	}
	n := v.NumField()
	if n < 1 {
		return
	}
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return
	}
	if t.NumField() != n {
		return
	}

	for i := 0; i < n; i++ {
		valueField := v.Field(i)
		// ignore private field
		if !valueField.CanInterface() {
			continue
		}
		if !valueField.CanAddr() {
			continue
		}

		typeField := t.Field(i)
		// parent struct fields
		if typeField.Anonymous {
			if valueField.Kind() == reflect.Struct {
				s.parseFields(valueField.Addr().Elem(), fields)
			}
			continue
		}

		// filed define
		fieldName := typeField.Tag.Get(sqlFieldTagName)
		if fieldName == "" {
			continue
		}

		info := field{name: fmt.Sprintf("[%s]", fieldName), filter: "=", order: "ASC"}
		info.value = valueField.Interface()
		info.address = valueField.Addr().Interface()
		if strings.ToLower(typeField.Tag.Get(sqlFieldAutoIncrementTagName)) == "true" {
			info.autoIncrement = true
		}
		if strings.ToLower(typeField.Tag.Get(sqlFieldPrimaryKeyTagName)) == "true" {
			info.primaryKey = true
		}
		filter := typeField.Tag.Get(sqlFieldFilterTagName)
		if len(filter) > 0 {
			info.filter = filter
		}
		order := typeField.Tag.Get(sqlFieldOrderTagName)
		if len(order) > 0 {
			info.order = order
		}
		index := typeField.Tag.Get(sqlFieldIndexTagName)
		if len(index) > 0 {
			indexVal, err := strconv.Atoi(index)
			if err == nil {
				info.index = indexVal
			}
		}
		fields[fieldName] = &info

		//fmt.Println("field name:", info.name,
		//	", address:", info.address,
		//	", value:", info.value)
	}
}

func (s *entity) parseFilterFields(v reflect.Value) {
	if v.Kind() != reflect.Struct {
		return
	}
	n := v.NumField()
	if n < 1 {
		return
	}
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return
	}
	if t.NumField() != n {
		return
	}

	for i := 0; i < n; i++ {
		valueField := v.Field(i)
		// ignore private field
		if !valueField.CanInterface() {
			continue
		}
		if !valueField.CanAddr() {
			continue
		}

		typeField := t.Field(i)
		// parent struct fields
		if typeField.Anonymous {
			if valueField.Kind() == reflect.Struct {
				s.parseFilterFields(valueField.Addr().Elem())
			}
			continue
		}

		// filed define
		fieldName := typeField.Tag.Get(sqlFieldTagName)
		if fieldName == "" {
			continue
		}

		info := field{name: fmt.Sprintf("[%s]", fieldName), filter: "=", order: "ASC"}
		info.value = valueField.Interface()
		info.address = valueField.Addr().Interface()
		if strings.ToLower(typeField.Tag.Get(sqlFieldAutoIncrementTagName)) == "true" {
			info.autoIncrement = true
		}
		if strings.ToLower(typeField.Tag.Get(sqlFieldPrimaryKeyTagName)) == "true" {
			info.primaryKey = true
		}
		filter := typeField.Tag.Get(sqlFieldFilterTagName)
		if len(filter) > 0 {
			info.filter = filter
		}
		order := typeField.Tag.Get(sqlFieldOrderTagName)
		if len(order) > 0 {
			info.order = order
		}
		index := typeField.Tag.Get(sqlFieldIndexTagName)
		if len(index) > 0 {
			indexVal, err := strconv.Atoi(index)
			if err == nil {
				info.index = indexVal
			}
		}
		s.fields = append(s.fields, &info)
	}
}

func newError(v ...interface{}) error {
	return errors.New(fmt.Sprint(v...))
}

func (s *entity) fieldByName(name string) *field {
	count := len(s.fields)
	for i := 0; i < count; i++ {
		f := s.fields[i]
		if f.name == name {
			return f
		}
	}

	return &field{}
}

func (s *entity) Name() string {
	return s.name
}

func (s *entity) FieldCount() int {
	return len(s.fields)
}

func (s *entity) Field(i int) sqldb.SqlField {
	return s.fields[i]
}

func (s *entity) ScanFields() string {
	sb := &strings.Builder{}

	count := len(s.fields)
	if count > 0 {
		sb.WriteString(s.fields[0].name)

		for i := 1; i < count; i++ {
			sb.WriteString(", ")
			sb.WriteString(s.fields[i].name)
		}
	}

	return sb.String()
}

func (s *entity) ScanArgs() []interface{} {
	args := make([]interface{}, 0)

	count := len(s.fields)
	for i := 0; i < count; i++ {
		args = append(args, s.fields[i].address)
	}

	return args
}

func (s *entity) Values() []interface{} {
	values := make([]interface{}, 0)

	count := len(s.fields)
	for i := 0; i < count; i++ {
		values = append(values, s.fields[i].value)
	}

	return values
}
