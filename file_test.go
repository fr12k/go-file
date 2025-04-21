package file_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/fr12k/go-file"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// @markdown
// TestNew illustrated how to initialise a file and read from it as well as write to it.
func TestNew(t *testing.T) {
	t.Parallel()
	filePath := "./testFile.txt"
	// Clean up the file after the test
	defer os.Remove(filePath)

	// Create a new file
	f := file.New(filePath)
	assert.NotNil(t, f, "Expected a non-nil file")

	// Test that the file path matches the expected one
	assert.Equal(t, filePath, f.FilePath, "Expected file path to match the input path")

	// Write to the file (create if it not exists)
	n, err := f.Write([]byte("Hello, World!"))
	require.NoError(t, err)
	assert.Equal(t, 13, n)

	// Read from the file
	cnt, err := f.Read()
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", string(cnt))
}

func TestOpen(t *testing.T) {
	t.Parallel()
	filePath := t.Name() + ".txt"
	// Clean up the file after the test
	defer os.Remove(filePath)

	// Create a new file
	f := file.Open()(filePath)
	f2 := file.OpenFile(f)(filePath)
	assert.NotNil(t, f2, "Expected a non-nil file")

	// Test that the file path matches the expected one
	assert.Equal(t, filePath, f2.FilePath, "Expected file path to match the input path")

	// Write to the file (create if it not exists)
	n, err := f2.Write([]byte("Hello, World!"))
	require.NoError(t, err)
	assert.Equal(t, 13, n)

	// Read from the file
	cnt, err := f2.Read()
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", string(cnt))
}

// @markdown
// TestBufferReader illustrates how to read from a io.Reader.
func TestBufferReader(t *testing.T) {
	t.Parallel()
	f := file.NewReader(io.NopCloser(strings.NewReader("Hello, World!")))

	content, err := f.Read()
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", string(content))

	err = f.Close()
	require.NoError(t, err)
}

// @markdown
// TestReadOfANonExistingFile illustrates what happens when you read from an non existing file.
func TestReadOfANonExistingFile(t *testing.T) {
	t.Parallel()
	f := file.New("nonexistent.txt")

	// Try to read the file
	_, err := f.Read()
	// Error of type ErrNotExist is thrown
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))

	// Close the file doesn't have an effect because its non existing
	err = f.Close()
	require.NoError(t, err)
}

// @markdown
// TestFileExist illustrates how to check if a file exists.
func TestFileExist(t *testing.T) {
	t.Parallel()
	tmpFile, closeFnc := createFile(t, "Hello, World!")
	defer closeFnc()

	// Create a File instance with a existent file
	f := file.New(tmpFile)

	// Try to read the file
	exists, err := f.Exists()
	require.NoError(t, err)
	assert.True(t, exists)
}

// @markdown
// TestNewWriter illustrates how to create a file writer to a non existing/existing directory
// and write to a file. If the directory does not exist, it will be created then.
func TestNewWriter(t *testing.T) {
	t.Parallel()
	baseDir := t.TempDir()
	// Clean up after the tests
	defer os.RemoveAll(baseDir)
	testFilePath := filepath.Join(baseDir, "not_exists", "output.log")

	f := file.NewWriter(testFilePath)

	// Write to the file
	_, err := f.Write([]byte("Hello, World!"))
	require.NoError(t, err)

	_, err = f.Write([]byte("Hello, World!"))
	require.NoError(t, err)

	// Verify the directory was created
	_, err = os.Stat(f.Writer.Directory)
	require.NoError(t, err)

	// Verify the file was created
	_, err = os.Stat(filepath.Join(f.Writer.Directory, f.Writer.FileName))
	require.NoError(t, err)

	// Read from the same file
	cnr, err := f.Read()
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!Hello, World!", string(cnr))

	// Close the file
	err = f.Close()
	require.NoError(t, err)
}

func TestFileExistFalse(t *testing.T) {
	t.Parallel()
	// Create a File instance with a non-existent file
	f := file.New("nonexistent.txt")

	// Try to read the file
	exists, err := f.Exists()
	require.NoError(t, err)
	assert.False(t, exists)
}

// @markdown
// TestReadError illustrates how to setup a file that always returns a defined error.
// This is useful for testing error handling.
func TestReadError(t *testing.T) {
	t.Parallel()
	// Create a File instance with a custom loader that fails
	f := file.NewReaderError(io.EOF)

	// Attempt to read, expecting an error
	_, err := f.Read()
	require.Error(t, err)
	assert.Equal(t, io.EOF, err)
}

func TestFileExistError(t *testing.T) {
	t.Parallel()
	// Create a File instance that always returns an error
	f := file.NewReaderError(os.ErrClosed)

	// Try to read the file
	exists, err := f.Exists()
	assert.Error(t, err)
	assert.False(t, exists)
}

