package filter

import "github.com/anytypeio/go-anytype-middleware/pkg/lib/pb/model"

type Order interface {
	Compare(a, b Getter) int
}

type SetOrder []Order

func (so SetOrder) Compare(a, b Getter) int {
	for _, o := range so {
		if comp := o.Compare(a, b); comp != 0 {
			return comp
		}
	}
	return 0
}

type KeyOrder struct {
	Key  string
	Type model.BlockContentDataviewSortType
}

func (ko KeyOrder) Compare(a, b Getter) int {
	av := a.Get(ko.Key)
	bv := b.Get(ko.Key)
	comp := av.Compare(bv)
	if ko.Type == model.BlockContentDataviewSort_Desc {
		comp = -comp
	}
	return comp
}