package lyspg

import "github.com/loveyourstack/lys/lystype"

// experimental

var (
	DataUpdateViewSuffix = "_data_update"
	DataUpdateColTags    = []string{"data_update_id", "affected_id", "affected_at", "affected_by", "affected_old_values", "affected_new_values"}
)

// DataUpdateCols are the fields expected in a data update table
type DataUpdateCols struct {
	DataUpdateId      int64            `db:"data_update_id" json:"data_update_id,omitempty"`
	AffectedId        int64            `db:"affected_id" json:"affected_id,omitempty"`
	AffectedAt        lystype.Datetime `db:"affected_at" json:"affected_at,omitempty"`
	AffectedBy        string           `db:"affected_by" json:"affected_by,omitempty"`
	AffectedOldValues string           `db:"affected_old_values" json:"affected_old_values,omitempty"`
	AffectedNewValues string           `db:"affected_new_values" json:"affected_new_values,omitempty"`
}
