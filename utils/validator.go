package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) bool {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

// ValidatePhone 验证手机号格式（中国手机号）
func ValidatePhone(phone string) bool {
	const phoneRegex = `^1[3-9]\d{9}$`
	re := regexp.MustCompile(phoneRegex)
	return re.MatchString(phone)
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string) (bool, string) {
	if len(password) < 6 {
		return false, "密码长度至少6位"
	}
	
	if len(password) > 32 {
		return false, "密码长度不能超过32位"
	}
	
	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	
	// 至少包含两种字符类型
	typeCount := 0
	if hasUpper {
		typeCount++
	}
	if hasLower {
		typeCount++
	}
	if hasNumber {
		typeCount++
	}
	if hasSpecial {
		typeCount++
	}
	
	if typeCount < 2 {
		return false, "密码需要包含至少两种字符类型（大写字母、小写字母、数字、特殊字符）"
	}
	
	return true, ""
}

// ValidateUsername 验证用户名格式
func ValidateUsername(username string) (bool, string) {
	if len(username) < 3 {
		return false, "用户名长度至少3位"
	}
	
	if len(username) > 20 {
		return false, "用户名长度不能超过20位"
	}
	
	// 只允许字母、数字、下划线
	const usernameRegex = `^[a-zA-Z0-9_]+$`
	re := regexp.MustCompile(usernameRegex)
	if !re.MatchString(username) {
		return false, "用户名只能包含字母、数字和下划线"
	}
	
	// 不能以数字开头
	if unicode.IsNumber(rune(username[0])) {
		return false, "用户名不能以数字开头"
	}
	
	return true, ""
}

// ValidateRequired 验证必填字段
func ValidateRequired(value string, fieldName string) (bool, string) {
	if strings.TrimSpace(value) == "" {
		return false, fieldName + "不能为空"
	}
	return true, ""
}

// ValidateLength 验证字符串长度
func ValidateLength(value string, min, max int, fieldName string) (bool, string) {
	length := len(strings.TrimSpace(value))
	if length < min {
		return false, fieldName + "长度至少" + string(rune(min)) + "位"
	}
	if max > 0 && length > max {
		return false, fieldName + "长度不能超过" + string(rune(max)) + "位"
	}
	return true, ""
}

// ValidateURL 验证URL格式
func ValidateURL(url string) bool {
	if url == "" {
		return true // 空URL也是有效的
	}
	
	const urlRegex = `^https?://[^\s/$.?#].[^\s]*$`
	re := regexp.MustCompile(urlRegex)
	return re.MatchString(url)
}

// ValidateNumericRange 验证数值范围
func ValidateNumericRange(value, min, max float64, fieldName string) (bool, string) {
	if value < min {
		return false, fieldName + "不能小于" + string(rune(int(min)))
	}
	if max > 0 && value > max {
		return false, fieldName + "不能大于" + string(rune(int(max)))
	}
	return true, ""
}

// SanitizeString 清理字符串（移除首尾空格）
func SanitizeString(value string) string {
	return strings.TrimSpace(value)
}

// IsEmptyString 检查字符串是否为空
func IsEmptyString(value string) bool {
	return strings.TrimSpace(value) == ""
}