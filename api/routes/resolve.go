package routes

import (
	"github.com/Aman913k/url-shortner/database"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func ResolveURL(c *gin.Context) {
	url := c.Param("url")

	r := database.CreateClient(0)
	defer r.Close()

	value, err := r.Get(database.Ctx, url).Result()
	if err == redis.Nil {
		c.JSON(404, gin.H{"error": "short not found in the database"})

	} else if err != nil {
		c.JSON(500, gin.H{"error": "cannot connect to DB"})
		return
	}

	rInr := database.CreateClient(1)
	defer rInr.Close()

	_ = rInr.Incr(database.Ctx, "counter")

	c.Redirect(301, value)

}
