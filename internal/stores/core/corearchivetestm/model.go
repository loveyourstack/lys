package corearchivetestm

import (
	"github.com/google/uuid"
	"github.com/loveyourstack/lys/lystype"
)

type Input struct {
	CInt  *int64  `db:"c_int" json:"c_int,omitempty"`
	CText *string `db:"c_text" json:"c_text,omitempty"`
}

type Model struct {
	Id      int64            `db:"id" json:"id,omitempty"`
	Iduu    uuid.UUID        `db:"id_uu" json:"id_uu,omitempty"`
	EntryAt lystype.Datetime `db:"entry_at" json:"entry_at,omitempty"`
	Input
}
