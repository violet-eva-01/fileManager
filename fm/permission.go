// Package main @author: Violet-Eva @date  : 2025/9/22 @notes :
package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// 权限
const (
	PermissionFileView     = "file:view"     // 查看文件
	PermissionFileDownload = "file:download" // 下载文件
	PermissionFileEdit     = "file:edit"     // 编辑文件
	PermissionFileDelete   = "file:delete"   // 删除文件
	PermissionFileRename   = "file:rename"   // 重命名文件
	PermissionDirView      = "dir:view"      // 查看目录
	PermissionDirCreate    = "dir:create"    // 创建目录
	PermissionDirUpload    = "dir:upload"    // 上传文件到目录
	PermissionDirDelete    = "dir:delete"    // 删除目录
	PermissionDirRename    = "dir:rename"    // 重命名目录
)

type User struct {
	Username                 string
	EncryptPassword          string
	Role                     string
	Permissions              map[string]bool
	BaseRolePathRestrictions []string
	BaseRolePathBlocking     []string
	IP                       string
}

func (u *User) String() string {
	return u.Username
}

func (u *User) IsAdmin() bool {
	if u.Role == "admin" {
		return true
	} else {
		return false
	}
}

func (u *User) HasPermission(permission string) bool {
	if u.IsAdmin() {
		return true
	}
	return u.Permissions[permission]
}

// isPathBlocked 检查路径是否被屏蔽
func (u *User) isPathBlocked(path string) bool {
	if u.IsAdmin() || len(u.BaseRolePathBlocking) == 0 {
		return false
	}

	for _, pattern := range u.BaseRolePathBlocking {
		if matchPathPattern(pattern, path) {
			return true
		}
	}

	return false
}

// IsPathAllowed 检查路径是否可以访问
func (u *User) IsPathAllowed(path string) bool {
	if u.IsAdmin() {
		return true
	}

	if u.isPathBlocked(path) {
		return false
	}

	for _, pattern := range u.BaseRolePathRestrictions {
		if matchPathPattern(pattern, path) {
			return true
		}
	}
	return false
}

// 检查路径是否匹配
func matchPathPattern(pattern, path string) bool {
	if pattern == "/" && path == "/" {
		return true
	}

	normalizedPattern := "/" + strings.Trim(pattern, "/")
	normalizedPath := "/" + strings.Trim(path, "/")
	if normalizedPattern == normalizedPath {
		return true
	}

	regexpStr := fmt.Sprintf("^%s", normalizedPattern)
	compile := regexp.MustCompile(regexpStr)
	fmt.Printf("matchPathPattern: pattern=%s path=%s\n", normalizedPattern, normalizedPath)
	return compile.MatchString(normalizedPath) || strings.HasPrefix(normalizedPattern, normalizedPath)
}

// 权限中间件
func (fm *FileManager) requirePermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tmpUser, isExist := c.Get("user")
		var user User
		if !isExist {
			user = guestUser
		} else {
			var exists bool
			user, exists = tmpUser.(User)
			if !exists {
				user = guestUser
			}
		}
		c.Set("user", user)
		if !user.HasPermission(requiredPermission) {
			c.String(http.StatusForbidden, "权限不足，无法访问此功能")
			c.Abort()
			return
		}

		path := c.Query("path")
		if !user.IsPathAllowed(path) {
			c.String(http.StatusForbidden, "没有权限访问此路径")
			c.Abort()
			return
		}
		c.Next()
	}
}

// 路径检查中间件
func checkPathPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			// 如果没有用户信息，赋予user角色
			user = guestUser
			c.Set("user", user)
		}
		currentUser := user.(User)

		path := c.PostForm("path")
		if path == "" {
			path = c.Query("path")
		}

		if !currentUser.IsPathAllowed(path) {
			c.String(http.StatusForbidden, "没有权限访问此路径")
			c.Abort()
			return
		}

		c.Next()
	}
}
