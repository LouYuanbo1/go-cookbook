package tempfs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
	"uuid"
)

type TempFileInfo struct {
	ID        string
	Path      string
	Size      int64
	CreatedAt time.Time
	ExpiresAt time.Time
}

type TempFs interface {
	RegisterTempFile(path string) (id string, err error)
	GetTempFile(id string) (*TempFileInfo, error)
	DeleteTempFile(id string) error
	PromoteTempFile(id, destDir, newFilename string) error
	Close() error // 新增关闭方法，停止后台协程
}

type tempFs struct {
	tempFiles map[string]*TempFileInfo
	mu        sync.RWMutex
	expiresIn time.Duration
	stopCh    chan struct{}  // 通知清理协程退出
	wg        sync.WaitGroup // 等待清理协程结束
}

func NewTempFs(expiresIn, cleanupInterval time.Duration) TempFs {
	t := &tempFs{
		tempFiles: make(map[string]*TempFileInfo),
		expiresIn: expiresIn,
		stopCh:    make(chan struct{}),
	}
	t.wg.Go(func() {
		t.cleanupLoop(cleanupInterval)
	})
	return t
}

// Close 停止后台清理协程，并等待其结束
func (t *tempFs) Close() error {
	close(t.stopCh)
	t.wg.Wait()
	return nil
}

// RegisterTempFile 创建临时文件并注册，返回唯一ID
func (t *tempFs) RegisterTempFile(path string) (id string, err error) {
	// 3. 获取实际路径信息
	stat, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	// 3. 生成ID并存储
	tempID := uuid.NewV7().String()
	now := time.Now()
	info := &TempFileInfo{
		ID:        tempID,
		Path:      path,
		Size:      stat.Size(),
		CreatedAt: now,
		ExpiresAt: now.Add(t.expiresIn),
	}

	t.mu.Lock()
	t.tempFiles[tempID] = info
	t.mu.Unlock()

	return tempID, nil
}

// GetTempFile 获取临时文件信息，若不存在或已过期则返回错误
func (t *tempFs) GetTempFile(id string) (*TempFileInfo, error) {
	t.mu.RLock()
	info, ok := t.tempFiles[id]
	t.mu.RUnlock()

	if !ok {
		return nil, errors.New("临时ID不存在或已过期")
	}
	if time.Now().After(info.ExpiresAt) {
		// 过期但不清除，由调用者或清理协程处理；也可直接返回错误
		return nil, errors.New("临时文件已过期")
	}
	return info, nil
}

// DeleteTempFile 删除临时文件（物理删除并从map中移除）
func (t *tempFs) DeleteTempFile(id string) error {
	t.mu.Lock()
	info, ok := t.tempFiles[id]
	if ok {
		delete(t.tempFiles, id)
	}
	t.mu.Unlock()

	if !ok {
		return nil // 已不存在，视为成功
	}

	// 物理删除文件（忽略错误，因为可能已被外部删除）
	_ = os.Remove(info.Path)
	return nil
}

// PromoteTempFile 将临时文件移动到正式目录，并重命名
func (t *tempFs) PromoteTempFile(id, destDir, newFilename string) error {
	// 加写锁，避免在移动过程中被清理或删除
	t.mu.Lock()
	info, ok := t.tempFiles[id]
	if !ok {
		t.mu.Unlock()
		return errors.New("临时ID不存在或已过期")
	}
	if time.Now().After(info.ExpiresAt) {
		t.mu.Unlock()
		return errors.New("临时文件已过期")
	}
	// 提前取出路径，然后从map中移除（防止后续被操作）
	srcPath := info.Path
	delete(t.tempFiles, id)
	t.mu.Unlock()

	// 创建目标目录
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}
	destPath := filepath.Join(destDir, newFilename)

	// 执行移动（此时map中已无记录，即使移动失败文件也还在，但可重新注册）
	if err := os.Rename(srcPath, destPath); err != nil {
		return fmt.Errorf("移动文件失败: %w", err)
	}
	return nil
}

// cleanupLoop 定期清理过期文件（后台协程）
func (t *tempFs) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.cleanupExpired()
		case <-t.stopCh:
			return // 收到停止信号，退出
		}
	}
}

// cleanupExpired 清理所有过期文件（内部已加锁）
func (t *tempFs) cleanupExpired() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	for id, info := range t.tempFiles {
		if now.After(info.ExpiresAt) {
			// 物理删除（忽略错误）
			_ = os.Remove(info.Path)
			delete(t.tempFiles, id)
		}
	}
}
