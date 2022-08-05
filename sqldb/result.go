package sqldb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

type SqlSelectResult struct {
	Columns []*SqlSelectColumn `json:"columns"`
	Rows    []*SqlSelectRow    `json:"rows"`

	scans []interface{}
}

func (s *SqlSelectResult) Init(columnTypes []*sql.ColumnType) {
	s.Columns = make([]*SqlSelectColumn, 0)
	s.Rows = make([]*SqlSelectRow, 0)
	s.scans = make([]interface{}, 0)
	c := len(columnTypes)
	for i := 0; i < c; i++ {
		columnId := fmt.Sprintf("col%05d", i+1)
		columnType := columnTypes[i]
		columnName := columnType.Name()
		column := &SqlSelectColumn{
			Id:   columnId,
			Name: columnName,
			Type: columnType.DatabaseTypeName(),
		}
		nullable, ok := columnType.Nullable()
		if ok {
			column.Nullable = nullable
		}
		length, ok := columnType.Length()
		if ok {
			column.Length = length
		}
		precision, scale, ok := columnType.DecimalSize()
		if ok {
			column.Precision = precision
			column.Scale = scale
		}
		s.Columns = append(s.Columns, column)

		scanType := columnType.ScanType()
		if scanType == nil {
			var value *string
			addr := &value
			s.scans = append(s.scans, addr)
			column.getValue = func() (id string, value interface{}) {
				if *addr == nil {
					return columnId, nil
				}
				return columnId, **addr
			}
			continue
		}

		scanKind := scanType.Kind()
		switch scanKind {
		case reflect.Bool:
			if column.Nullable {
				var value *bool
				addr := &value
				s.scans = append(s.scans, addr)
				column.getValue = func() (id string, value interface{}) {
					if *addr == nil {
						return columnId, nil
					}
					return columnId, **addr
				}
			} else {
				var value bool
				addr := &value
				s.scans = append(s.scans, addr)
				column.getValue = func() (id string, value interface{}) {
					return columnId, *addr
				}
			}
		default:
			if strings.ToUpper(column.Type) == "DECIMAL" {
				var value *float64
				addr := &value
				s.scans = append(s.scans, addr)
				column.getValue = func() (id string, value interface{}) {
					if *addr == nil {
						return columnId, nil
					}
					return columnId, **addr
				}
			} else {
				value := reflect.New(scanType).Interface()
				addr := &value
				timeValue, tok := value.(*time.Time)
				if tok {
					timeAddr := &timeValue
					s.scans = append(s.scans, timeAddr)
					column.getValue = func() (id string, value interface{}) {
						if *timeAddr == nil {
							return columnId, nil
						} else {
							return columnId, (*timeAddr).Format("2006-01-02 15:04:05.000")
						}
					}
				} else {
					s.scans = append(s.scans, addr)
					column.getValue = func() (id string, value interface{}) {
						return columnId, *addr
					}
				}
			}
		}
	}
}

func (s SqlSelectResult) Scans() []interface{} {
	return s.scans
}

func (s *SqlSelectResult) AddScan(v interface{}) {
	if s.scans == nil {
		s.scans = make([]interface{}, 0)
	}

	s.scans = append(s.scans, v)
}

func (s *SqlSelectResult) AppendRow(columns []*SqlSelectColumn) {
	if s.Rows == nil {
		s.Rows = make([]*SqlSelectRow, 0)
	}

	row := &SqlSelectRow{
		kv: make(map[string]interface{}),
	}
	s.Rows = append(s.Rows, row)

	c := len(columns)
	for i := 0; i < c; i++ {
		column := columns[i]
		if column == nil {
			continue
		}

		id, val := column.GetValue()
		if len(id) < 1 {
			continue
		}

		row.kv[id] = val
	}
}

type SqlSelectColumn struct {
	Id        string `json:"id" note:"ID"`
	Name      string `json:"name" note:"名称"`
	Type      string `json:"type" note:"类型"`
	Nullable  bool   `json:"nullable" note:"是否可空"`
	Length    int64  `json:"length" note:"长度"`
	Precision int64  `json:"precision" note:"精度"`
	Scale     int64  `json:"scale" note:"小数位数"`

	getValue func() (id string, value interface{})
}

func (s SqlSelectColumn) GetValue() (id string, value interface{}) {
	if s.getValue != nil {
		return s.getValue()
	} else {
		return "", nil
	}
}

func (s *SqlSelectColumn) SetValue(v func() (id string, value interface{})) {
	s.getValue = v
}

type SqlSelectRow struct {
	kv map[string]interface{}
}

func (s *SqlSelectRow) SetValue(k string, v interface{}) {
	if s.kv == nil {
		s.kv = make(map[string]interface{})
	}

	s.kv[k] = v
}

func (s SqlSelectRow) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.kv)
}
