package parser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/guocedb/common/errors"
	"github.com/turtacn/guocedb/common/constants"
)

func TestParse_Select(t *testing.T) {
	p := NewParser()
	ctx := context.Background()

	sql := "SELECT 1"
	stmt, err := p.Parse(ctx, sql)
	assert.NoError(t, err)
	assert.NotNil(t, stmt)
}

func TestParse_Error(t *testing.T) {
	p := NewParser()
	ctx := context.Background()

	sql := "SELECT FROM" // Invalid SQL
	stmt, err := p.Parse(ctx, sql)
	assert.Error(t, err)
	assert.Nil(t, stmt)

	// Check if it is a common error
	assert.IsType(t, &errors.Error{}, err)
	e, ok := err.(*errors.Error)
	assert.True(t, ok)
	assert.Equal(t, constants.ErrCodeSyntax, e.Code)
}
