// Package fm @author: Violet-Eva @date  : 2025/9/22 @notes :
package fm

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// 格式化文件大小显示
func formatFileSize(size int64) string {
	switch {
	case size >= 1<<50: // 1PB以上
		return fmt.Sprintf("%.2f PB", float64(size)/(1<<50))
	case size >= 1<<40: // 1TB以上
		return fmt.Sprintf("%.2f TB", float64(size)/(1<<40))
	case size >= 1<<30: // 1GB以上
		return fmt.Sprintf("%.2f GB", float64(size)/(1<<30))
	case size >= 1<<20: // 1MB以上
		return fmt.Sprintf("%.2f MB", float64(size)/(1<<20))
	case size >= 1<<10: // 1KB以上
		return fmt.Sprintf("%.2f KB", float64(size)/(1<<10))
	default: // 字节
		return fmt.Sprintf("%d B", size)
	}
}

// 判断是否为隐藏文件
func isHiddenFile(name string) bool {
	return strings.HasPrefix(name, ".")
}

// 文件管理器
func (fm *FileManager) generateFileManagerHTML(fullPath, path string, editMode bool, user User) string {
	var htmlBuilder strings.Builder

	// 检查路径是否存在
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		htmlBuilder.WriteString("<html><head><title>文件管理器</title></head>")
		htmlBuilder.WriteString("<body><h1>错误</h1>")
		htmlBuilder.WriteString("<fm>路径不存在: " + err.Error() + "</fm>")
		htmlBuilder.WriteString("<fm><a href=\"/file?path=\">返回根目录</a></fm>")
		if user.Username != fm.guestUser.Username {
			htmlBuilder.WriteString("<fm><a href=\"/logout\">退出登录</a></fm>")
		} else {
			htmlBuilder.WriteString("<fm><a href=\"/login\">登录获取更高权限</a></fm>")
		}
		htmlBuilder.WriteString("</body></html>")
		return htmlBuilder.String()
	}

	// 如果是文件，根据模式显示查看或编辑界面
	if !fileInfo.IsDir() {
		if editMode {
			return fm.generateFileEditorHTML(fullPath, path, user)
		} else {
			return fm.generateFileViewerHTML(fullPath, path, user)
		}
	}

	// 构建当前路径的URL参数
	currentParams := url.Values{}
	if path != "" {
		currentParams.Add("path", path)
	}
	currentURL := "/file?" + currentParams.Encode()

	// 目录列表页面
	htmlBuilder.WriteString("<!DOCTYPE html>")
	htmlBuilder.WriteString("<html><head>")
	htmlBuilder.WriteString("<meta charset=\"UTF-8\">")
	htmlBuilder.WriteString("<title>文件管理器 - " + path + "</title>")
	htmlBuilder.WriteString("<style>")
	htmlBuilder.WriteString("body { font-family: Arial, sans-serif; max-width: 1400px; margin: 0 auto; padding: 20px; }")
	htmlBuilder.WriteString("h1 { color: #333; border-bottom: 2px solid #4CAF50; padding-bottom: 10px; }")
	htmlBuilder.WriteString(".user-info { text-align: right; color: #666; margin-bottom: 10px; }")
	htmlBuilder.WriteString(".path-permissions { color: #666; font-style: italic; margin: -10px 0 15px 0; font-size: 0.9em; }")
	htmlBuilder.WriteString(".header-actions { display: flex; justify-content: space-between; margin: 10px 0 20px; align-items: center; }")
	htmlBuilder.WriteString(".action-buttons { display: flex; gap: 10px; }")
	htmlBuilder.WriteString(".file-list-container { margin-top: 20px; border: 1px solid #ddd; border-radius: 4px; overflow: hidden; }")
	htmlBuilder.WriteString(".file-list { list-style: none; padding: 0; margin: 0; }")
	htmlBuilder.WriteString(".file-header { padding: 12px 15px; background-color: #4CAF50; color: white; display: grid;")
	htmlBuilder.WriteString("grid-template-columns: 4fr 1fr 1fr 1fr 2fr; font-weight: bold; }")
	htmlBuilder.WriteString(".file-header span:last-child { text-align: center; }")
	htmlBuilder.WriteString(".file-item { padding: 12px 15px; display: grid;")
	htmlBuilder.WriteString("grid-template-columns: 4fr 1fr 1fr 1fr 2fr; align-items: center; border-bottom: 1px solid #ddd; }")
	htmlBuilder.WriteString(".file-item:nth-child(even) { background-color: #f9f9f9; }")
	htmlBuilder.WriteString(".file-item:hover { background-color: #f1f1f1; }")

	// 颜色样式设置
	htmlBuilder.WriteString(".file-item .file-name { font-family: Arial, sans-serif; font-size: 16px; font-weight: normal; color: #000000; }") // 普通文件为黑色
	htmlBuilder.WriteString(".file-item .dir-name { color: #0000FF; }")                                                                        // 普通目录为蓝色
	htmlBuilder.WriteString(".file-item .hidden-file { color: #808080; }")                                                                     // 隐藏文件为灰色
	htmlBuilder.WriteString(".file-item .hidden-dir { color: #00008B; }")                                                                      // 隐藏目录为深蓝色

	// 确保链接颜色继承父元素的颜色
	htmlBuilder.WriteString(".file-item .file-name a { color: inherit; text-decoration: none; }")
	htmlBuilder.WriteString(".file-item .file-name a:hover { text-decoration: underline; }")

	htmlBuilder.WriteString(".file-type, .file-size, .file-mtime { font-size: 14px; color: #666; }")
	htmlBuilder.WriteString(".actions { display: flex; gap: 8px; justify-content: flex-end; }")
	htmlBuilder.WriteString("button, a { padding: 5px 10px; border: none; border-radius: 3px; cursor: pointer; text-decoration: none; font-size: 14px; }")
	htmlBuilder.WriteString(".view-btn { background-color: #4CAF50; color: white; }")
	htmlBuilder.WriteString(".view-btn:hover { background-color: #45a049; }")
	htmlBuilder.WriteString(".edit-btn { background-color: #FFC107; color: black; }")
	htmlBuilder.WriteString(".edit-btn:hover { background-color: #e6ac00; }")
	htmlBuilder.WriteString(".delete-btn { background-color: #F44336; color: white; }")
	htmlBuilder.WriteString(".delete-btn:hover { background-color: #d32f2f; }")
	htmlBuilder.WriteString(".download-btn { background-color: #2196F3; color: white; }")
	htmlBuilder.WriteString(".download-btn:hover { background-color: #0b7dda; }")
	htmlBuilder.WriteString(".rename-btn { background-color: #FF9800; color: white; }")
	htmlBuilder.WriteString(".rename-btn:hover { background-color: #e68900; }")
	htmlBuilder.WriteString(".create-file-btn { background-color: #2196F3; color: white; }")
	htmlBuilder.WriteString(".create-file-btn:hover { background-color: #0b7dda; }")
	htmlBuilder.WriteString(".create-dir-btn { background-color: #9C27B0; color: white; }")
	htmlBuilder.WriteString(".create-dir-btn:hover { background-color: #7B1FA2; }")
	htmlBuilder.WriteString(".upload-btn { background-color: #FF9800; color: white; }")
	htmlBuilder.WriteString(".upload-btn:hover { background-color: #e68900; }")
	htmlBuilder.WriteString(".refresh-btn { background-color: #555555; color: white; }")
	htmlBuilder.WriteString(".refresh-btn:hover { background-color: #333333; }")
	htmlBuilder.WriteString(".parent-link { display: inline-block; }")
	htmlBuilder.WriteString(".logout-btn { background-color: #f44336; color: white; }")
	htmlBuilder.WriteString(".logout-btn:hover { background-color: #d32f2f; }")
	htmlBuilder.WriteString(".login-btn { background-color: #2196F3; color: white; }")
	htmlBuilder.WriteString(".login-btn:hover { background-color: #0b7dda; }")

	// 弹窗样式
	htmlBuilder.WriteString(".modal { display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%; background-color: rgba(0,0,0,0.5); }")
	htmlBuilder.WriteString(".modal-content { background-color: #fefefe; margin: 15% auto; padding: 20px; border: 1px solid #888; width: 50%; max-width: 500px; border-radius: 5px; }")
	htmlBuilder.WriteString(".close { color: #aaa; float: right; font-size: 28px; font-weight: bold; cursor: pointer; }")
	htmlBuilder.WriteString(".close:hover { color: black; }")
	htmlBuilder.WriteString(".modal-form-group { margin: 15px 0; }")
	htmlBuilder.WriteString("input[type='text'] { padding: 8px; margin: 5px 0; width: 100%; border: 1px solid #ddd; border-radius: 3px; box-sizing: border-box; }")
	htmlBuilder.WriteString("input[type='file'] { margin: 10px 0; }")
	htmlBuilder.WriteString(".file-upload-info { color: #666; font-size: 14px; margin-top: 5px; }")
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

	// 页面标题
	htmlBuilder.WriteString("<h1>文件管理器")
	if path != "" {
		htmlBuilder.WriteString(" - " + path)
	}
	htmlBuilder.WriteString("</h1>")

	// 头部操作区
	htmlBuilder.WriteString("<div class='header-actions'>")

	// 上级目录链接
	if path != "" {
		parentPath := filepath.Dir(path)
		if parentPath == "." {
			parentPath = ""
		}

		if user.IsPathAllowed(parentPath) {
			params := url.Values{}
			params.Add("path", parentPath)
			htmlBuilder.WriteString("<a href=\"/file?" + params.Encode() + "\" class='parent-link'>↑ 上级目录</a>")
		} else {
			// 上级目录无权限时不显示链接
			htmlBuilder.WriteString("<span></span>")
		}
	} else {
		htmlBuilder.WriteString("<span></span>")
	}

	// 右侧操作按钮组
	htmlBuilder.WriteString("<div class='action-buttons'>")
	htmlBuilder.WriteString("<button class='refresh-btn' onclick='window.location.href=\"" + currentURL + "\"'>刷新</button>")

	// 有权限的用户可以上传文件
	if user.HasPermission(PermissionDirUpload) {
		htmlBuilder.WriteString("<button class='upload-btn' onclick='openUploadModal()'>上传文件</button>")
	}

	// 有权限的用户可以创建目录
	if user.HasPermission(PermissionDirCreate) {
		htmlBuilder.WriteString("<button class='create-file-btn' onclick='openCreateModal(false)'>创建文件</button>")
		htmlBuilder.WriteString("<button class='create-dir-btn' onclick='openCreateModal(true)'>创建目录</button>")
	}

	htmlBuilder.WriteString("</div>")
	htmlBuilder.WriteString("</div>")

	// 创建新文件/目录的弹窗
	if user.HasPermission(PermissionDirCreate) {
		htmlBuilder.WriteString("<div id='createModal' class='modal'>")
		htmlBuilder.WriteString("<div class='modal-content'>")
		htmlBuilder.WriteString("<span class='close' onclick='document.getElementById(\"createModal\").style.display=\"none\"'>&times;</span>")
		htmlBuilder.WriteString("<h3 id='createModalTitle'>创建新项</h3>")
		htmlBuilder.WriteString("<form method='post' action='/file/action'>")
		htmlBuilder.WriteString("<input type='hidden' name='action' value='create'>")
		htmlBuilder.WriteString("<input type='hidden' name='path' value='" + path + "'>")
		htmlBuilder.WriteString("<input type='hidden' name='is_dir' id='isDirInput' value='false'>")
		htmlBuilder.WriteString("<div class='modal-form-group'>")
		htmlBuilder.WriteString("<label for='itemName'>名称:</label>")
		htmlBuilder.WriteString("<input type='text' id='itemName' name='name' required>")
		htmlBuilder.WriteString("</div>")

		htmlBuilder.WriteString("<div class='modal-form-group'>")
		htmlBuilder.WriteString("<button type='submit' class='view-btn'>创建</button>")
		htmlBuilder.WriteString("<button type='button' class='cancel-btn' onclick='document.getElementById(\"createModal\").style.display=\"none\"'>取消</button>")
		htmlBuilder.WriteString("</div>")
		htmlBuilder.WriteString("</form>")
		htmlBuilder.WriteString("</div>")
		htmlBuilder.WriteString("</div>")
	}

	// 文件上传弹窗
	if user.HasPermission(PermissionDirUpload) {
		htmlBuilder.WriteString("<div id='uploadModal' class='modal'>")
		htmlBuilder.WriteString("<div class='modal-content'>")
		htmlBuilder.WriteString("<span class='close' onclick='document.getElementById(\"uploadModal\").style.display=\"none\"'>&times;</span>")
		htmlBuilder.WriteString("<h3>上传文件</h3>")
		htmlBuilder.WriteString("<form method='post' action='/file/upload' enctype='multipart/form-data'>")
		htmlBuilder.WriteString("<input type='hidden' name='path' value='" + path + "'>")
		htmlBuilder.WriteString("<div class='modal-form-group'>")
		htmlBuilder.WriteString("<label for='fileUpload'>选择文件:</label><br>")
		htmlBuilder.WriteString("<input type='file' id='fileUpload' name='file' required>")
		htmlBuilder.WriteString("<fm class='file-upload-info'>最大上传限制: " + fmt.Sprintf("%d", fm.maxUploadSize>>30) + " GB</fm>")
		htmlBuilder.WriteString("</div>")

		htmlBuilder.WriteString("<div class='modal-form-group'>")
		htmlBuilder.WriteString("<button type='submit' class='view-btn'>上传</button>")
		htmlBuilder.WriteString("<button type='button' class='cancel-btn' onclick='document.getElementById(\"uploadModal\").style.display=\"none\"'>取消</button>")
		htmlBuilder.WriteString("</div>")
		htmlBuilder.WriteString("</form>")
		htmlBuilder.WriteString("</div>")
		htmlBuilder.WriteString("</div>")
	}

	// 重命名弹窗
	if user.HasPermission(PermissionDirRename) || user.HasPermission(PermissionFileRename) {
		htmlBuilder.WriteString("<div id='renameModal' class='modal'>")
		htmlBuilder.WriteString("<div class='modal-content'>")
		htmlBuilder.WriteString("<span class='close' onclick='document.getElementById(\"renameModal\").style.display=\"none\"'>&times;</span>")
		htmlBuilder.WriteString("<h3 id='renameModalTitle'>重命名</h3>")
		htmlBuilder.WriteString("<form method='post' action='/file/action'>")
		htmlBuilder.WriteString("<input type='hidden' name='action' value='rename'>")
		htmlBuilder.WriteString("<input type='hidden' name='path' id='renamePath' value=''>")
		htmlBuilder.WriteString("<div class='modal-form-group'>")
		htmlBuilder.WriteString("<label for='newName'>新名称:</label>")
		htmlBuilder.WriteString("<input type='text' id='newName' name='new_name' required>")
		htmlBuilder.WriteString("</div>")

		htmlBuilder.WriteString("<div class='modal-form-group'>")
		htmlBuilder.WriteString("<button type='submit' class='view-btn'>确认</button>")
		htmlBuilder.WriteString("<button type='button' class='cancel-btn' onclick='document.getElementById(\"renameModal\").style.display=\"none\"'>取消</button>")
		htmlBuilder.WriteString("</div>")
		htmlBuilder.WriteString("</form>")
		htmlBuilder.WriteString("</div>")
		htmlBuilder.WriteString("</div>")
	}

	// 列出目录内容
	files, err := os.ReadDir(fullPath)
	if err != nil {
		htmlBuilder.WriteString("<fm>无法读取目录: " + err.Error() + "</fm>")
		htmlBuilder.WriteString("</body></html>")
		return htmlBuilder.String()
	}

	// 过滤无权限的文件和目录
	var filteredFiles []os.DirEntry
	for _, file := range files {
		fileName := file.Name()
		filePath := filepath.Join(path, fileName)
		encodedPath := strings.ReplaceAll(filePath, "\\", "/")

		// 检查是否有权限访问该项目
		if user.IsPathAllowed(encodedPath) {
			filteredFiles = append(filteredFiles, file)
		}
	}

	// 排序过滤后的文件
	sort.Slice(filteredFiles, func(i, j int) bool {
		iIsDir, _ := filteredFiles[i].Info()
		jIsDir, _ := filteredFiles[j].Info()

		if iIsDir.IsDir() && !jIsDir.IsDir() {
			return true
		}
		if !iIsDir.IsDir() && jIsDir.IsDir() {
			return false
		}

		return strings.ToLower(filteredFiles[i].Name()) < strings.ToLower(filteredFiles[j].Name())
	})

	if len(filteredFiles) == 0 {
		htmlBuilder.WriteString("<fm>目录为空或没有可访问的项目</fm>")
	} else {
		htmlBuilder.WriteString("<div class='file-list-container'>")
		htmlBuilder.WriteString("<ul class='file-list'>")
		// 文件列表标题行
		htmlBuilder.WriteString("<li class='file-header'>")
		htmlBuilder.WriteString("<span>名称</span>")
		htmlBuilder.WriteString("<span>类型</span>")
		htmlBuilder.WriteString("<span>大小</span>")
		htmlBuilder.WriteString("<span>修改时间</span>")
		htmlBuilder.WriteString("<span>操作</span>")
		htmlBuilder.WriteString("</li>")

		for _, file := range filteredFiles {
			fileName := file.Name()
			filePath := filepath.Join(path, fileName)
			encodedPath := strings.ReplaceAll(filePath, "\\", "/")

			fileInfo, err = file.Info()
			var fileType, fileSize, fileMtime string

			if err == nil {
				if file.IsDir() {
					fileType = "目录"
				} else {
					fileType = "文件"
				}

				if file.IsDir() {
					fileSize = "-"
				} else {
					fileSize = formatFileSize(fileInfo.Size())
				}

				fileMtime = fileInfo.ModTime().Format("2006-01-02 15:04:05")
			} else {
				fileType = "未知"
				fileSize = "未知"
				fileMtime = "未知"
			}

			htmlBuilder.WriteString("<li class='file-item'>")

			htmlBuilder.WriteString("<span class='file-name")
			if isHiddenFile(fileName) {
				htmlBuilder.WriteString(" hidden-file")
			}
			if file.IsDir() {
				if isHiddenFile(fileName) {
					htmlBuilder.WriteString(" hidden-dir")
				} else {
					htmlBuilder.WriteString(" dir-name")
				}
			}
			htmlBuilder.WriteString("'>")

			if file.IsDir() {
				params := url.Values{}
				params.Add("path", encodedPath)
				htmlBuilder.WriteString("<a href=\"/file?" + params.Encode() + "\">" + fileName + "/</a>")
			} else {
				// 文件名链接指向查看模式
				viewParams := url.Values{}
				viewParams.Add("path", encodedPath)
				htmlBuilder.WriteString("<a href=\"/file?" + viewParams.Encode() + "\" class='file-link'>" + fileName + "</a>")
			}
			htmlBuilder.WriteString("</span>")

			// 文件类型
			htmlBuilder.WriteString("<span class='file-type'>" + fileType + "</span>")

			// 文件大小
			htmlBuilder.WriteString("<span class='file-size'>" + fileSize + "</span>")

			// 修改时间
			htmlBuilder.WriteString("<span class='file-mtime'>" + fileMtime + "</span>")

			// 操作按钮
			htmlBuilder.WriteString("<span class='actions'>")

			// 查看按钮
			if !file.IsDir() && user.HasPermission(PermissionFileView) {
				viewParams := url.Values{}
				viewParams.Add("path", encodedPath)
				htmlBuilder.WriteString("<a href=\"/file?" + viewParams.Encode() + "\" class='view-btn'>查看</a>")
			}

			// 下载按钮
			if !file.IsDir() && user.HasPermission(PermissionFileDownload) {
				downloadParams := url.Values{}
				downloadParams.Add("path", encodedPath)
				htmlBuilder.WriteString("<a href=\"/file/download?" + downloadParams.Encode() + "\" class='download-btn'>下载</a>")
			}

			// 编辑按钮
			if !file.IsDir() && user.HasPermission(PermissionFileEdit) {
				editParams := url.Values{}
				editParams.Add("path", encodedPath)
				editParams.Add("edit", "true")
				htmlBuilder.WriteString("<a href=\"/file?" + editParams.Encode() + "\" class='edit-btn'>编辑</a>")
			}

			// 重命名按钮
			if file.IsDir() {
				if user.HasPermission(PermissionDirRename) {
					htmlBuilder.WriteString("<button class='rename-btn' onclick='openRenameModal(\"" + encodedPath + "\", \"" + fileName + "\")'>重命名</button>")
				}
			} else {
				if user.HasPermission(PermissionFileRename) {
					htmlBuilder.WriteString("<button class='rename-btn' onclick='openRenameModal(\"" + encodedPath + "\", \"" + fileName + "\")'>重命名</button>")
				}
			}

			// 删除按钮
			if file.IsDir() {
				if user.HasPermission(PermissionDirDelete) {
					htmlBuilder.WriteString("<form method='post' action='/file/action' onsubmit='return confirm(\"确定要删除目录 " + fileName + " 吗?\")' style='margin:0;'>")
					htmlBuilder.WriteString("<input type='hidden' name='action' value='delete'>")
					htmlBuilder.WriteString("<input type='hidden' name='path' value='" + encodedPath + "'>")
					htmlBuilder.WriteString("<button type='submit' class='delete-btn'>删除</button>")
					htmlBuilder.WriteString("</form>")
				}
			} else {
				if user.HasPermission(PermissionFileDelete) {
					htmlBuilder.WriteString("<form method='post' action='/file/action' onsubmit='return confirm(\"确定要删除文件 " + fileName + " 吗?\")' style='margin:0;'>")
					htmlBuilder.WriteString("<input type='hidden' name='action' value='delete'>")
					htmlBuilder.WriteString("<input type='hidden' name='path' value='" + encodedPath + "'>")
					htmlBuilder.WriteString("<button type='submit' class='delete-btn'>删除</button>")
					htmlBuilder.WriteString("</form>")
				}
			}

			htmlBuilder.WriteString("</span>")
			htmlBuilder.WriteString("</li>")
		}
		htmlBuilder.WriteString("</ul>")
		htmlBuilder.WriteString("</div>")
	}

	// 弹窗控制脚本
	htmlBuilder.WriteString("<script>")

	// 打开创建弹窗并设置类型
	htmlBuilder.WriteString("function openCreateModal(isDir) {")
	htmlBuilder.WriteString("  document.getElementById('isDirInput').value = isDir;")
	htmlBuilder.WriteString("  document.getElementById('createModalTitle').textContent = isDir ? '创建目录' : '创建文件';")
	htmlBuilder.WriteString("  document.getElementById('createModal').style.display = 'block';")
	htmlBuilder.WriteString("}")

	// 打开上传弹窗
	htmlBuilder.WriteString("function openUploadModal() {")
	htmlBuilder.WriteString("  document.getElementById('uploadModal').style.display = 'block';")
	htmlBuilder.WriteString("}")

	// 打开重命名弹窗
	htmlBuilder.WriteString("function openRenameModal(path, name) {")
	htmlBuilder.WriteString("  document.getElementById('renamePath').value = path;")
	htmlBuilder.WriteString("  document.getElementById('newName').value = name;")
	htmlBuilder.WriteString("  document.getElementById('newName').select();")
	htmlBuilder.WriteString("  document.getElementById('renameModal').style.display = 'block';")
	htmlBuilder.WriteString("}")

	// 点击外部关闭弹窗
	htmlBuilder.WriteString("window.onclick = function(event) {")
	htmlBuilder.WriteString("  var createModal = document.getElementById('createModal');")
	htmlBuilder.WriteString("  var uploadModal = document.getElementById('uploadModal');")
	htmlBuilder.WriteString("  var renameModal = document.getElementById('renameModal');")
	htmlBuilder.WriteString("  if (event.target == createModal) {")
	htmlBuilder.WriteString("    createModal.style.display = 'none';")
	htmlBuilder.WriteString("  }")
	htmlBuilder.WriteString("  if (event.target == uploadModal) {")
	htmlBuilder.WriteString("    uploadModal.style.display = 'none';")
	htmlBuilder.WriteString("  }")
	htmlBuilder.WriteString("  if (event.target == renameModal) {")
	htmlBuilder.WriteString("    renameModal.style.display = 'none';")
	htmlBuilder.WriteString("  }")
	htmlBuilder.WriteString("}")
	htmlBuilder.WriteString("</script>")
	htmlBuilder.WriteString("</body></html>")
	return htmlBuilder.String()
}
