<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [ory-libs/env](#ory-libsenv)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# ory-libs/env

Adds defaults to `os.GetEnv()` and saves you 3 lines of code:

```go
import "github.com/ory-am/common/env"

func main() {
  port := env.Getenv("PORT", "80")
}
```

versus

```go
import "os"

func main() {
  port := os.Getenv("PORT")
  if port == "" {
    port = "80"
  }
}
```
