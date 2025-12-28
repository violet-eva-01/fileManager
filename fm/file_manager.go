// Package fm @author: Violet-Eva @date  : 2025/9/22 @notes :
package fm

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
)

type FileManager struct {
	rootDir       string              // 目录
	maxUploadSize int64               // 文件最大上传大小
	cookieName    string              // cookie 名称
	maxAge        int                 // cookie 存续时间
	privateKey    *rsa.PrivateKey     // 私钥
	publicKey     *rsa.PublicKey      // 公钥
	port          string              // 端口
	hashPassword  func(string) string // 密码加密函数
	users         map[string]User     // 用户清单
	guestUser     User                // 默认游客权限
	ginMode       string              // gin的mode
	log           *zap.SugaredLogger  // 日志
	handlerFunc   []gin.HandlerFunc
	AuditLogs     []AuditLog
	localIP       string
}

// DefaultHashPassword 默认帐户密码加密规则
func DefaultHashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// 默认游客账户
var guestUser = User{
	Username: "guest",
	Permissions: map[string]bool{
		PermissionFileView: true,
		PermissionDirView:  true,
	},
	BaseRolePathRestrictions: []string{"/"},
}

func NewFileManager(dir string, logger *zap.Logger) *FileManager {
	privateKey, publicKey, err := generateRSAKeyPair()
	if err != nil {
		fmt.Printf("生成RSA密钥对失败: %v\n", err)
		return nil
	}
	return &FileManager{
		rootDir:       dir,
		maxUploadSize: 10 << 30,
		cookieName:    "fm_session",
		privateKey:    privateKey,
		publicKey:     publicKey,
		maxAge:        36000,
		port:          "8080",
		hashPassword:  DefaultHashPassword,
		log:           logger.Sugar(),
		guestUser:     guestUser,
	}
}

func (fm *FileManager) SetMaxUploadSize(maxUploadSize int64) *FileManager {
	fm.maxUploadSize = maxUploadSize
	return fm
}

func (fm *FileManager) SetCookieName(name string) *FileManager {
	fm.cookieName = name
	return fm
}

func (fm *FileManager) SetKey(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) *FileManager {
	fm.privateKey = privateKey
	fm.publicKey = publicKey
	return fm
}

func (fm *FileManager) SetPort(port string) *FileManager {
	fm.port = port
	return fm
}

func (fm *FileManager) SetHashPassword(password func(string) string) *FileManager {
	fm.hashPassword = password
	return fm
}

func (fm *FileManager) SetUsers(users map[string]User) *FileManager {
	fm.users = users
	return fm
}

func (fm *FileManager) SetGuestUser(guestUser User) *FileManager {
	fm.guestUser = guestUser
	return fm
}

func (fm *FileManager) SetGinMode(mode string) *FileManager {
	switch mode {
	case "debug":
		fm.ginMode = gin.DebugMode
	case "release":
		fm.ginMode = gin.ReleaseMode
	case "test":
		fm.ginMode = gin.TestMode
	default:
		fm.ginMode = gin.DebugMode
	}
	fm.ginMode = mode
	return fm
}

func (fm *FileManager) SetLog(zapLogger *zap.Logger) *FileManager {
	fm.log = zapLogger.Sugar()
	return fm
}

func (fm *FileManager) SetHandlerFunc(handlerFunc ...gin.HandlerFunc) *FileManager {
	fm.handlerFunc = handlerFunc
	return fm
}

type AuditLog struct {
	Ts         time.Time `json:"ts"          spark:"column:ts;type:date"`
	Method     string    `json:"method"      spark:"column:method"`
	URI        string    `json:"uri"         spark:"column:uri"`
	ClientIP   string    `json:"client_ip"   spark:"column:client_ip"`
	ResponseIP string    `json:"response_ip" spark:"column:response_ip"`
	StatusCode int       `json:"status_code" spark:"column:status_code"`
	StartTime  time.Time `json:"start_time"  spark:"column:start_time;type:timestamp_us"`
	EndTime    time.Time `json:"end_time"    spark:"column:end_time;type:timestamp_us"`
	Latency    string    `json:"latency"     spark:"column:latency"`
	Error      string    `json:"error"       spark:"column:error"`
}

