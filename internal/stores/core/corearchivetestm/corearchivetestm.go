package corearchivetestm

import (
	"github.com/loveyourstack/lys/lystype"
)

type Input struct {
	CInt  *int64  `db:"c_int" json:"c_int,omitzero"`
	CText *string `db:"c_text" json:"c_text,omitzero"`
}

type Model struct {
	Id        int64            `db:"id" json:"id,omitzero"`
	CreatedAt lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
	Input
}
