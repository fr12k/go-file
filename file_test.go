package file

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// @export
// TestNew illustrated how to initialise a file and read from it as well as write to it.
func TestNew(t *testing.T) {
	t.Parallel()
	filePath := "./testFile.txt"
	// Clean up the file after the test
	defer os.Remove(filePath)

	// Create a new file
	file := New(filePath)
	assert.NotNil(t, file, "Expected a non-nil file")

	// Test that the file path matches the expected one
	assert.Equal(t, filePath, file.FilePath, "Expected file path to match the input path")

	// Write to the file (create if it not exists)
	n, err := file.Write([]byte("Hello, World!"))
	require.NoError(t, err)
	assert.Equal(t, 13, n)

	// Read from the file
	cnt, err := file.Read()
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", string(cnt))
}

// @export
// TestBufferReader illustrates how to read from a io.Reader.
func TestBufferReader(t *testing.T) {
	t.Parallel()
	file := NewReader(io.NopCloser(strings.NewReader("Hello, World!")))

	content, err := file.Read()
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", string(content))

	err = file.Close()
	require.NoError(t, err)
}

// @export
// TestReadOfANonExistingFile illustrates what happens when you read from an non existing file.
func TestReadOfANonExistingFile(t *testing.T) {
	t.Parallel()
	file := New("nonexistent.txt")

	// Try to read the file
	_, err := file.Read()
	// Error of type ErrNotExist is thrown
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))

	// Close the file doesn't have an effect because its non existing
	err = file.Close()
	require.NoError(t, err)
}

// @export
// TestFileExist illustrates how to check if a file exists.
func TestFileExist(t *testing.T) {
	t.Parallel()
	tmpFile, closeFnc := createFile(t, "Hello, World!")
	defer closeFnc()

	// Create a File instance with a existent file
	file := New(tmpFile)

	// Try to read the file
	exists, err := file.Exists()
	require.NoError(t, err)
	assert.True(t, exists)
}

// @export
// TestNewWriter illustrates how to create a file writer to a non existing/existing directory
// and write to a file. If the directory does not exist, it will be created then.
func TestNewWriter(t *testing.T) {
	t.Parallel()
	baseDir := t.TempDir()
	// Clean up after the tests
	defer os.RemoveAll(baseDir)
	testFilePath := filepath.Join(baseDir, "not_exists", "output.log")

	file := NewWriter(testFilePath)

	// Write to the file
	_, err := file.Write([]byte("Hello, World!"))
	require.NoError(t, err)

	_, err = file.Write([]byte("Hello, World!"))
	require.NoError(t, err)

	// Verify the directory was created
	_, err = os.Stat(file.Writer.Directory)
	require.NoError(t, err)

	// Verify the file was created
	_, err = os.Stat(filepath.Join(file.Writer.Directory, file.Writer.FileName))
	require.NoError(t, err)

	// Read from the same file
	cnr, err := file.Read()
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!Hello, World!", string(cnr))

	// Close the file
	err = file.Close()
	require.NoError(t, err)
}

func TestFileExistFalse(t *testing.T) {
	t.Parallel()
	// Create a File instance with a non-existent file
	file := New("nonexistent.txt")

	// Try to read the file
	exists, err := file.Exists()
	require.NoError(t, err)
	assert.False(t, exists)
}

// @export
// TestReadError illustrates how to setup a file that always returns a defined error.
// This is useful for testing error handling.
func TestReadError(t *testing.T) {
	t.Parallel()
	// Create a File instance with a custom loader that fails
	file := &File{
		FilePath: "fakefile",
		reader: func() (io.Reader, error) {
			return nil, io.EOF // Simulate a load error
		},
	}

	// Attempt to read, expecting an error
	_, err := file.Read()
	require.Error(t, err)
	assert.Equal(t, io.EOF, err)
}

func TestFileExistError(t *testing.T) {
	t.Parallel()
	// Create a File instance that always returns an error
	file := NewReaderError(os.ErrClosed)

	// Try to read the file
	exists, err := file.Exists()
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

		file := NewWriter(testFilePath)
		writer, err := file.writer()()
		require.NoError(t, err)

		assert.Equal(t, filepath.Dir(testFilePath), writer.Directory)
		assert.Equal(t, filepath.Base(testFilePath), writer.FileName)

		// Verify the file was created
		_, err = os.Stat(filepath.Join(writer.Directory, writer.FileName))
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

		file := NewWriter(dir + "/")
		_, err = file.writer()()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create directory")
	})

	t.Run("FailsToCreateFile", func(t *testing.T) {
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
	file := NewWriterBuffer(&buf, testFilePath)
	writer, err := file.writer()()
	require.NotNil(t, writer)
	require.NoError(t, err)

	assert.Equal(t, filepath.Dir(testFilePath), writer.Directory)
	assert.Equal(t, filepath.Base(testFilePath), writer.FileName)
}

func TestNewWriterError(t *testing.T) {
	t.Parallel()
	file := NewWriterError(os.ErrClosed)
	_, err := file.writer()()
	assert.Error(t, err)

	file = NewWriterError(nil)
	n, err := file.Write([]byte("Hello, World!"))
	assert.ErrorContains(t, err, "unexpected Writer is nil")
	assert.Equal(t, -1, n)
}

func TestClose(t *testing.T) {
	t.Parallel()
	file := File{
		Reader: ErrReaderCloser{iotest.ErrReader(os.ErrClosed)},
		Writer: &Writer{Writer: ErrWriterCloser{ErrWriter{os.ErrDeadlineExceeded}}},
	}
	err := file.Close()
	assert.ErrorIs(t, err, os.ErrClosed)
	assert.ErrorIs(t, err, os.ErrDeadlineExceeded)
}

func TestCloseWriter(t *testing.T) {
	t.Parallel()
	file := File{
		Writer: &Writer{Writer: ErrWriterCloser{ErrWriter{os.ErrDeadlineExceeded}}},
	}
	err := file.Close()
	assert.ErrorIs(t, err, os.ErrDeadlineExceeded)
}

func TestCloseReader(t *testing.T) {
	t.Parallel()
	file := File{
		Reader: ErrReaderCloser{iotest.ErrReader(os.ErrClosed)},
	}
	err := file.Close()
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
