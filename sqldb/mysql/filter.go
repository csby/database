package mysql

type filter struct {
	fieldOr bool
	groupOr bool
	fields  interface{}
}

func newFilter(entity interface{}, fieldOr, groupOr bool) *filter {
	return &filter{
		fieldOr: fieldOr,
		groupOr: groupOr,
		fields:  entity,
	}
}

func (s *filter) FieldOr() bool {
	return s.fieldOr
}

func (s *filter) GroupOr() bool {
	return s.groupOr
}

func (s *filter) Fields() interface{} {
	return s.fields
}
