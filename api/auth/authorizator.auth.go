package auth

import (
	"fmt"
	"strings"

	"github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

func allowed(level string, claims map[string]interface{}, c *gin.Context) bool {
	enrolledCourses := claims["courses"].(map[string]interface{})
	uid := claims["uid"]
	cid, _ := c.Get("cids")
	sid, exists := c.Get("sid")
	allowed := false

	val, found := enrolledCourses[cid.(string)]
	fmt.Println("cid:", cid, level, val)
	if found && (level == "any" || val == level) {
		fmt.Println("yooooo")
		return true
	}

	if level == "student" && exists {
		sub, err := sm.GetUsersSubmission(sid, uid)
		if err != nil {
			fmt.Println("err", err)
		}
		if sub != nil {
			allowed = true
		}
	}

	return allowed
}

func determineLevel(route string) string {
	if _, found := routeLevels["admin"][route]; found {
		return "admin"
	}

	if _, found := routeLevels["any"][route]; found {
		return "any"
	}

	if _, found := routeLevels["assitant"][route]; found {
		return "assitant"
	}

	if _, found := routeLevels["teacher"][route]; found {
		return "teacher"
	}

	if _, found := routeLevels["student"][route]; found {
		return "student"
	}

	return "whitelisted"
}

// Authorizator a default function for a gin jwt, that authorizes a user.
func Authorizator(d interface{}, c *gin.Context) bool {
	route := strings.TrimPrefix(c.Request.URL.String(), "/api/v1/plague_doctor/")
	for _, p := range c.Params {
		route = strings.Replace(route, p.Value, ":"+p.Key, 1)
	}

	claims := jwt.ExtractClaims(c)
	uids := claims["uid"].(string)
	val, _ := primitive.ObjectIDFromHex(uids)
	c.Set("uid", val)

	userLevelForRouteShouldBe := determineLevel(route)
	fmt.Println("user level:", userLevelForRouteShouldBe, route)
	if userLevelForRouteShouldBe == "whitelisted" {
		return true
	}

	admin := claims["admin"].(bool)
	if userLevelForRouteShouldBe == "admin" && admin {
		return true
	} else if userLevelForRouteShouldBe == "admin" && !admin {
		return false
	}

	return allowed(userLevelForRouteShouldBe, claims, c)
}
