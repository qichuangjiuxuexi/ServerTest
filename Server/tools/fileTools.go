package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetSqlPath() string {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("获取当前目录失败: %w", err))
	}

	// 获取父目录路径
	rootDir := filepath.Dir(currentDir)
	parentDir := filepath.Dir(rootDir)
	return parentDir
}

// ensureDirectoryExists 确保目录存在，不存在则创建
func EnsureDirectoryExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		fmt.Printf("目录 %s 不存在，正在创建...\n", dirPath)
		return os.MkdirAll(dirPath, 0755) // 使用MkdirAll可以创建多级目录
	} else if err != nil {
		return fmt.Errorf("检查目录 %s 时出错: %w", dirPath, err)
	}
	return nil
}

// 确保文件存在，不存在则创建
func EnsureFileExists(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("文件 %s 不存在，正在创建...\n", filePath)
		// 创建空文件
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("创建文件 %s 失败: %w", filePath, err)
		}
		file.Close()
		fmt.Printf("文件 %s 已创建\n", filePath)
	} else if err != nil {
		return fmt.Errorf("检查文件 %s 时出错: %w", filePath, err)
	}
	return nil
}
