package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vnestcc/dashboard/utils/values"
)

type Claims struct {
	ID   uint   `json:"id"`
	Role string `json:"roles"`
	jwt.RegisteredClaims
}

var JWTVerifyHandler gin.HandlerFunc = func(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
		return
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
		return
	}
	tokenString := parts[1]
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(values.GetConfig().Server.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		ctx.Set("message", err.Error())
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	ctx.Set("claims", claims)
	ctx.Next()
}

func RoleCheckHandler(roles ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claimsVal, exists := ctx.Get("claims")
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No claims found"})
			return
		}
		claims, ok := claimsVal.(*Claims)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Invalid claims type"})
			return
		}
		for _, required := range roles {
			if claims.Role == required {
				ctx.Next()
				return
			}
		}
	}
}

var UserMiddleware gin.HandlersChain = gin.HandlersChain{
	JWTVerifyHandler,
	RoleCheckHandler("user"),
}
var AdminMiddleware gin.HandlersChain = gin.HandlersChain{
	JWTVerifyHandler,
	RoleCheckHandler("admin"),
}
var ModeratorMiddleware gin.HandlersChain = gin.HandlersChain{
	JWTVerifyHandler,
	RoleCheckHandler("moderator", "admin"),
}
var VCMiddleware gin.HandlersChain = gin.HandlersChain{
	JWTVerifyHandler,
	RoleCheckHandler("vc"),
}
