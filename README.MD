# CassieGo Toolkit

[![Go Report Card](https://goreportcard.com/badge/github.com/iameggi/cassie)](https://goreportcard.com/report/github.com/iameggi/cassie)
[![Go.Dev](https://pkg.go.dev/badge/github.com/iameggi/cassie)](https://pkg.go.dev/github.com/iameggi/cassie)

**CassieGo** is a high-performance modular toolkit for building robust and efficient Go web services.

It’s **not a framework** — it’s a collection of production-ready tools designed to work seamlessly with Go’s standard `net/http` package and popular frameworks like **Gin**, **Chi**, **Echo** and *Etc*.

- **Efficient:** Uses safe, generic object pools to reduce memory allocations and GC load.  
- **Resilient:** Provides essential middleware for stability, observability, and fault tolerance.  
- **Modular:** Import only the packages you need.

---

##  Installation 

```bash
go get github.com/iameggi/cassie
```

---

## 🛠️ Available Packages

### 1\. `bucket` (Efficient Pooling)

Provides a safe, type-aware wrapper around `sync.Pool` using Go generics (Go 1.18+).  
This eliminates unsafe type assertions and common pooling bugs.

**Key Features:**
- **Type-Safe:** `Get()` returns a `*T` — no type assertions required.  
- **Auto-Reset:** `Put()` automatically calls your `Reset` function, preventing data leaks between requests.

#### Simple Usage (Built-in Pools)

Cassie provides pre-configured pools for `*bytes.Buffer` and `*strings.Builder`.

```go
import "github.com/iameggi/cassie/bucket"

func MyHandler(w http.ResponseWriter, r *http.Request) {
    // Safe callback pattern
    bucket.WithByteBuffer(func(buf *bytes.Buffer) {
        buf.WriteString("Hello from Cassie pool!")
        w.Write(buf.Bytes())
    })

    // Manual pattern (full control)
    sb := bucket.StringBuilderBucket.Get()
    defer bucket.StringBuilderBucket.Put(sb)
    sb.WriteString("Building strings efficiently...")
}
```

#### Advanced Usage (Custom Pools)

Easily create safe, reusable pools for any struct.

```go
import "github.com/iameggi/cassie/bucket"

// 1MB pool for image processing
var largePool = bucket.NewBytePool(1024 * 1024)

// Custom struct pool
type MyObject struct { ID int }
func resetMyObject(o *MyObject) { o.ID = 0 }

var myPool = bucket.New(func() *MyObject { return &MyObject{} }, resetMyObject)
```

---

### 2\. `middleware` (Resilient Server Layer)

Standard `net/http` middleware compatible with any Go web framework.

#### `middleware.Recovery`

Prevents your server from crashing due to a `panic` in a handler.

```go
import (
    "github.com/iameggi/cassie/middleware"
    "log"
    "os"
)

myHandler := http.HandlerFunc(...)
logger := log.New(os.Stderr, "", log.LstdFlags)

// Wrap your handler to safely catch panics
handler := middleware.Recovery(logger)(myHandler)
http.ListenAndServe(":8080", handler)
```

#### `middleware.Logger`

Structured JSON request logging (requires `zerolog`).

```go
import (
    "github.com/iameggi/cassie/middleware"
    "github.com/rs/zerolog"
    "os"
)

logger := zerolog.New(os.Stdout)
handler := middleware.Logger(logger)(myHandler)
```

#### `middleware.Limiter`

Protects your server from overload by limiting concurrent requests.

```go
import "github.com/iameggi/cassie/middleware"

// Allow only 100 concurrent requests
limiter := middleware.NewLimiter(100)
handler := limiter.Wrap(myHandler)
```

---

### 3\. `helpers` (Simplified Handlers)

Eliminates repetitive boilerplate code inside your handlers.

#### `helpers.SendJSON`

Efficiently sends JSON responses using Cassie’s internal buffer pool.

```go
import "github.com/iameggi/cassie/helpers"

func MyHandler(w http.ResponseWriter, r *http.Request) {
    type User struct {
        Name string `json:"name"`
    }

    // Automatically sets Content-Type, status, and encodes JSON using pooled buffers
    helpers.SendJSON(w, http.StatusOK, User{Name: "Cassie"})
}
```

#### `helpers.SendError`

A simple shortcut for sending structured JSON error responses.

```go
import "github.com/iameggi/cassie/helpers"

func FindUserHandler(w http.ResponseWriter, r *http.Request) {
    // ... user not found ...
    helpers.SendError(w, http.StatusNotFound, "User not found")
}
```

---

##  License

This project is licensed under the [MIT License](https://www.google.com/search?q=LICENSE).

