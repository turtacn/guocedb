package plan

import (
	"io"

	"github.com/turtacn/guocedb/compute/sql"
)

// Except is a node that returns the rows from the left query that are not in the right query.
type Except struct {
	BinaryNode
	Distinct bool
}

// NewExcept creates a new Except node.
func NewExcept(left, right sql.Node, distinct bool) *Except {
	return &Except{
		BinaryNode: BinaryNode{Left: left, Right: right},
		Distinct:   distinct,
	}
}

// Schema implements the Node interface.
func (e *Except) Schema() sql.Schema {
	return e.Left.Schema()
}

// RowIter implements the Node interface.
func (e *Except) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	left, err := e.Left.RowIter(ctx)
	if err != nil {
		return nil, err
	}

	right, err := e.Right.RowIter(ctx)
	if err != nil {
		left.Close()
		return nil, err
	}

	return &exceptIter{
		left:     left,
		right:    right,
		distinct: e.Distinct,
		seenLeft: make(map[uint64]bool),
		inRight:  nil, // initialized on first call
		ctx:      ctx,
	}, nil
}

// WithChildren implements the Node interface.
func (e *Except) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(e, len(children), 2)
	}
	return NewExcept(children[0], children[1], e.Distinct), nil
}

// TransformUp implements the Transformable interface.
func (e *Except) TransformUp(f sql.TransformNodeFunc) (sql.Node, error) {
	left, err := e.Left.TransformUp(f)
	if err != nil {
		return nil, err
	}

	right, err := e.Right.TransformUp(f)
	if err != nil {
		return nil, err
	}

	return f(NewExcept(left, right, e.Distinct))
}

// TransformExpressionsUp implements the Transformable interface.
func (e *Except) TransformExpressionsUp(f sql.TransformExprFunc) (sql.Node, error) {
	left, err := e.Left.TransformExpressionsUp(f)
	if err != nil {
		return nil, err
	}

	right, err := e.Right.TransformExpressionsUp(f)
	if err != nil {
		return nil, err
	}

	return NewExcept(left, right, e.Distinct), nil
}

func (e *Except) String() string {
	p := sql.NewTreePrinter()
	if e.Distinct {
		_ = p.WriteNode("Except DISTINCT")
	} else {
		_ = p.WriteNode("Except ALL")
	}
	_ = p.WriteChildren(e.Left.String(), e.Right.String())
	return p.String()
}

type exceptIter struct {
	left, right sql.RowIter
	distinct    bool
	seenLeft    map[uint64]bool
	inRight     map[uint64]bool
	ctx         *sql.Context
}

func (i *exceptIter) Next() (sql.Row, error) {
	if i.inRight == nil {
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

		// Row must NOT be in right side
		if i.inRight[hash] {
			continue
		}

		// If distinct, check if we already emitted this row
		if i.distinct {
			if i.seenLeft[hash] {
				continue
			}
			i.seenLeft[hash] = true
		}

		return row, nil
	}
}

func (i *exceptIter) loadRight() error {
	i.inRight = make(map[uint64]bool)
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
		i.inRight[hash] = true
	}
	return i.right.Close()
}

func (i *exceptIter) Close() error {
	return i.left.Close()
}
