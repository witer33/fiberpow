package fiberpow

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

//go:embed views/challenge.html
var challengeTemplate string

const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Generates a random string using crypto/rand.
func randString(length int) (string, error) {
	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = chars[b%byte(len(chars))]
	}

	return string(bytes), nil
}

// Config is the Middleware Config.
type Config struct {
	Filter      func(*fiber.Ctx) bool
	PowInterval time.Duration
	Difficulty  int
	RedisClient *redis.Client
}

// this stores individual information for an ip address.
type ipStatus struct {
	// if verified is true skip the pow check
	Verified bool `json:"verified"`
	// secretNumber is the number that the user has to find
	SecretNumber int `json:"secretNumber"`
	// secretSuffix is the hash suffix
	SecretSuffix string `json:"secretSuffix"`
	// hash is the hash given to the user
	Hash string `json:"hash"`
}

// New initialize the middleware with the default config or a custom one.
func New(config ...Config) fiber.Handler {
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}

	// Set default config.

	if cfg.PowInterval == 0 {
		cfg.PowInterval = time.Second * 20
	}

	if cfg.Difficulty == 0 {
		cfg.Difficulty = 30000
	}

	if cfg.RedisClient == nil {
		panic("RedisClient is required")
	}

	ctx := context.Background()

	// Middleware handler.
	return func(c *fiber.Ctx) error {
		// Handles Config.Filter.
		if cfg.Filter != nil && cfg.Filter(c) {
			return c.Next()
		}

		// Checks if ipStatus is already present for this ip.
		statusInterface, err := cfg.RedisClient.Get(ctx, c.IP()).Result()
		var status ipStatus

		if err == redis.Nil {
			// Generates a new ipStatus for this ip.
			secretNumber, err := rand.Int(rand.Reader, big.NewInt(int64(cfg.Difficulty)))
			if err != nil {
				return err
			}

			secretSuffix, err := randString(32)

			if err != nil {
				return err
			}

			status = ipStatus{
				Verified:     false,
				SecretNumber: int(secretNumber.Int64()),
				SecretSuffix: secretSuffix,
			}

			// Generates the hash for the user.
			byteHash := sha256.Sum256([]byte(fmt.Sprintf("%d-%s", status.SecretNumber, status.SecretSuffix)))
			status.Hash = hex.EncodeToString(byteHash[:])
			encodedStatus, err := json.Marshal(status)

			if err != nil {
				return err
			}

			err = cfg.RedisClient.Set(ctx, c.IP(), string(encodedStatus), cfg.PowInterval).Err()
			if err != nil {
				return err
			}
		} else if err == nil {
			err := json.Unmarshal([]byte(statusInterface), &status)
			if err != nil {
				return err
			}
		} else {
			return err
		}

		// Skips already verified ip.
		if status.Verified {
			return c.Next()
		}

		// Checks if the user already solved the challenge.
		secretCookie := c.Cookies("_challenge_n", "-1")
		secretNumber, err := strconv.Atoi(secretCookie)
		if err != nil {
			secretNumber = -1
		}

		if secretNumber == status.SecretNumber {
			status.Verified = true
			return c.Next()
		}

		// Renders the challenge template.
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.SendString(fmt.Sprintf(challengeTemplate, status.Hash, status.SecretSuffix))
	}
}
