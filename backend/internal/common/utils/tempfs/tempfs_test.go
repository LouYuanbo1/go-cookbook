package tempfs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTempFs(t *testing.T) TempFs {
	t.Helper()
	return NewTempFs(1*time.Hour, 1*time.Hour)
}

func createTempFile(t *testing.T, dir string, content string) string {
	t.Helper()
	file, err := os.CreateTemp(dir, "test-*.txt")
	require.NoError(t, err)
	_, err = file.WriteString(content)
	require.NoError(t, err)
	file.Close()
	return file.Name()
}

func TestNewTempFs(t *testing.T) {
	ts := NewTempFs(1*time.Hour, 1*time.Hour)
	assert.NotNil(t, ts)
	ts.Close()
}

func TestRegisterTempFile_Success(t *testing.T) {
	ts := newTestTempFs(t)
	defer ts.Close()

	dir := t.TempDir()
	filePath := createTempFile(t, dir, "test content")

	id, err := ts.RegisterTempFile(filePath)
	require.NoError(t, err)
	assert.NotEmpty(t, id)
}

func TestRegisterTempFile_NonExistentFile(t *testing.T) {
	ts := newTestTempFs(t)
	defer ts.Close()

	_, err := ts.RegisterTempFile("/nonexistent/path/file.txt")
	assert.Error(t, err)
}

func TestRegisterTempFile_UniqueIDs(t *testing.T) {
	ts := newTestTempFs(t)
	defer ts.Close()

	dir := t.TempDir()
	file1 := createTempFile(t, dir, "content1")
	file2 := createTempFile(t, dir, "content2")

	id1, err := ts.RegisterTempFile(file1)
	require.NoError(t, err)
	id2, err := ts.RegisterTempFile(file2)
	require.NoError(t, err)

	assert.NotEqual(t, id1, id2)
}

func TestGetTempFile_Success(t *testing.T) {
	ts := newTestTempFs(t)
	defer ts.Close()

	dir := t.TempDir()
	filePath := createTempFile(t, dir, "test content")

	id, err := ts.RegisterTempFile(filePath)
	require.NoError(t, err)

	info, err := ts.GetTempFile(id)
	require.NoError(t, err)
	assert.Equal(t, id, info.ID)
	assert.Equal(t, filePath, info.Path)
	assert.Equal(t, int64(12), info.Size) // len("test content")
	assert.False(t, info.CreatedAt.IsZero())
	assert.True(t, info.ExpiresAt.After(info.CreatedAt))
}

func TestGetTempFile_NotFound(t *testing.T) {
	ts := newTestTempFs(t)
	defer ts.Close()

	_, err := ts.GetTempFile("nonexistent-id")
	assert.Error(t, err)
}

func TestGetTempFile_Expired(t *testing.T) {
	// Create a tempfs with 0 expiration (immediately expired)
	ts := NewTempFs(0, 1*time.Hour)
	defer ts.Close()

	dir := t.TempDir()
	filePath := createTempFile(t, dir, "test")

	id, err := ts.RegisterTempFile(filePath)
	require.NoError(t, err)

	// Wait a tiny bit for expiration
	time.Sleep(10 * time.Millisecond)

	_, err = ts.GetTempFile(id)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "过期")
}

