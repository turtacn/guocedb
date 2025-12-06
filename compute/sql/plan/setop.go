package plan

import (
	"fmt"

	"github.com/mitchellh/hashstructure"
	"github.com/turtacn/guocedb/compute/sql"
)

// hashRow computes a hash for a given row.
func hashRow(row sql.Row) (uint64, error) {
	h, err := hashstructure.Hash(row, nil)
	if err != nil {
		return 0, fmt.Errorf("unable to hash row: %s", err)
	}
	return h, nil
}
