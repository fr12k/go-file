Generated TESTS.md
# file Package

The `file` package provides a high-level abstraction for reading and writing files in Go. 
It offers an easy-to-use API for handling file operations with built-in support for lazy initialization and concurrency safety.

## Features
- **Lazy Initialization**: File readers and writers are initialized only when needed.
- **Concurrency-Safe**: Uses `sync.OnceValues` to ensure resources are initialized only once.
- **Unified API**: Provides a structured interface for handling file I/O.
- **Error Handling**: Includes constructors to create files with predefined errors.
- **Supports Buffers**: Allows reading/writing from/to an in-memory buffer instead of a file.

## Installation
To use this package, simply import it in your Go project:

```go
import "github.com/fr12k/go-file"
```

## Example Usage

# Go Test Cases

## file_test.go


### TestNew

TestNew illustrated how to initialise a file and read from it as well as write to it.

```go
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
```

### TestBufferReader

TestBufferReader illustrates how to read from a io.Reader.

```go
func TestBufferReader(t *testing.T) {
	t.Parallel()
	file := NewReader(io.NopCloser(strings.NewReader("Hello, World!")))

	content, err := file.Read()
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", string(content))

	err = file.Close()
	require.NoError(t, err)
}
```

### TestReadOfANonExistingFile

TestReadOfANonExistingFile illustrates what happens when you read from an non existing file.

```go
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
```

### TestFileExist

TestFileExist illustrates how to check if a file exists.

```go
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
```

### TestNewWriter

TestNewWriter illustrates how to create a file writer to a non existing/existing directory
and write to a file. If the directory does not exist, it will be created then.

```go
func TestNewWriter(t *testing.T) {
	t.Parallel()
	baseDir := t.TempDir()
	// Clean up after the tests
	defer os.RemoveAll(baseDir)
	testFilePath := filepath.Join(baseDir, "not_exists", "output.log")
	file := NewWriter(testFilePath)

	writer, err := file.writer()()
	require.NoError(t, err)

	assert.Equal(t, filepath.Dir(testFilePath), writer.Directory)
	assert.Equal(t, filepath.Base(testFilePath), writer.FileName)

	_, err = file.Write([]byte("Hello, World!"))
	require.NoError(t, err)

	_, err = file.Write([]byte("Hello, World!"))
	require.NoError(t, err)

	// Verify the directory was created
	_, err = os.Stat(writer.Directory)
	require.NoError(t, err)

	// Verify the file was created
	_, err = os.Stat(filepath.Join(writer.Directory, writer.FileName))
	require.NoError(t, err)

	cnr, err := file.Read()
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!Hello, World!", string(cnr))

	// Test Close
	err = file.Close()
	require.NoError(t, err)
}
```

### TestReadError

TestReadError illustrates how to setup a file that always returns a defined error.
This is useful for testing error handling.

```go
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
```


## API Reference

### Structs
#### `File`
Represents a file abstraction for reading and writing.

- `Exists() (bool, error)`: Checks if the file exists.
- `Read() ([]byte, error)`: Reads the entire file.
- `Write(p []byte) (int, error)`: Writes to the file.
- `Close() error`: Closes the file.

## Contributing
Contributions are welcome! Please submit either an issue or a pull request with improvements or fixes.

## License

`go-mask` is licensed under the Apache 2 License. See the [LICENSE](LICENSE) file for more details.

