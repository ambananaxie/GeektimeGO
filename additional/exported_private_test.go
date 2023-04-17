package additional

import (
	_ "database/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	_ "unsafe"
)
//go:linkname convertAssign database/sql.convertAssign
func convertAssign(dest, src any) error

func TestConvertAssign(t *testing.T) {
	var result int
	err := convertAssign(&result, "123")
	require.NoError(t, err)
	assert.Equal(t, 123, result)
}