func getServerIP() (string, error) {
	dial, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("no non-loopback IPv4 address found")
	}
	defer dial.Close()
	addr := dial.LocalAddr().(*net.UDPAddr)
	return addr.IP.String(), nil
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// 复制请求信息，将数据写入审计日志
func responseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}

		c.Writer = blw

		c.Next()

		responseBody := blw.body.String()
		if blw.Status() >= 400 {
			c.Set("errorMessage", responseBody)
		}
	}
}

// GinZapLogger 记录审计日志和服务日志
func (fm *FileManager) ginZapLogger() gin.HandlerFunc {
	getUsername := func(c *gin.Context) string {
		user, exist := c.Get("user")
		if exist {
			return user.(User).Username
		}
		return ""
	}
	return func(c *gin.Context) {
		startTime := time.Now()
		username := getUsername(c)
		c.Next()
		if username == "" {
			username = getUsername(c)
		}
		endTime := time.Now()
		latency := endTime.Sub(startTime)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		uri := c.Request.RequestURI
		errorMessage, _ := c.Get("errorMessage")

		var al AuditLog
		al.Ts = startTime
		al.Method = method
		al.URI = uri
		al.ClientIP = clientIP
		al.ResponseIP = fm.localIP
		al.StatusCode = statusCode
		al.StartTime = startTime
		al.EndTime = endTime
		al.Latency = fmt.Sprintf("%s", latency)
		al.Error = fmt.Sprintf("%s", errorMessage)

		logger := fm.log.With(
			zap.String("method", method),
			zap.String("uri", uri),
			zap.String("request_user", username),
			zap.String("client_ip", clientIP),
			zap.Int("status_code", statusCode),
			zap.String("start_time", startTime.Format("2006-01-02 15:04:05.000000")),
			zap.Duration("latency", latency),
		)

		if statusCode >= 500 {
			logger.Errorf("request failed ,err is: %s", errorMessage)
		} else if statusCode >= 400 {
			logger.Warnf("request failed ,warn is: %s", errorMessage)
		} else {
			logger.Info("request success")
		}
	}
}

func (fm *FileManager) recovery() gin.HandlerFunc {
	return gin.Recovery()
}

func (fm *FileManager) Run() {
	defer fm.log.Sync()
	gin.SetMode(fm.ginMode)
	engine := gin.New()
	if ip, err := getServerIP(); err != nil {
		fm.log.Warn(err.Error())
	} else {
		fm.localIP = ip
	}

	engine.Use(fm.ginZapLogger())
	engine.Use(responseLogger())
	engine.Use(fm.recovery())
	engine.Use(fm.handlerFunc...)

	engine.GET("/login", fm.showLoginForm)
	engine.POST("/login", fm.handleLogin)
	engine.GET("/logout", fm.handleLogout)

	authorized := engine.Group("/", fm.jwtAuthMiddleware())
	{
		authorized.GET("/file", fm.requirePermission(PermissionDirView), fm.handleFileManager)
		authorized.GET("/file/download", fm.requirePermission(PermissionFileDownload), checkPathPermission(), fm.handleFileDownload)
		authorized.POST("/file/upload", fm.requirePermission(PermissionDirUpload), checkPathPermission(), fm.handleFileUpload)
		authorized.GET("/file/edit", fm.requirePermission(PermissionFileEdit), checkPathPermission(), fm.handleFileEditor)
		authorized.POST("/file/action", fm.requirePermission(PermissionDirView), checkPathPermission(), fm.handleFileAction)
	}

	err := engine.Run(":" + fm.port)
	if err != nil {
		fm.log.Error(err)
	}
}