func TestNewWriterComplex(t *testing.T) {
	t.Parallel()
	// Test directory structure
	baseDir := "testdata"

	t.Run("CreatesWriterWhenDirectoryExists", func(t *testing.T) {
		t.Parallel()
		// Clean up after the tests
		defer os.RemoveAll(baseDir)
		testFilePath := filepath.Join(baseDir, "logs-exists", "output.log")
		// Ensure the directory exists
		err := os.MkdirAll(filepath.Dir(testFilePath), os.ModePerm)
		require.NoError(t, err)

		f := file.NewWriter(testFilePath)
		n, err := f.Write([]byte{})
		require.NoError(t, err)
		assert.Equal(t, 0, n)

		assert.Equal(t, filepath.Dir(testFilePath), f.Writer.Directory)
		assert.Equal(t, filepath.Base(testFilePath), f.Writer.FileName)

		// Verify the file was created
		_, err = os.Stat(filepath.Join(f.Writer.Directory, f.Writer.FileName))
		require.NoError(t, err)
	})

	t.Run("FailsToCreateDirectory", func(t *testing.T) {
		t.Parallel()
		// Clean up after the tests
		defer os.RemoveAll(baseDir)
		// Create a file at the directory path to cause MkdirAll to fail
		err := os.MkdirAll(baseDir, os.ModePerm)
		require.NoError(t, err)
		dir := filepath.Join(baseDir, "logs")
		err = os.WriteFile(dir, []byte{}, 0o600) // Create a file where the directory should be
		require.NoError(t, err)
		defer os.Remove(dir)

		f := file.NewWriter(dir + "/")
		_, err = f.Write([]byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create directory")
	})
}

func TestNewWriterBuffer(t *testing.T) {
	t.Parallel()
	// Test directory structure
	baseDir := "."
	testFilePath := filepath.Join(baseDir, "output.log")

	// Clean up after the tests
	defer os.RemoveAll(baseDir)
	var buf bytes.Buffer
	f := file.NewWriterBuffer(&buf, testFilePath)
	_, err := f.Write([]byte{})
	require.NoError(t, err)

	assert.Equal(t, filepath.Dir(testFilePath), f.Writer.Directory)
	assert.Equal(t, filepath.Base(testFilePath), f.Writer.FileName)
}

func TestNewWriterError(t *testing.T) {
	t.Parallel()
	f := file.NewWriterError(os.ErrClosed)
	_, err := f.Write([]byte{})
	assert.Error(t, err)

	f = file.NewWriterError(nil)
	n, err := f.Write([]byte("Hello, World!"))
	assert.ErrorContains(t, err, "unexpected Writer is nil")
	assert.Equal(t, -1, n)
}

func TestClose(t *testing.T) {
	t.Parallel()
	f := file.File{
		Reader: ErrReaderCloser{iotest.ErrReader(os.ErrClosed)},
		Writer: &file.Writer{Writer: ErrWriterCloser{ErrWriter{os.ErrDeadlineExceeded}}},
	}
	err := f.Close()
	assert.ErrorIs(t, err, os.ErrClosed)
	assert.ErrorIs(t, err, os.ErrDeadlineExceeded)
}

func TestCloseWriter(t *testing.T) {
	t.Parallel()
	f := file.File{
		Writer: &file.Writer{Writer: ErrWriterCloser{ErrWriter{os.ErrDeadlineExceeded}}},
	}
	err := f.Close()
	assert.ErrorIs(t, err, os.ErrDeadlineExceeded)
}

func TestCloseReader(t *testing.T) {
	t.Parallel()
	f := file.File{
		Reader: ErrReaderCloser{iotest.ErrReader(os.ErrClosed)},
	}
	err := f.Close()
	assert.ErrorIs(t, err, os.ErrClosed)
}

// Test Utility

type ErrWriterCloser struct {
	io.Writer
}

func (e ErrWriterCloser) Close() error {
	_, err := e.Write(nil)
	return err
}

type ErrWriter struct {
	err error
}

func (e ErrWriter) Write(_ []byte) (n int, err error) {
	return -1, e.err
}

type ErrReaderCloser struct {
	io.Reader
}

func (e ErrReaderCloser) Close() error {
	_, err := e.Read(nil)
	return err
}

func createFile(t *testing.T, cnt string) (name string, clean func()) {
	t.Helper()
	// Create a temporary file with test content
	tmpFile, err := os.CreateTemp("", "testfile")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(cnt)
	require.NoError(t, err)

	require.NoError(t, tmpFile.Close())
	return tmpFile.Name(), func() {
		os.Remove(tmpFile.Name())
	}
}
