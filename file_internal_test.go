package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWriteFailsToCreateFile(t *testing.T) {
	t.Parallel()
	// Create a temporary directory
	baseDir, err := os.MkdirTemp("", "readonly-test")
	assert.NoError(t, err)
	defer os.RemoveAll(baseDir)

	testFilePath := filepath.Join(baseDir, "output.log")

	file := NewWriter(testFilePath)
	fnc := file.writer()

	os.RemoveAll(baseDir)
	_, err = fnc()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create file")

	_, err = file.Write([]byte("Hello, World!"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create file")
}
