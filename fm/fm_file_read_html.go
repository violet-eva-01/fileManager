// Package fm @author: Violet-Eva @date  : 2025/9/22 @notes :
package fm

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// 文件只读查看页面
func (fm *FileManager)generateFileViewerHTML(fullPath, path string, user User) string {
	var htmlBuilder strings.Builder

	// 读取文件内容
	content, err := os.ReadFile(fullPath)
	fileContent := ""
	if err == nil {
		fileContent = strings.ReplaceAll(string(content), "\"", "&quot;")
		fileContent = strings.ReplaceAll(fileContent, "<", "&lt;")
		fileContent = strings.ReplaceAll(fileContent, ">", "&gt;")
		// 将换行符转换为<br>标签以便在HTML中显示
		fileContent = strings.ReplaceAll(fileContent, "\n", "<br>")
	}

	// 获取文件信息
	fileInfo, _ := os.Stat(fullPath)
	var fileInfoStr string
	if fileInfo != nil {
		size := formatFileSize(fileInfo.Size())
		mtime := fileInfo.ModTime().Format("2006-01-02 15:04:05")
		fileInfoStr = fmt.Sprintf("文件大小: %s | 最后修改时间: %s", size, mtime)
	}

	// 获取文件名和父目录
	fileName := filepath.Base(path)
	parentPath := filepath.Dir(path)
	if parentPath == "." {
		parentPath = ""
	}

	// 构建返回链接的参数
	params := url.Values{}
	params.Add("path", parentPath)
	backURL := "/file?" + params.Encode()

	// 构建编辑模式的链接
	editParams := url.Values{}
	editParams.Add("path", path)
	editParams.Add("edit", "true")
	editURL := "/file?" + editParams.Encode()

	// 构建下载链接
	downloadParams := url.Values{}
	downloadParams.Add("path", path)
	downloadURL := "/file/download?" + downloadParams.Encode()

	htmlBuilder.WriteString("<!DOCTYPE html>")
	htmlBuilder.WriteString("<html><head>")
	htmlBuilder.WriteString("<meta charset=\"UTF-8\">")
	htmlBuilder.WriteString("<title>查看文件 - " + fileName + "</title>")
	htmlBuilder.WriteString("<style>")
	htmlBuilder.WriteString("body { font-family: Arial, sans-serif; max-width: 1400px; margin: 0 auto; padding: 20px; }")
	htmlBuilder.WriteString("h1 { color: #333; border-bottom: 2px solid #4CAF50; padding-bottom: 10px; }")
	htmlBuilder.WriteString(".user-info { text-align: right; color: #666; margin-bottom: 10px; }")
	htmlBuilder.WriteString(".path-permissions { color: #666; font-style: italic; margin: -10px 0 15px 0; font-size: 0.9em; }")
	htmlBuilder.WriteString(".file-info { color: #666; margin: 10px 0; font-size: 14px; }")
	htmlBuilder.WriteString(".file-content { margin: 20px 0; padding: 15px; background-color: #f8f8f8; border: 1px solid #ddd; border-radius: 4px; font-family: monospace; white-space: pre-wrap; word-wrap: break-word; }")
	htmlBuilder.WriteString("button, a { padding: 8px 15px; border: none; border-radius: 3px; cursor: pointer; text-decoration: none; font-size: 14px; margin-right: 10px; }")
	htmlBuilder.WriteString(".edit-btn { background-color: #FFC107; color: black; }")
	htmlBuilder.WriteString(".edit-btn:hover { background-color: #e6ac00; }")
	htmlBuilder.WriteString(".download-btn { background-color: #2196F3; color: white; }")
	htmlBuilder.WriteString(".download-btn:hover { background-color: #0b7dda; }")
	htmlBuilder.WriteString(".back-btn { background-color: #2196F3; color: white; }")
	htmlBuilder.WriteString(".back-btn:hover { background-color: #0b7dda; }")
	htmlBuilder.WriteString(".logout-btn { background-color: #f44336; color: white; }")
	htmlBuilder.WriteString(".logout-btn:hover { background-color: #d32f2f; }")
	htmlBuilder.WriteString(".login-btn { background-color: #2196F3; color: white; }")
	htmlBuilder.WriteString(".login-btn:hover { background-color: #0b7dda; }")
	htmlBuilder.WriteString(".actions { margin: 15px 0; }")
	htmlBuilder.WriteString("</style>")
	htmlBuilder.WriteString("</head><body>")

	// 用户信息和登录/登出按钮
	htmlBuilder.WriteString("<div class='user-info'>")
	if user.Username == fm.guestUser.Username {
		htmlBuilder.WriteString("当前用户: 游客 | ")
		htmlBuilder.WriteString("<a href='/login' class='login-btn'>登录获取更高权限</a>")
	} else {
		htmlBuilder.WriteString("当前用户: " + user.Username + " | ")
		htmlBuilder.WriteString("<a href='/logout' class='logout-btn'>退出登录</a>")
	}
	htmlBuilder.WriteString("</div>")

	htmlBuilder.WriteString("<h1>查看文件: " + fileName + "</h1>")

	// 显示文件信息
	if fileInfoStr != "" {
		htmlBuilder.WriteString("<div class='file-info'>" + fileInfoStr + "</div>")
	}

	// 操作按钮
	htmlBuilder.WriteString("<div class='actions'>")

	// 下载按钮
	if user.HasPermission(PermissionFileDownload) {
		htmlBuilder.WriteString("<a href=\"" + downloadURL + "\" class='download-btn'>下载文件</a>")
	}

	// 编辑按钮
	if user.HasPermission(PermissionFileEdit) {
		htmlBuilder.WriteString("<a href=\"" + editURL + "\" class='edit-btn'>编辑文件</a>")
	}

	htmlBuilder.WriteString("<a href=\"" + backURL + "\" class='back-btn'>返回目录</a>")
	htmlBuilder.WriteString("</div>")

	// 错误信息
	if err != nil {
		htmlBuilder.WriteString("<p style='color: red;'>读取文件时出错: " + err.Error() + "</p>")
	} else {
		// 显示文件内容
		htmlBuilder.WriteString("<div class='file-content'>")
		htmlBuilder.WriteString(fileContent)
		htmlBuilder.WriteString("</div>")
	}
	htmlBuilder.WriteString("</body></html>")
	return htmlBuilder.String()
}
