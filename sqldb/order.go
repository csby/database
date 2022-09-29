package sqldb

type Order struct {
	Name  string `json:"-" note:"字段"`
	Index int    `json:"index" note:"序号, 越小越靠前"`
	Sort  int    `json:"sort" note:"排序, 0-不排序; 1-按升序; -1-按降序"`
}

type OrderCollection []Order

func (s OrderCollection) Len() int {
	return len(s)
}

func (s OrderCollection) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s OrderCollection) Less(i, j int) bool {
	return s[i].Index < s[j].Index
}
