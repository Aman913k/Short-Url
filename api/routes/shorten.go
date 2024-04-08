package routes

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Aman913k/url-shortner/database"
	"github.com/Aman913k/url-shortner/helpers"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *gin.Context) {

	body := new(request)

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "cannot parse JSON"})
		return
	}

	// implement rate limiting

	r2 := database.CreateClient(1)
	fmt.Println(r2)
	defer r2.Close()

	fmt.Println(c.ClientIP())

	// Check rate limit
	val, err := r2.Get(database.Ctx, c.ClientIP()).Result()
	fmt.Println(val)
	if err == redis.Nil {
		_ = r2.Set(database.Ctx, c.ClientIP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else {
		val, _ = r2.Get(database.Ctx, c.ClientIP()).Result()
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := r2.TTL(database.Ctx, c.ClientIP()).Result()
			reset := (limit / time.Nanosecond / time.Minute) // Calculate reset time in minutes

			if reset <= 0 {
				// If reset time is zero or negative, reset the rate limit
				_ = r2.Set(database.Ctx, c.ClientIP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
			} else {
				c.JSON(503, gin.H{
					"error":            "Rate limit exceeded",
					"rate_limit_reset": reset,
				})
				return
			}
		}
	}

	// check if the input sent by the user is an actual URL
	if !govalidator.IsURL(body.URL) {
		c.JSON(400, gin.H{"error": "Invalid Error"})
		return
	}

	// check for domain error
	if !helpers.RemoveDomainError(body.URL) {
		c.JSON(503, gin.H{"error": "You Can't"})
		return
	}

	// enforce https, SSL
	body.URL = helpers.EnforceHTTP(body.URL)

	var id string

	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(0)
	defer r.Close()

	val, _ = r.Get(database.Ctx, id).Result()
	if val != "" {
		c.JSON(403, gin.H{"error": "URL custom short is already in use"})
		return
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	err = r.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()

	if err != nil {
		c.JSON(500, gin.H{"error": "unable to connect to the Server"})
		return
	}

	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  30,
		XRateLimitReset: 30,
	}

	r2.Decr(database.Ctx, c.ClientIP())

	val, _ = r2.Get(database.Ctx, c.ClientIP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)
	ttl, _ := r2.TTL(database.Ctx, c.ClientIP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	c.JSON(200, resp)
}
