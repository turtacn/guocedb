package plan

import (
	"io"

	"github.com/turtacn/guocedb/compute/sql"
)

// Union is a node that returns the union of the results of two queries.
type Union struct {
	BinaryNode
	Distinct bool
}

// NewUnion creates a new Union node.
func NewUnion(left, right sql.Node, distinct bool) *Union {
	return &Union{
		BinaryNode: BinaryNode{Left: left, Right: right},
		Distinct:   distinct,
	}
}

// Schema implements the Node interface.
func (u *Union) Schema() sql.Schema {
	return u.Left.Schema()
}

// RowIter implements the Node interface.
func (u *Union) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	left, err := u.Left.RowIter(ctx)
	if err != nil {
		return nil, err
	}

	right, err := u.Right.RowIter(ctx)
	if err != nil {
		left.Close()
		return nil, err
	}

	return &unionIter{
		left:     left,
		right:    right,
		distinct: u.Distinct,
		seen:     make(map[uint64]bool),
		ctx:      ctx,
	}, nil
}

// WithChildren implements the Node interface.
func (u *Union) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(u, len(children), 2)
	}
	return NewUnion(children[0], children[1], u.Distinct), nil
}

// TransformUp implements the Transformable interface.
func (u *Union) TransformUp(f sql.TransformNodeFunc) (sql.Node, error) {
	left, err := u.Left.TransformUp(f)
	if err != nil {
		return nil, err
	}

	right, err := u.Right.TransformUp(f)
	if err != nil {
		return nil, err
	}

	return f(NewUnion(left, right, u.Distinct))
}

// TransformExpressionsUp implements the Transformable interface.
func (u *Union) TransformExpressionsUp(f sql.TransformExprFunc) (sql.Node, error) {
	left, err := u.Left.TransformExpressionsUp(f)
	if err != nil {
		return nil, err
	}

	right, err := u.Right.TransformExpressionsUp(f)
	if err != nil {
		return nil, err
	}

	return NewUnion(left, right, u.Distinct), nil
}

func (u *Union) String() string {
	p := sql.NewTreePrinter()
	if u.Distinct {
		_ = p.WriteNode("Union DISTINCT")
	} else {
		_ = p.WriteNode("Union ALL")
	}
	_ = p.WriteChildren(u.Left.String(), u.Right.String())
	return p.String()
}

type unionIter struct {
	left, right sql.RowIter
	distinct    bool
	seen        map[uint64]bool
	stage       int // 0: left, 1: right, 2: done
	ctx         *sql.Context
}

func (i *unionIter) Next() (sql.Row, error) {
	for {
		if i.stage == 0 {
			row, err := i.left.Next()
			if err == io.EOF {
				i.stage = 1
				continue
			}
			if err != nil {
				return nil, err
			}

			if i.distinct {
				hash, err := hashRow(row)
				if err != nil {
					return nil, err
				}
				if i.seen[hash] {
					continue
				}
				i.seen[hash] = true
			}
			return row, nil
		} else if i.stage == 1 {
			row, err := i.right.Next()
			if err == io.EOF {
				i.stage = 2
				return nil, io.EOF
			}
			if err != nil {
				return nil, err
			}

			if i.distinct {
				hash, err := hashRow(row)
				if err != nil {
					return nil, err
				}
				if i.seen[hash] {
					continue
				}
				i.seen[hash] = true
			}
			return row, nil
		} else {
			return nil, io.EOF
		}
	}
}

func (i *unionIter) Close() error {
	err1 := i.left.Close()
	err2 := i.right.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