func (fm *FileManager) handleFileManager(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		user = guestUser
	}

	path := c.Query("path")
	editMode := c.Query("edit") == "true"
	fullPath := filepath.Join(fm.rootDir, path)
	html := fm.generateFileManagerHTML(fullPath, path, editMode, user.(User))
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// 文件下载
func (fm *FileManager) handleFileDownload(c *gin.Context) {
	path := c.Query("path")
	fullPath := filepath.Join(fm.rootDir, path)

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		c.String(http.StatusNotFound, "文件不存在: %v", err)
		return
	}

	if fileInfo.IsDir() {
		c.String(http.StatusBadRequest, "不能下载目录")
		return
	}

	var file *os.File
	if file, err = os.Open(fullPath); err != nil {
		c.String(http.StatusInternalServerError, "无法打开文件: %v", err)
		return
	}
	defer file.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileInfo.Name()))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	_, err = io.Copy(c.Writer, file)
	if err != nil {
		c.String(http.StatusInternalServerError, "下载文件失败: %v", err)
		return
	}
}

// 处理文件上传
func (fm *FileManager) handleFileUpload(c *gin.Context) {
	path := c.PostForm("path")
	targetDir := filepath.Join(fm.rootDir, path)
	user := c.MustGet("user").(User)

	fileInfo, err := os.Stat(targetDir)
	if err != nil || !fileInfo.IsDir() {
		c.String(http.StatusBadRequest, "目标目录不存在或不是目录: %v", err)
		return
	}

	if err = c.Request.ParseMultipartForm(fm.maxUploadSize); err != nil {
		c.String(http.StatusBadRequest, "上传文件过大: %v", err)
		return
	}

	file, handler, err := c.Request.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "获取上传文件失败: %v", err)
		return
	}
	defer file.Close()

	newFilePath := filepath.Join(path, handler.Filename)
	if !user.IsPathAllowed(newFilePath) {
		c.String(http.StatusForbidden, "没有权限上传文件到该位置")
		return
	}

	dstPath := filepath.Join(targetDir, handler.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		c.String(http.StatusInternalServerError, "创建文件失败: %v", err)
		return
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
		c.String(http.StatusInternalServerError, "保存文件失败: %v", err)
		return
	}

	params := url.Values{}
	params.Add("path", path)
	c.Redirect(http.StatusSeeOther, "/file?"+params.Encode())
}

// 文件编辑
func (fm *FileManager) handleFileEditor(c *gin.Context) {
	path := c.Query("path")
	params := url.Values{}
	params.Add("path", path)
	params.Add("edit", "true")
	c.Redirect(http.StatusSeeOther, "/file?"+params.Encode())
}

