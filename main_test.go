package fiberpow

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	app := fiber.New()

	rand.Seed(time.Now().UnixNano())

	app.Use(New(Config{
		Difficulty: 20000,
	}))

	taskCount := uint32(0)
	app.Get("/", func(c *fiber.Ctx) error {
		seconds := rand.Intn(10)
		fmt.Printf("Waiting %d seconds with %d tasks\n", seconds, atomic.LoadUint32(&taskCount))
		atomic.AddUint32(&taskCount, 1)
		time.Sleep(time.Duration(seconds) * time.Second)
		return c.SendString(fmt.Sprintf("done in %d seconds!", seconds))
	})

	app.Listen(":3000")
}
