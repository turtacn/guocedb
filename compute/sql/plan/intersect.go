package plan

import (
	"io"

	"github.com/turtacn/guocedb/compute/sql"
)

// Intersect is a node that returns the intersection of the results of two queries.
type Intersect struct {
	BinaryNode
	Distinct bool
}

// NewIntersect creates a new Intersect node.
func NewIntersect(left, right sql.Node, distinct bool) *Intersect {
	return &Intersect{
		BinaryNode: BinaryNode{Left: left, Right: right},
		Distinct:   distinct,
	}
}

// Schema implements the Node interface.
func (i *Intersect) Schema() sql.Schema {
	return i.Left.Schema()
}

// RowIter implements the Node interface.
func (i *Intersect) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	left, err := i.Left.RowIter(ctx)
	if err != nil {
		return nil, err
	}

	right, err := i.Right.RowIter(ctx)
	if err != nil {
		left.Close()
		return nil, err
	}

	return &intersectIter{
		left:     left,
		right:    right,
		distinct: i.Distinct,
		seenLeft: make(map[uint64]bool),
		rightCounts: nil, // initialized on first call
		ctx:      ctx,
	}, nil
}

// WithChildren implements the Node interface.
func (i *Intersect) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(i, len(children), 2)
	}
	return NewIntersect(children[0], children[1], i.Distinct), nil
}

// TransformUp implements the Transformable interface.
func (i *Intersect) TransformUp(f sql.TransformNodeFunc) (sql.Node, error) {
	left, err := i.Left.TransformUp(f)
	if err != nil {
		return nil, err
	}

	right, err := i.Right.TransformUp(f)
	if err != nil {
		return nil, err
	}

	return f(NewIntersect(left, right, i.Distinct))
}

// TransformExpressionsUp implements the Transformable interface.
func (i *Intersect) TransformExpressionsUp(f sql.TransformExprFunc) (sql.Node, error) {
	left, err := i.Left.TransformExpressionsUp(f)
	if err != nil {
		return nil, err
	}

	right, err := i.Right.TransformExpressionsUp(f)
	if err != nil {
		return nil, err
	}

	return NewIntersect(left, right, i.Distinct), nil
}

func (i *Intersect) String() string {
	p := sql.NewTreePrinter()
	if i.Distinct {
		_ = p.WriteNode("Intersect DISTINCT")
	} else {
		_ = p.WriteNode("Intersect ALL")
	}
	_ = p.WriteChildren(i.Left.String(), i.Right.String())
	return p.String()
}

type intersectIter struct {
	left, right sql.RowIter
	distinct    bool
	seenLeft    map[uint64]bool
	rightCounts map[uint64]int
	ctx         *sql.Context
}

func (i *intersectIter) Next() (sql.Row, error) {
	if i.rightCounts == nil {
		if err := i.loadRight(); err != nil {
			return nil, err
		}
	}

	for {
		row, err := i.left.Next()
		if err != nil {
			return nil, err
		}

		hash, err := hashRow(row)
		if err != nil {
			return nil, err
		}

		count := i.rightCounts[hash]
		if count <= 0 {
			continue
		}

		if i.distinct {
			if i.seenLeft[hash] {
				continue
			}
			i.seenLeft[hash] = true
			return row, nil
		}

		// For Intersect All, we decrement the count
		i.rightCounts[hash]--
		return row, nil
	}
}

func (i *intersectIter) loadRight() error {
	i.rightCounts = make(map[uint64]int)
	for {
		row, err := i.right.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		hash, err := hashRow(row)
		if err != nil {
			return err
		}
		i.rightCounts[hash]++
	}
	return i.right.Close()
}

func (i *intersectIter) Close() error {
	return i.left.Close()
}
