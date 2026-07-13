package lysextdata

import (
	"context"

	"github.com/loveyourstack/lys/lystype"
)

// SyncKey identifies the source and type of external data being synced, e.g. "EcbCurrencies".
type SyncKey string

// ISyncStore is a store that manages the last sync timestamp for external data.
type ISyncStore interface {
	SelectLastSyncAt(ctx context.Context, syncKey SyncKey) (lastSyncAt lystype.Datetime, err error)
	Upsert(ctx context.Context, syncKey SyncKey) error
}
