package handlers

import (
	"net/http"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Middleware to authorize a requests requiring a logged in user
func Authorize(address string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Redirect link for errors
		link := "http://" + address
		
		s := sessions.Default(c)
		valid := s.Get("user-id")
		if valid == nil {
			c.HTML(http.StatusUnauthorized, "error.tmpl", gin.H{"message": "You are not authorized to access this page. Please login.", "link": link})
			c.Abort()
		}
		c.Next()
	}
}
