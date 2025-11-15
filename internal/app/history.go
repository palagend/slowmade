package app

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/palagend/slowmade/pkg/logging"
	"go.uber.org/zap"
)

// HistoryManager 管理命令历史记录
type HistoryManager struct {
	commands []string
	filePath string
	maxSize  int
	logger   *zap.Logger
}

// NewHistoryManager 创建历史记录管理器
func NewHistoryManager() *HistoryManager {
	homeDir, err := os.UserHomeDir()
	filePath := ""
	if err == nil {
		filePath = filepath.Join(homeDir, ".slowmade_history")
	}

	return &HistoryManager{
		commands: make([]string, 0),
		filePath: filePath,
		maxSize:  1000,
		logger:   logging.Get(),
	}
}

// Add 添加命令到历史记录
func (h *HistoryManager) Add(command string) {
	command = strings.TrimSpace(command)
	if command == "" {
		return
	}

	// 去重：如果与上一条命令相同，则不添加
	if len(h.commands) > 0 && h.commands[len(h.commands)-1] == command {
		return
	}

	h.commands = append(h.commands, command)

	// 限制历史记录大小
	if len(h.commands) > h.maxSize {
		h.commands = h.commands[len(h.commands)-h.maxSize:]
	}
}

// GetLast 获取最后 n 条历史记录
func (h *HistoryManager) GetLast(n int) []string {
	if n <= 0 || len(h.commands) == 0 {
		return []string{}
	}

	if n > len(h.commands) {
		n = len(h.commands)
	}

	return h.commands[len(h.commands)-n:]
}

// Save 保存历史记录到文件
func (h *HistoryManager) Save() error {
	if h.filePath == "" {
		return nil
	}

	file, err := os.Create(h.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, cmd := range h.commands {
		if _, err := file.WriteString(cmd + "\n"); err != nil {
			return err
		}
	}

	return nil
}

// Load 从文件加载历史记录
func (h *HistoryManager) Load() error {
	if h.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(h.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在是正常的
		}
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			h.commands = append(h.commands, line)
		}
	}

	return nil
}
