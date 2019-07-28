package oracle

import (
	"fmt"
	"github.com/csby/database/sqldb"
	"strings"
)

type builder struct {
	query              []string
	args               []interface{}
	insertFields       []string
	insertPlaceholders []string
	hasWhere           bool
	hasOrder           bool
	hasSet             bool
}

func (s *builder) Reset() sqldb.SqlBuilder {
	s.query = make([]string, 0)
	s.args = make([]interface{}, 0)
	s.insertFields = make([]string, 0)
	s.insertPlaceholders = make([]string, 0)
	s.hasWhere = false
	s.hasOrder = false
	s.hasSet = false

	return s
}

func (s *builder) Select(query string, distinct bool) sqldb.SqlBuilder {
	s.query = make([]string, 1)
	if distinct {
		s.query[0] = fmt.Sprint("SELECT DISTINCT ", query)
	} else {
		s.query[0] = fmt.Sprint("SELECT ", query)
	}

	return s
}

func (s *builder) Insert(query string) sqldb.SqlBuilder {
	s.query = make([]string, 1)
	s.query[0] = fmt.Sprint("INSERT ", query)

	return s
}

func (s *builder) Delete(query string) sqldb.SqlBuilder {
	s.query = make([]string, 1)
	s.query[0] = fmt.Sprint("DELETE FROM ", query)

	return s
}

func (s *builder) Update(query string) sqldb.SqlBuilder {
	s.query = make([]string, 1)
	s.query[0] = fmt.Sprint("UPDATE ", query)

	return s
}

func (s *builder) From(query string) sqldb.SqlBuilder {
	if s.query == nil {
		s.query = make([]string, 0)
	}
	s.query = append(s.query, fmt.Sprint(" FROM ", query))

	return s
}

func (s *builder) Value(filed string, value interface{}) sqldb.SqlBuilder {
	s.insertFields = append(s.insertFields, filed)
	s.insertPlaceholders = append(s.insertPlaceholders, s.argName())
	s.args = append(s.args, value)

	return s
}

func (s *builder) Set(filed string, value interface{}) sqldb.SqlBuilder {
	if s.hasSet {
		s.query = append(s.query, fmt.Sprint(", ", filed, " = "), s.argName())
	} else {
		s.hasSet = true
		s.query = append(s.query, fmt.Sprint("SET ", filed, " = ", s.argName()))
	}

	if s.args == nil {
		s.args = make([]interface{}, 0)
	}
	s.args = append(s.args, value)

	return s
}

func (s *builder) WhereFormatAnd(format string, a ...interface{}) sqldb.SqlBuilder {
	if s.query == nil {
		s.query = make([]string, 0)
	}

	if s.hasWhere {
		s.query = append(s.query, "AND ")
	} else {
		s.hasWhere = true
		s.query = append(s.query, "WHERE ")
	}

	s.query = append(s.query, fmt.Sprintf(format, s.formatArgs(a)...))

	return s
}

func (s *builder) WhereFormatOr(format string, a ...interface{}) sqldb.SqlBuilder {
	if s.query == nil {
		s.query = make([]string, 0)
	}

	if s.hasWhere {
		s.query = append(s.query, "OR ")
	} else {
		s.hasWhere = true
		s.query = append(s.query, "WHERE ")
	}

	s.query = append(s.query, fmt.Sprintf(format, s.formatArgs(a)...))

	return s
}

func (s *builder) WhereFormat(format string, a ...interface{}) sqldb.SqlBuilder {
	if s.query == nil {
		s.query = make([]string, 0)
	}

	if s.hasWhere {
		s.query = append(s.query, " ")
	} else {
		s.hasWhere = true
		s.query = append(s.query, "WHERE ")
	}

	s.query = append(s.query, fmt.Sprintf(format, s.formatArgs(a)...))

	return s
}

func (s *builder) WhereAnd(query string, args ...interface{}) sqldb.SqlBuilder {
	if s.query == nil {
		s.query = make([]string, 0)
	}

	if s.hasWhere {
		s.query = append(s.query, fmt.Sprint("AND ", query))
	} else {
		s.hasWhere = true
		s.query = append(s.query, fmt.Sprint("WHERE ", query))
	}

	if s.args == nil {
		s.args = make([]interface{}, 0)
	}
	s.args = append(s.args, args...)

	return s
}

func (s *builder) WhereOr(query string, args ...interface{}) sqldb.SqlBuilder {
	if s.query == nil {
		s.query = make([]string, 0)
	}

	if s.hasWhere {
		s.query = append(s.query, fmt.Sprint("OR ", query))
	} else {
		s.hasWhere = true
		s.query = append(s.query, fmt.Sprint("WHERE ", query))
	}

	if s.args == nil {
		s.args = make([]interface{}, 0)
	}
	s.args = append(s.args, args...)

	return s
}

func (s *builder) Where(query string, args ...interface{}) sqldb.SqlBuilder {
	if s.query == nil {
		s.query = make([]string, 0)
	}

	if s.hasWhere {
		s.query = append(s.query, fmt.Sprint(" ", query))
	} else {
		s.hasWhere = true
		s.query = append(s.query, fmt.Sprint("WHERE ", query))
	}

	if s.args == nil {
		s.args = make([]interface{}, 0)
	}
	s.args = append(s.args, args...)

	return s
}

func (s *builder) Order(query string) sqldb.SqlBuilder {
	if s.query == nil {
		s.query = make([]string, 0)
	}
	if s.hasOrder {
		s.query = append(s.query, fmt.Sprint(", ", query))
	} else {
		s.hasOrder = true
		s.query = append(s.query, fmt.Sprint("ORDER BY ", query))
	}

	return s
}

func (s *builder) Append(query string, args ...interface{}) sqldb.SqlBuilder {
	if s.query == nil {
		s.query = make([]string, 0)
	}
	s.query = append(s.query, query)

	if s.args == nil {
		s.args = make([]interface{}, 0)
	}
	s.args = append(s.args, args...)

	return s
}

func (s *builder) AppendFormat(format string, a ...interface{}) sqldb.SqlBuilder {
	if s.query == nil {
		s.query = make([]string, 0)
	}
	s.query = append(s.query, fmt.Sprintf(format, s.formatArgs(a)...))

	return s
}

func (s *builder) Query() string {
	if len(s.insertFields) > 0 {
		return fmt.Sprint(strings.Join(s.query, " "), " (", strings.Join(s.insertFields, ","), ") values (", strings.Join(s.insertPlaceholders, ","), ")")
	}

	return fmt.Sprint(strings.Join(s.query, " "))
}

func (s *builder) Args() []interface{} {
	return s.args
}

func (s *builder) formatArgs(args []interface{}) []interface{} {
	as := make([]interface{}, 0)

	for argNum := 0; argNum < len(args); argNum++ {
		arg := args[argNum]
		switch av := arg.(type) {
		case []int64, []int32, []int16, []int8, []int, []uint64, []uint32, []uint16, []uint8, []uint:
			{
				text := fmt.Sprint(av)
				text = strings.Replace(text, " ", ",", -1)
				text = strings.Replace(text, "[", "(", -1)
				text = strings.Replace(text, "]", ")", -1)
				as = append(as, text)
				break
			}
		case []string:
			{
				text := strings.Join(av, "','")
				as = append(as, fmt.Sprintf("('%s')", text))
				break
			}
		default:
			{
				as = append(as, av)
				break
			}
		}
	}

	return as
}

func (s *builder) argName() string {
	return fmt.Sprintf(":%d", len(s.args)+1)
}

func (s *builder) ArgName() string {
	return s.argName()
}
