# fiberpow

Fiberpow is a [Fiber](https://github.com/gofiber/fiber) middleware, it's aim is blocking (or slowing) bots, by periodically asking clients for a Proof Of Work challenge.

### Config explaination

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

## Installation
```
go get github.com/witer33/fiberpow
```
## Usage
### Basic config
```go
import (
	"github.com/gofiber/fiber/v2"
	"github.com/go-redis/redis/v8"
	"github.com/witer33/fiberpow"
)

func main() {

	app := fiber.New()

	app.Use(fiberpow.New(fiberpow.Config{
		RedisClient: redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0}),
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
	"github.com/go-redis/redis/v8"
)

func main() {

	app := fiber.New()

	app.Use(fiberpow.New(fiberpow.Config{
		PowInterval: 10 * time.Minute,
		Difficulty:  60000,
		Filter: func(c *fiber.Ctx) bool {
			return c.IP() == "127.0.0.1"
		},
		RedisClient: redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0}),
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	app.Listen(":3000")
}
```
