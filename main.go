package fiberpow

import (
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/patrickmn/go-cache"
	"math/big"
	"strconv"
	"time"
)

//go:embed views/challenge.html
var challengeTemplate string

const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

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

type Config struct {
	Filter      func(*fiber.Ctx) bool
	PowInterval time.Duration
	Difficulty  int
}

type ipStatus struct {
	verified     bool
	secretNumber int
	secretSuffix string
	hash         string
}

func New(config ...Config) fiber.Handler {
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.PowInterval == 0 {
		cfg.PowInterval = time.Second * 20
	}

	if cfg.Difficulty == 0 {
		cfg.Difficulty = 30000
	}

	ipCache := cache.New(cfg.PowInterval, 10*time.Second)

	return func(c *fiber.Ctx) error {
		if cfg.Filter != nil && cfg.Filter(c) {
			return c.Next()
		}

		statusInterface, found := ipCache.Get(c.IP())
		var status *ipStatus

		if !found {
			secretNumber, err := rand.Int(rand.Reader, big.NewInt(int64(cfg.Difficulty)))
			if err != nil {
				return err
			}

			secretSuffix, err := randString(32)

			if err != nil {
				return err
			}

			status = &ipStatus{
				verified:     false,
				secretNumber: int(secretNumber.Int64()),
				secretSuffix: secretSuffix,
			}

			byteHash := sha256.Sum256([]byte(fmt.Sprintf("%d-%s", status.secretNumber, status.secretSuffix)))
			status.hash = hex.EncodeToString(byteHash[:])
			ipCache.Set(c.IP(), status, cache.DefaultExpiration)
		} else {
			status = statusInterface.(*ipStatus)
		}

		if status.verified {
			return c.Next()
		}

		secretCookie := c.Cookies("_challenge_n", "-1")
		secretNumber, err := strconv.Atoi(secretCookie)
		if err != nil {
			secretNumber = -1
		}

		if secretNumber == status.secretNumber {
			status.verified = true
			return c.Next()
		}

		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.SendString(fmt.Sprintf(challengeTemplate, status.hash, status.secretSuffix))
	}
}
