package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() {

	// 1. 遍历所有子目录,获取最新的文件复制为index.html
	err := updateIndexHtml("docs")
	if err != nil {
		fmt.Println(err)
		return
	}

	// 2. 生成more.html页面,列出所有html文件
	err = generateMoreHtml("docs")
	if err != nil {
		fmt.Println(err)
		return
	}
}

// 更新index.html
func updateIndexHtml(root string) error {

	var latestFile string
	var latestTime int64

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fileName := info.Name()
		if !strings.HasSuffix(fileName, ".html") {
			return nil
		}

		timestamp := extractTimestamp(fileName)
		if timestamp > latestTime {
			latestTime = timestamp
			latestFile = path
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 复制最新文件为index.html
	copyFile(latestFile, filepath.Join(root, "index.html"))
	fmt.Println("update index.html successfully!")

	return nil
}

// 从文件名中提取时间戳
func extractTimestamp(name string) int64 {
	parts := strings.Split(name, "-")
	if len(parts) < 2 {
		return 0
	}

	timestamp, err := strconv.ParseInt(parts[1][:8], 10, 64)
	if err != nil {
		return 0
	}

	return timestamp
}

// 复制文件
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func generateMoreHtml(docsDir string) error {
	// 读取docs目录下的所有文件/目录,除了index.html和more.html
	var files []string
	err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.Name() == "index.html" || info.Name() == "more.html" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return err
	}

	// 对文件路径按月份分类
	fileMap := make(map[string][]string)
	for _, file := range files {
		month := filepath.Base(filepath.Dir(file))
		fileMap[month] = append(fileMap[month], file)
	}

	// 对月份进行排序
	months := make([]string, 0, len(fileMap))
	for month := range fileMap {
		months = append(months, month)
	}
	sort.Slice(months, func(i, j int) bool {
		return months[i] > months[j]
	})

	if len(months) > 12 {
		months = months[:12] // 取最近12个月
	}

	// 生成more.html内容
	var content string
	content += "<html>\n"
	content += "<body>\n"
	content += "<h1>GopherDaily归档</h1>\n"
	content += "<p>此页面列出GopherDaily最近12个月的所有html文件，以月份分类，从新到旧排序，点击超链查看对应日期的GopherDaily内容。</p>\n"
	content += "<p><a href=\"index.html\">回到首页</a></p>\n"
	for _, month := range months {
		content += fmt.Sprintf("<h2>%s</h2>\n", month)
		files := fileMap[month]
		// 对月份下文件进行排序
		sort.Slice(files, func(i, j int) bool {
			return files[i] > files[j]
		})

		for _, file := range files {
			fileName := filepath.Base(file)
			//fileLink := fmt.Sprintf("<a href=\"%s\">%s</a>", file, fileName)
			file = strings.TrimPrefix(file, "docs/")
			fileLink := fmt.Sprintf("<a href=\"%s\">%s</a>", file, fileName)
			content += fmt.Sprintf("<p>%s</p>\n", fileLink)
		}
	}
	content += "</body>\n"
	content += "</html>"

	// 写入more.html文件
	err = ioutil.WriteFile(filepath.Join(docsDir, "/more/index.html"), []byte(content), 0644)
	if err != nil {
		return err
	}

	fmt.Println("more/index.html generated successfully!")
	return nil
}
