// Package fm @author: Violet-Eva @date  : 2025/12/26 @notes :
package fm

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// 登录菜单
func (fm *FileManager) showLoginForm(c *gin.Context) {
	if _, i := c.Get("user"); i {
		c.Redirect(http.StatusSeeOther, "/file")
		return
	}

	html := `
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>登录 - 文件管理器</title>
		<style>
			body { font-family: Arial, sans-serif; max-width: 400px; margin: 50px auto; padding: 20px; }
			.login-form { border: 1px solid #ddd; padding: 20px; border-radius: 5px; }
			.form-group { margin-bottom: 15px; }
			label { display: block; margin-bottom: 5px; }
			input[type="text"], input[type="password"] { width: 100%; padding: 8px; box-sizing: border-box; }
			button { padding: 8px 15px; background-color: #4CAF50; color: white; border: none; border-radius: 3px; cursor: pointer; }
			button:hover { background-color: #45a049; }
			.error { color: red; margin-bottom: 15px; }
			.guest-access { margin-top: 15px; padding-top: 15px; border-top: 1px solid #ddd; }
		</style>
	</head>
	<body>
		<div class="login-form">
			<h2>文件管理器登录</h2>
			<p>登录以获取更高权限，或直接访问以游客角色浏览</p>
			` + c.Query("error") + `
			<form method="post" action="/login">
				<div class="form-group">
					<label for="username">用户名:</label>
					<input type="text" id="username" name="username" required>
				</div>
				<div class="form-group">
					<label for="password">密码:</label>
					<input type="password" id="password" name="password" required>
				</div>
				<div class="form-group">
					<button type="submit">登录</button>
				</div>
			</form>
			<div class="guest-access">
				<a href="/file">以游客用户身份访问</a>
			</div>
		</div>
	</body>
	</html>
	`

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// 创建默认JWT密钥
func generateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	publicKey := &privateKey.PublicKey
	return privateKey, publicKey, nil
}

type JWTClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// 生成Token
func (fm *FileManager) generateJWTToken(username string) (string, error) {
	claims := JWTClaims{
		username,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token 24小时后过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    fm.cookieName + "-jwt",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(fm.privateKey)
}

// 验证Token
func (fm *FileManager) validateJWTToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return fm.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// 获取请求的Token
func (fm *FileManager) extractTokenFromRequest(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
		return authHeader
	}

	token, err := c.Cookie(fm.cookieName)
	if err == nil {
		return token
	}

	token = c.Query("token")
	return token
}

// 认证插件
func (fm *FileManager) jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := fm.extractTokenFromRequest(c)
		if tokenString != "" {
			claims, err := fm.validateJWTToken(tokenString)
			if err != nil {
				c.Set("user", guestUser)
			} else {
				c.Set("user", fm.users[claims.Username])
			}
		}
		c.Next()
	}
}

// 登录
func (fm *FileManager) handleLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	user, exists := fm.users[username]
	if !exists {
		c.Redirect(http.StatusSeeOther, "/login?error=<p class='error'>账户不存在或账户名输入错误</p>")
		return
	} else if user.EncryptPassword != fm.hashPassword(password) {
		c.Redirect(http.StatusSeeOther, "/login?error=<p class='error'>密码输入错误</p>")
		return
	}

	token, err := fm.generateJWTToken(username)
	if err != nil {
		fm.log.Warn("生成JWT token失败,报错: ", err)
		c.Redirect(http.StatusSeeOther, "/login?error=<p class='error'>登录失败，请重试</p>")
		return
	}

	c.SetCookie(
		fm.cookieName, // cookie名称
		token,         // token值
		fm.maxAge,     // 过期时间（秒）
		"/",           // 路径
		"",            // 域名
		false,         // 是否仅HTTPS
		true,          // 是否仅HTTP（防止XSS攻击）
	)

	c.Set("user", user)
	fm.log.Infof("%s login success with JWT token", username)

	if c.GetHeader("Content-Type") == "application/json" {
		c.JSON(http.StatusOK, gin.H{
			"token":    token,
			"username": username,
			"message":  "登录成功",
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/file")
}

func (fm *FileManager) handleLogout(c *gin.Context) {
	// 从请求中获取当前token
	tokenString := fm.extractTokenFromRequest(c)

	if tokenString != "" {
		// 解析token获取用户信息用于日志记录
		claims, err := fm.validateJWTToken(tokenString)
		if err == nil {
			fm.log.Infof("%s logout success", claims.Username)
		}
	}

	// 清除cookie中的JWT token
	c.SetCookie(
		fm.cookieName, // cookie名称
		"",            // 空值
		-1,            // 过期时间，-1表示删除cookie
		"/",           // 路径
		"",            // 域名
		false,         // 是否仅HTTPS
		true,          // 是否仅HTTP
	)

	if c.GetHeader("Content-Type") == "application/json" {
		c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
		return
	}

	c.Set("user", guestUser)
	c.Redirect(http.StatusSeeOther, "/file")
}
