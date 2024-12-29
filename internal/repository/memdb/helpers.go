package memdb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// SetupTestRepository Helper function to create a new Repository with a temporary data file.
func SetupTestRepository(t *testing.T) *Repository {
	t.Helper()

	tempDir := t.TempDir()
	dataFilePath := fmt.Sprintf("%s/%s", tempDir, COUPONS_DEFAULT_DATA_PATH)

	repo, err := NewRepository(dataFilePath)
	assert.NoError(t, err, "Failed to create repository")
	return repo
}
