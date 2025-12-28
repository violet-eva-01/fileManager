// Package main @author: Violet-Eva @date  : 2025/9/22 @notes :
package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// 生成文件编辑页面的HTML
func generateFileEditorHTML(fullPath, path string, user User) string {
	var htmlBuilder strings.Builder

	// 读取文件内容
	content, err := os.ReadFile(fullPath)
	fileContent := ""
	if err == nil {
		fileContent = strings.ReplaceAll(string(content), "\"", "&quot;")
		fileContent = strings.ReplaceAll(string(content), "<", "&lt;")
		fileContent = strings.ReplaceAll(string(content), ">", "&gt;")
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

	// 构建取消编辑链接的参数
	viewParams := url.Values{}
	viewParams.Add("path", path)
	cancelURL := "/file?" + viewParams.Encode()

	// 构建下载链接
	downloadParams := url.Values{}
	downloadParams.Add("path", path)
	downloadURL := "/file/download?" + downloadParams.Encode()

	htmlBuilder.WriteString("<!DOCTYPE html>")
	htmlBuilder.WriteString("<html><head>")
	htmlBuilder.WriteString("<meta charset=\"UTF-8\">")
	htmlBuilder.WriteString("<title>编辑文件 - " + fileName + "</title>")
	htmlBuilder.WriteString("<style>")
	htmlBuilder.WriteString("body { font-family: Arial, sans-serif; max-width: 1400px; margin: 0 auto; padding: 20px; }")
	htmlBuilder.WriteString("h1 { color: #333; border-bottom: 2px solid #4CAF50; padding-bottom: 10px; }")
	htmlBuilder.WriteString(".user-info { text-align: right; color: #666; margin-bottom: 10px; }")
	htmlBuilder.WriteString(".path-permissions { color: #666; font-style: italic; margin: -10px 0 15px 0; font-size: 0.9em; }")
	htmlBuilder.WriteString(".file-info { color: #666; margin: 10px 0; font-size: 14px; }")
	htmlBuilder.WriteString("textarea { width: 100%; height: 500px; padding: 10px; font-family: monospace; font-size: 14px; border: 1px solid #ddd; border-radius: 4px; }")
	htmlBuilder.WriteString("button, a { padding: 8px 15px; border: none; border-radius: 3px; cursor: pointer; text-decoration: none; font-size: 14px; margin-right: 10px; }")
	htmlBuilder.WriteString(".save-btn { background-color: #4CAF50; color: white; }")
	htmlBuilder.WriteString(".save-btn:hover { background-color: #45a049; }")
	htmlBuilder.WriteString(".cancel-btn { background-color: #ccc; color: black; }")
	htmlBuilder.WriteString(".cancel-btn:hover { background-color: #bbb; }")
	htmlBuilder.WriteString(".download-btn { background-color: #2196F3; color: white; }")
	htmlBuilder.WriteString(".download-btn:hover { background-color: #0b7dda; }")
	htmlBuilder.WriteString(".back-btn { background-color: #2196F3; color: white; }")
	htmlBuilder.WriteString(".back-btn:hover { background-color: #0b7dda; }")
	htmlBuilder.WriteString(".logout-btn { background-color: #f44336; color: white; }")
	htmlBuilder.WriteString(".logout-btn:hover { background-color: #d32f2f; }")
	htmlBuilder.WriteString(".actions { margin: 15px 0; }")
	htmlBuilder.WriteString("</style>")
	htmlBuilder.WriteString("</head><body>")

	// 用户信息和登录/登出按钮
	htmlBuilder.WriteString("<div class='user-info'>")
	if user.Username == guestUser.Username {
		htmlBuilder.WriteString("当前用户: 游客 | ")
		htmlBuilder.WriteString("<a href='/login' class='login-btn'>登录获取更高权限</a>")
	} else {
		htmlBuilder.WriteString("当前用户: " + user.Username + " (" + user.Role + ") | ")
		htmlBuilder.WriteString("<a href='/logout' class='logout-btn'>退出登录</a>")
	}
	htmlBuilder.WriteString("</div>")

	htmlBuilder.WriteString("<h1>编辑文件: " + fileName + "</h1>")

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

	htmlBuilder.WriteString("<a href=\"" + backURL + "\" class='back-btn'>返回目录</a>")
	htmlBuilder.WriteString("</div>")

	// 错误信息
	if err != nil {
		htmlBuilder.WriteString("<p style='color: red;'>读取文件时出错: " + err.Error() + "</p>")
	}

	// 编辑表单
	htmlBuilder.WriteString("<form method='post' action='/file/action'>")
	htmlBuilder.WriteString("<input type='hidden' name='action' value='edit'>")
	htmlBuilder.WriteString("<input type='hidden' name='path' value='" + path + "'>")
	htmlBuilder.WriteString("<textarea name='content'>" + fileContent + "</textarea>")
	htmlBuilder.WriteString("<div class='actions'>")
	htmlBuilder.WriteString("<button type='submit' class='save-btn'>保存</button>")
	htmlBuilder.WriteString("<a href=\"" + cancelURL + "\" class='cancel-btn'>取消更改</a>")
	htmlBuilder.WriteString("</div>")
	htmlBuilder.WriteString("</form>")
	htmlBuilder.WriteString("</body></html>")
	return htmlBuilder.String()
}
