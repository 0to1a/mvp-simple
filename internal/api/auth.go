package api

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"net/http"
	core "project/internal"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Error message constants
const (
	ErrInvalidBody           = "invalid body"
	ErrFailedToLoadCompanies = "failed to load companies"
)

func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		if v, ok := claims["sub"]; ok {
			n, ok := v.(float64)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid sub in token"})
				return
			}
			c.Set("user_id", int32(n))
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing sub in token"})
			return
		}
		c.Next()
	}
}

type AuthHandler struct{ App *core.App }

// generateOTP creates a 6-digit OTP
func generateOTP() string {
	b := make([]byte, 3)
	rand.Read(b)
	otp := ""
	for _, v := range b {
		otp += fmt.Sprintf("%02d", int(v)%100)
	}
	return otp[:6]
}

// createJWTToken generates a JWT token for the user with company information
func (h *AuthHandler) createJWTToken(userID, companyID int32, isAdmin bool, tokenType string) (string, error) {
	var expiration time.Duration
	if tokenType == "refresh" {
		expiration = 7 * 24 * time.Hour // 7 days for refresh token
	} else {
		expiration = 24 * time.Hour // 24 hours for access token
	}

	claims := jwt.MapClaims{
		"sub":        userID,
		"company_id": companyID,
		"is_admin":   isAdmin,
		"type":       tokenType, // "access" or "refresh"
		"exp":        time.Now().Add(expiration).Unix(),
		"iat":        time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.App.Cfg.JWTSecret))
}

func (h *AuthHandler) LoginRequest(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	// Check if user exists in database
	user, err := h.App.Queries.GetUserByEmail(c, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Create cache key for OTP
	cacheKey := fmt.Sprintf("otp:%s", user.Email)

	// Check if OTP already exists in cache
	if cachedData, exists := h.App.CacheGet(cacheKey); exists {
		if otpData, ok := cachedData.(map[string]interface{}); ok {
			if lastSent, ok := otpData["last_sent"].(time.Time); ok {
				// If OTP was sent less than 1 minute ago, don't send again
				if time.Since(lastSent) < time.Minute {
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error":       "OTP already sent, please wait before requesting again",
						"retry_after": int((time.Minute - time.Since(lastSent)).Seconds()),
					})
					return
				}
			}
		}
	}

	// Generate new OTP
	otp := generateOTP()

	// Store OTP in cache for 15 minutes
	otpData := map[string]interface{}{
		"otp":       otp,
		"email":     user.Email,
		"user_id":   user.ID,
		"last_sent": time.Now(),
	}
	h.App.CacheSet(cacheKey, otpData, 15*time.Minute)

	// Send OTP via email using the email service
	err = h.App.EmailService.SendEmail(user.Email, user.Name, "Your OTP Code", h.createOTPEmailHTML(otp, user.Name))
	if err != nil {
		// Log error but don't fail the request - OTP is still cached
		fmt.Printf("Failed to send OTP email to %s: %v\n", user.Email, err)
		// Also log the OTP for development purposes when email fails
		fmt.Printf("OTP for %s: %s\n", user.Email, otp)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP sent to your email",
		"email":   user.Email,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		OTP   string `json:"otp" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	// Check OTP in cache
	cacheKey := fmt.Sprintf("otp:%s", req.Email)
	cachedData, exists := h.App.CacheGet(cacheKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OTP expired or not found"})
		return
	}

	otpData, ok := cachedData.(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid OTP data"})
		return
	}

	// Validate OTP and email
	storedOTP, otpOk := otpData["otp"].(string)
	storedEmail, emailOk := otpData["email"].(string)
	userID, userOk := otpData["user_id"].(int32)

	if !otpOk || !emailOk || !userOk || storedOTP != req.OTP || storedEmail != req.Email {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid OTP or email"})
		return
	}

	defaultCompany, err := h.App.Queries.GetDefaultUserCompany(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get default company"})
		return
	}

	// Create JWT tokens
	accessToken, err := h.createJWTToken(userID, defaultCompany.CompanyID, defaultCompany.IsAdmin.Bool, "access")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create access token"})
		return
	}

	refreshToken, err := h.createJWTToken(userID, defaultCompany.CompanyID, defaultCompany.IsAdmin.Bool, "refresh")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create refresh token"})
		return
	}

	// Clear OTP from cache after successful login
	h.App.Cache.Delete(cacheKey)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// parseAndValidateRefreshToken validates and parses a refresh token, extracts user ID
func (h *AuthHandler) parseAndValidateRefreshToken(refreshToken string) (jwt.MapClaims, int32, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(h.App.Cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, 0, fmt.Errorf("invalid refresh token")
	}

	// Verify token type
	if tokenType, ok := claims["type"].(string); !ok || tokenType != "refresh" {
		return nil, 0, fmt.Errorf("invalid token type")
	}

	// Extract user ID
	v, ok := claims["sub"]
	if !ok {
		return nil, 0, fmt.Errorf("missing user ID in token")
	}
	n, ok := v.(float64)
	if !ok {
		return nil, 0, fmt.Errorf("invalid user ID in token")
	}

	return claims, int32(n), nil
}

// resolveCompanyAccess determines company ID and admin status based on request and token
func (h *AuthHandler) resolveCompanyAccess(c *gin.Context, requestedCompanyID *int32, claims jwt.MapClaims, userID int32) (int32, bool, error) {
	if requestedCompanyID != nil {
		// Validate user has access to requested company
		companies, err := h.App.Queries.GetUserCompanies(c, userID)
		if err != nil {
			return 0, false, fmt.Errorf(ErrFailedToLoadCompanies)
		}
		for _, comp := range companies {
			if comp.CompanyID == *requestedCompanyID {
				return *requestedCompanyID, comp.IsAdmin.Bool, nil
			}
		}
		return 0, false, fmt.Errorf("user not in specified company")
	}

	// Use company info from token
	v, ok := claims["company_id"]
	if !ok {
		return 0, false, fmt.Errorf("missing company ID in token")
	}
	n, ok := v.(float64)
	if !ok {
		return 0, false, fmt.Errorf("invalid company ID in token")
	}
	companyID := int32(n)

	// Try to get admin status from token, fallback to database
	if v, ok := claims["is_admin"].(bool); ok {
		return companyID, v, nil
	}

	// Lookup admin status from database
	companies, err := h.App.Queries.GetUserCompanies(c, userID)
	if err != nil {
		return 0, false, fmt.Errorf(ErrFailedToLoadCompanies)
	}
	for _, comp := range companies {
		if comp.CompanyID == companyID {
			return companyID, comp.IsAdmin.Bool, nil
		}
	}

	return companyID, false, nil
}

// RefreshToken handles refresh token requests and generates new access tokens
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
		CompanyID    *int32 `json:"company_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidBody})
		return
	}

	// Parse, validate token and extract user ID
	claims, userID, err := h.parseAndValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Resolve company access and admin status
	companyID, isAdmin, err := h.resolveCompanyAccess(c, req.CompanyID, claims, userID)
	if err != nil {
		errorMsg := err.Error()
		switch {
		case errorMsg == "user not in specified company":
			c.JSON(http.StatusForbidden, gin.H{"error": errorMsg})
		case errorMsg == ErrFailedToLoadCompanies:
			c.JSON(http.StatusInternalServerError, gin.H{"error": errorMsg})
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": errorMsg})
		}
		return
	}

	// Generate new tokens
	newAccessToken, err := h.createJWTToken(userID, companyID, isAdmin, "access")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create access token"})
		return
	}

	newRefreshToken, err := h.createJWTToken(userID, companyID, isAdmin, "refresh")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}

