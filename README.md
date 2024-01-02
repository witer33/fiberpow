# fiberpow

Fiberpow is a [Fiber](https://github.com/gofiber/fiber) middleware, it aims to block (or at least slow down) bots, by periodically asking clients for a proof of work challenge.

## Config explaination
#### Difficulty
```go
int
```
Maximum number of calculated hashes by the client, default: 30000.

#### PowInterval
```go
time.Duration
```
Interval between challenges for the same IP.

#### Filter
```go
func(c *fiber.Ctx) bool
```
Use this if you need to skip the PoW challenge in certain conditions, true equals skip.

#### Storage
```go
fiber.Storage
```
Database used to keep track of challenges, for reference use https://github.com/gofiber/storage.

## Installation
```
go get github.com/witer33/fiberpow
```
## Usage
### Basic config
```go
import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/redis/v3"
	"github.com/witer33/fiberpow"
)

func main() {
	app := fiber.New()

	app.Use(fiberpow.New(fiberpow.Config{
		Storage: redis.New(),
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	app.Listen(":3000")
}
```
### Custom config
```go
import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/witer33/fiberpow"
	"github.com/gofiber/storage/redis/v3"
)

func main() {
	app := fiber.New()

	app.Use(fiberpow.New(fiberpow.Config{
		PowInterval: 10 * time.Minute,
		Difficulty:  60000,
		Filter: func(c *fiber.Ctx) bool {
			return c.IP() == "127.0.0.1"
		},
		Storage: redis.New(),
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	app.Listen(":3000")
}
```
