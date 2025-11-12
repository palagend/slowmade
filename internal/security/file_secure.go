package security

import (
	"os"
)

// WriteSecureFile 安全写入文件（仅用户可读）
func WriteSecureFile(filename string, data []byte) error {
	// 使用O_EXCL确保文件不存在，避免覆盖攻击
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		// 如果写入失败，尝试安全删除文件
		os.Remove(filename)
		return err
	}

	// 强制同步到磁盘
	return file.Sync()
}

// SecureDelete 安全删除文件（多次覆盖）
func SecureDelete(filename string) error {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return err
	}

	size := fileInfo.Size()

	// 多次覆盖文件内容
	for pass := 0; pass < 3; pass++ {
		file, err := os.OpenFile(filename, os.O_WRONLY, 0)
		if err != nil {
			return err
		}

		// 用不同模式覆盖
		var pattern byte
		switch pass {
		case 0:
			pattern = 0x00 // 全零
		case 1:
			pattern = 0xFF // 全一
		case 2:
			pattern = 0xAA // 交替
		}

		data := make([]byte, size)
		for i := range data {
			data[i] = pattern
		}

		if _, err := file.Write(data); err != nil {
			file.Close()
			return err
		}

		file.Sync()
		file.Close()
	}

	// 最终删除文件
	return os.Remove(filename)
}