// 处理其他文件操作（编辑、新增、删除、重命名）
func (fm *FileManager) handleFileAction(c *gin.Context) {
	action := c.PostForm("action")
	path := c.PostForm("path")
	fullPath := filepath.Join(fm.rootDir, path)

	user, exists := c.Get("user")
	if !exists {
		user = guestUser
	}
	currentUser := user.(User)

	fileInfo, err := os.Stat(fullPath)
	isDir := false
	pathExists := true

	if err != nil {
		if os.IsNotExist(err) {
			pathExists = false
		} else {
			c.String(http.StatusInternalServerError, "获取文件信息失败: %v", err)
			return
		}
	} else {
		isDir = fileInfo.IsDir()
	}

	if action == "create" {
		name := c.PostForm("name")
		if isDirCreate := c.PostForm("is_dir"); isDirCreate == "true" {
			newPath := filepath.Join(path, name)

			if !currentUser.IsPathAllowed(newPath) {
				c.String(http.StatusForbidden, "没有权限在该位置创建内容")
				return
			}
		}
	}

	switch action {
	case "edit":
		if !pathExists {
			c.String(http.StatusNotFound, "文件不存在")
			return
		}
		if isDir {
			c.String(http.StatusBadRequest, "不能编辑目录")
			return
		}
		if !currentUser.HasPermission(PermissionFileEdit) {
			c.String(http.StatusForbidden, "没有文件编辑权限")
			return
		}
	case "create":
		name := c.PostForm("name")
		if name == "" {
			c.String(http.StatusBadRequest, "名称不能为空")
			return
		}

		isDirCreate := c.PostForm("is_dir") == "true"
		if isDirCreate {
			if !currentUser.HasPermission(PermissionDirCreate) {
				c.String(http.StatusForbidden, "没有目录创建权限")
				return
			}
		} else {
			// 创建文件需要目录的上传权限
			if !currentUser.HasPermission(PermissionDirUpload) {
				c.String(http.StatusForbidden, "没有文件创建权限")
				return
			}
		}
	case "delete":
		if !pathExists {
			c.String(http.StatusNotFound, "路径不存在")
			return
		}

		if isDir {
			if !currentUser.HasPermission(PermissionDirDelete) {
				c.String(http.StatusForbidden, "没有目录删除权限")
				return
			}
		} else {
			if !currentUser.HasPermission(PermissionFileDelete) {
				c.String(http.StatusForbidden, "没有文件删除权限")
				return
			}
		}
	case "rename":
		if !pathExists {
			c.String(http.StatusNotFound, "路径不存在")
			return
		}

		newName := c.PostForm("new_name")
		if newName == "" {
			c.String(http.StatusBadRequest, "新名称不能为空")
			return
		}

		// 检查重命名后的路径是否允许访问
		parentDir := filepath.Dir(path)
		newPath := filepath.Join(parentDir, newName)
		if !currentUser.IsPathAllowed(newPath) {
			c.String(http.StatusForbidden, "没有权限使用该名称或路径")
			return
		}

		if isDir {
			if !currentUser.HasPermission(PermissionDirRename) {
				c.String(http.StatusForbidden, "没有目录重命名权限")
				return
			}
		} else {
			if !currentUser.HasPermission(PermissionFileRename) {
				c.String(http.StatusForbidden, "没有文件重命名权限")
				return
			}
		}
	default:
		c.String(http.StatusBadRequest, "未知操作")
		return
	}

	switch action {
	case "edit":
		content := c.PostForm("content")

		// 在保存新内容前创建备份
		if err = createFileBackup(fullPath); err != nil {
			c.String(http.StatusInternalServerError, "创建文件备份失败: %v", err)
			return
		}

		// 保存新内容
		if err = os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			c.String(http.StatusInternalServerError, "编辑文件失败: %v", err)
			return
		}
	case "create":
		name := c.PostForm("name")
		isDirCreate := c.PostForm("is_dir") == "true"
		newPath := filepath.Join(fullPath, name)

		if isDirCreate {
			if err = os.MkdirAll(newPath, 0755); err != nil {
				c.String(http.StatusInternalServerError, "创建目录失败: %v", err)
				return
			}
		} else {
			if _, err = os.Create(newPath); err != nil {
				c.String(http.StatusInternalServerError, "创建文件失败: %v", err)
				return
			}
		}
	case "delete":
		if err = os.RemoveAll(fullPath); err != nil {
			c.String(http.StatusInternalServerError, "删除失败: %v", err)
			return
		}
		// 删除后返回父目录
		parentPath := filepath.Dir(path)
		if parentPath == "." {
			parentPath = ""
		}
		path = parentPath
	case "rename":
		newName := c.PostForm("new_name")
		// 获取父目录和旧名称
		parentDir := filepath.Dir(fullPath)
		newPath := filepath.Join(parentDir, newName)

		// 检查新名称是否已存在
		if _, err = os.Stat(newPath); err == nil {
			c.String(http.StatusBadRequest, "名称已存在: %s", newName)
			return
		}

		// 执行重命名
		if err = os.Rename(fullPath, newPath); err != nil {
			c.String(http.StatusInternalServerError, "重命名失败: %v", err)
			return
		}

		// 重命名后返回父目录
		path = filepath.Dir(path)
		if path == "." {
			path = ""
		}
	}

	// 操作完成后重定向
	params := url.Values{}
	params.Add("path", path)
	c.Redirect(http.StatusSeeOther, "/file?"+params.Encode())
}

// 创建文件备份
func createFileBackup(filePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在")
	}

	// 获取当前时间戳
	timestamp := time.Now().Format("20060102150405.000000")

	// 构建备份文件名：原文件名_时间戳
	dir := filepath.Dir(filePath)
	filename := filepath.Base(filePath)
	backupFilename := fmt.Sprintf("%s_%s", filename, timestamp)
	backupPath := filepath.Join(dir, backupFilename)

	// 打开源文件
	sourceFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// 创建备份文件
	destFile, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// 复制文件内容
	if _, err = io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// 确保所有数据都写入磁盘
	if err = destFile.Sync(); err != nil {
		return err
	}

	return nil
}
