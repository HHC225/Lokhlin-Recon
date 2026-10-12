package middleware

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
)

var responseLanguages = []language.Tag{language.MustParse("ko-KR"), language.MustParse("en-US")}
var responseLanguageMatcher = language.NewMatcher(responseLanguages)

func I18nMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		selected := responseLanguages[0]
		if tags, _, err := language.ParseAcceptLanguage(c.GetHeader("Accept-Language")); err == nil && len(tags) != 0 {
			_, index, _ := responseLanguageMatcher.Match(tags...)
			selected = responseLanguages[index]
		}
		c.Set("language", selected.String())
		c.Header("Content-Language", selected.String())
		c.Next()
	}
}