// ListCompanies returns all companies for the authenticated user
func (h *AuthHandler) ListCompanies(c *gin.Context) {
	uid, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var userID int32
	if v, ok := uid.(int32); ok {
		userID = v
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id type"})
		return
	}
	companies, err := h.App.Queries.GetUserCompanies(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load companies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"companies": companies})
}

// createOTPEmailHTML creates the HTML content for OTP email
func (h *AuthHandler) createOTPEmailHTML(otpCode, userName string) string {
	name := userName
	if name == "" {
		name = "User"
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Your OTP Code</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            text-align: center;
            background-color: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            margin-bottom: 20px;
        }
        .otp-code {
            font-size: 32px;
            font-weight: bold;
            color: #007bff;
            text-align: center;
            padding: 20px;
            background-color: #e9ecef;
            border-radius: 8px;
            letter-spacing: 4px;
            margin: 20px 0;
        }
        .warning {
            background-color: #fff3cd;
            border: 1px solid #ffeaa7;
            border-radius: 4px;
            padding: 15px;
            margin: 20px 0;
        }
        .footer {
            text-align: center;
            color: #6c757d;
            font-size: 14px;
            margin-top: 30px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>OTP Verification</h1>
    </div>
    
    <p>Hello %s,</p>
    
    <p>You have requested an OTP (One-Time Password) for verification. Please use the code below:</p>
    
    <div class="otp-code">%s</div>
    
    <div class="warning">
        <strong>Important:</strong>
        <ul>
            <li>This code will expire in 15 minutes</li>
            <li>Do not share this code with anyone</li>
            <li>If you didn't request this code, please ignore this email</li>
        </ul>
    </div>
    
    <p>If you have any questions or need assistance, please contact our support team.</p>
    
    <div class="footer">
        <p>This is an automated message, please do not reply to this email.</p>
    </div>
</body>
</html>`, name, otpCode)
}
