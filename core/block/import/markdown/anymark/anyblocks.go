package anymark

import (
	"strings"

	"github.com/anytypeio/go-anytype-middleware/pkg/lib/pb/model"
)

func preprocessBlocks(blocks []*model.Block) (blocksOut []*model.Block) {

	blocksOut = []*model.Block{}
	accum := []*model.Block{}

	for _, b := range blocks {
		if t := b.GetText(); t != nil && t.Style == model.BlockContentText_Code {
			accum = append(accum, b)
		} else {
			if len(accum) > 0 {
				blocksOut = append(blocksOut, combineCodeBlocks(accum))
				accum = []*model.Block{}
			}

			blocksOut = append(blocksOut, b)
		}

	}

	if len(accum) > 0 {
		blocksOut = append(blocksOut, combineCodeBlocks(accum))
	}

	for _, b := range blocks {
		for i, cId := range b.ChildrenIds {
			if len(cId) == 0 {
				b.ChildrenIds = append(b.ChildrenIds[:i], b.ChildrenIds[i+1:]...)
			}
		}
	}

	return blocksOut
}

func combineCodeBlocks(accum []*model.Block) (res *model.Block) {
	var textArr []string

	for _, b := range accum {
		if b.GetText() != nil {
			textArr = append(textArr, b.GetText().Text)
		}
	}

	res = &model.Block{
		Content: &model.BlockContentOfText{
			Text: &model.BlockContentText{
				Text:  strings.Join(textArr, "\n"),
				Style: model.BlockContentText_Code,
				Marks: &model.BlockContentTextMarks{
					Marks: []*model.BlockContentTextMark{},
				},
			},
		},
	}

	return res
}