func TestDeleteTempFile_Success(t *testing.T) {
	ts := newTestTempFs(t)
	defer ts.Close()

	dir := t.TempDir()
	filePath := createTempFile(t, dir, "test content")

	id, err := ts.RegisterTempFile(filePath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	err = ts.DeleteTempFile(id)
	assert.NoError(t, err)

	// Verify file is deleted
	_, err = os.Stat(filePath)
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestDeleteTempFile_NotFound(t *testing.T) {
	ts := newTestTempFs(t)
	defer ts.Close()

	err := ts.DeleteTempFile("nonexistent-id")
	assert.NoError(t, err) // Should return nil (already removed)
}

func TestPromoteTempFile_Success(t *testing.T) {
	ts := newTestTempFs(t)
	defer ts.Close()

	dir := t.TempDir()
	filePath := createTempFile(t, dir, "test content")

	id, err := ts.RegisterTempFile(filePath)
	require.NoError(t, err)

	destDir := t.TempDir()
	err = ts.PromoteTempFile(id, destDir, "promoted-file.txt")
	assert.NoError(t, err)

	// Verify the file was moved
	destPath := filepath.Join(destDir, "promoted-file.txt")
	_, err = os.Stat(destPath)
	assert.NoError(t, err)

	// Verify original file no longer exists
	_, err = os.Stat(filePath)
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestPromoteTempFile_NotFound(t *testing.T) {
	ts := newTestTempFs(t)
	defer ts.Close()

	err := ts.PromoteTempFile("nonexistent-id", t.TempDir(), "new-file.txt")
	assert.Error(t, err)
}

func TestPromoteTempFile_Expired(t *testing.T) {
	ts := NewTempFs(0, 1*time.Hour)
	defer ts.Close()

	dir := t.TempDir()
	filePath := createTempFile(t, dir, "test")

	id, err := ts.RegisterTempFile(filePath)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	err = ts.PromoteTempFile(id, t.TempDir(), "new-file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "过期")
}

func TestPromoteTempFile_CreatesDestDir(t *testing.T) {
	ts := newTestTempFs(t)
	defer ts.Close()

	dir := t.TempDir()
	filePath := createTempFile(t, dir, "test content")

	id, err := ts.RegisterTempFile(filePath)
	require.NoError(t, err)

	// Use a nested directory that doesn't exist yet
	destDir := filepath.Join(t.TempDir(), "nested", "dir")
	err = ts.PromoteTempFile(id, destDir, "promoted.txt")
	assert.NoError(t, err)

	// Verify file exists in nested directory
	_, err = os.Stat(filepath.Join(destDir, "promoted.txt"))
	assert.NoError(t, err)
}

func TestDoublePromote_Fails(t *testing.T) {
	ts := newTestTempFs(t)
	defer ts.Close()

	dir := t.TempDir()
	filePath := createTempFile(t, dir, "test")

	id, err := ts.RegisterTempFile(filePath)
	require.NoError(t, err)

	// First promote should succeed
	err = ts.PromoteTempFile(id, t.TempDir(), "file1.txt")
	assert.NoError(t, err)

	// Second promote should fail (ID already removed)
	err = ts.PromoteTempFile(id, t.TempDir(), "file2.txt")
	assert.Error(t, err)
}

func TestRegisterTempFile_MultipleFiles(t *testing.T) {
	ts := newTestTempFs(t)
	defer ts.Close()

	dir := t.TempDir()
	ids := make([]string, 5)
	for i := 0; i < 5; i++ {
		filePath := createTempFile(t, dir, "content")
		id, err := ts.RegisterTempFile(filePath)
		require.NoError(t, err)
		ids[i] = id
	}

	// Verify all IDs are unique
	seen := make(map[string]bool)
	for _, id := range ids {
		assert.False(t, seen[id], "duplicate ID: %s", id)
		seen[id] = true
	}

	// Verify all can be retrieved
	for _, id := range ids {
		_, err := ts.GetTempFile(id)
		assert.NoError(t, err)
	}
}

func TestCleanupExpired(t *testing.T) {
	// We test the cleanup loop indirectly by using a short expiration
	ts := NewTempFs(50*time.Millisecond, 100*time.Millisecond)
	defer ts.Close()

	dir := t.TempDir()
	filePath := createTempFile(t, dir, "test content")

	id, err := ts.RegisterTempFile(filePath)
	require.NoError(t, err)

	// Wait for expiration and cleanup
	time.Sleep(200 * time.Millisecond)

	// Should be cleaned up
	_, err = ts.GetTempFile(id)
	assert.Error(t, err)
}

func TestTempFileInfo(t *testing.T) {
	ts := newTestTempFs(t)
	defer ts.Close()

	dir := t.TempDir()
	filePath := createTempFile(t, dir, "test content")

	id, err := ts.RegisterTempFile(filePath)
	require.NoError(t, err)

	info, err := ts.GetTempFile(id)
	require.NoError(t, err)

	assert.Equal(t, id, info.ID)
	assert.Equal(t, filePath, info.Path)
	assert.Equal(t, int64(12), info.Size)
	assert.False(t, info.CreatedAt.IsZero())
	assert.False(t, info.ExpiresAt.IsZero())
	assert.True(t, info.ExpiresAt.After(info.CreatedAt))
}

func TestTempFs_Close(t *testing.T) {
	ts := NewTempFs(1*time.Hour, 1*time.Hour)
	// Close should not panic or error
	err := ts.Close()
	assert.NoError(t, err)
}

func TestTempFs_Concurrency(t *testing.T) {
	ts := newTestTempFs(t)
	defer ts.Close()

	dir := t.TempDir()
	done := make(chan bool)

	// Concurrent register and get
	go func() {
		for i := 0; i < 10; i++ {
			filePath := createTempFile(t, dir, "content")
			id, err := ts.RegisterTempFile(filePath)
			if err == nil {
				ts.GetTempFile(id) //nolint:errcheck
			}
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			filePath := createTempFile(t, dir, "content")
			id, err := ts.RegisterTempFile(filePath)
			if err == nil {
				ts.DeleteTempFile(id) //nolint:errcheck
			}
		}
		done <- true
	}()

	<-done
	<-done
}
