package badger

import (
	"io"

	"github.com/turtacn/guocedb/compute/sql"
)

// Partition implements sql.Partition.
type Partition struct {
	key []byte
}

// Key returns the partition key.
func (p *Partition) Key() []byte {
	return p.key
}

// partitionIter implements sql.PartitionIter.
type partitionIter struct {
	partitions []*Partition
	idx        int
}

// Next returns the next partition.
func (p *partitionIter) Next() (sql.Partition, error) {
	if p.idx >= len(p.partitions) {
		return nil, io.EOF
	}
	part := p.partitions[p.idx]
	p.idx++
	return part, nil
}

// Close closes the iterator.
func (p *partitionIter) Close() error {
	return nil
}
