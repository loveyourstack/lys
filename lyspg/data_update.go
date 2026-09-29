package lyspg

import "github.com/loveyourstack/lys/lystype"

// experimental

var (
	DataUpdateViewSuffix = "_data_update"
	DataUpdateColTags    = []string{"data_update_id", "affected_id", "affected_at", "affected_by", "affected_old_values", "affected_new_values"}
)

// DataUpdateCols are the fields expected in a data update table
type DataUpdateCols struct {
	DataUpdateId      int64            `db:"data_update_id" json:"data_update_id,omitzero"`
	AffectedId        int64            `db:"affected_id" json:"affected_id,omitzero"`
	AffectedAt        lystype.Datetime `db:"affected_at" json:"affected_at,omitzero"`
	AffectedBy        string           `db:"affected_by" json:"affected_by,omitzero"`
	AffectedOldValues string           `db:"affected_old_values" json:"affected_old_values,omitzero"`
	AffectedNewValues string           `db:"affected_new_values" json:"affected_new_values,omitzero"`
}
