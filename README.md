Raygun
======

An unofficial library to send crash reporting to [Raygun](https://raygun.com/).

Crash reporting to Raygun is extremely simple: [It's an authenticated http POST with json data](https://raygun.com/raygun-providers/rest-json-api).

This library provides the struct and some helper methods to help fill them with data.

[Godoc documentation](https://godoc.org/github.com/codeclysm/raygun)

# Recover panics
The classic way to use raygun is to recover panics and send info to the raygun server:

```go
import "github.com/codeclysm/raygun"

func main() {
	defer func() {
		if e := recover(); e != nil {
			post := raygun.NewPost()
			post.Details.Error = raygun.FromErr(err)

			raygun.Submit(post, "thisismysecretkey", nil)
		}
	}

	panic("This is a panic")
}
```

# Customize error report
A `raygun.Post` is just a struct, so you can edit all the fields before sending it. You can fill info about a Request, or about the Window size:

```go

	post := raygun.NewPost()
	post.Details.Request = raygun.Request{}
	post.Details.Environment = raygun.Environment{
		WindowBoundsWidth: 500
	}
```

# Helpers
`raygun.FromErr` and `raygun.FromReq` build info from errors and requests. `FromErr` also includes a stacktrace, compatible with the stacktraces from https://github.com/pkg/errors and https://github.com/juju/errors:

```go
	err := raygun.FromErr(errors.New("new error))
```

```json
{
    "message":"new error",
    "stackTrace":[
        {
            "lineNumber":77,
            "className":"github.com/codeclysm/raygun_test",
            "fileName":"/opt/workspace/github.com/codeclysm/raygun/raygun_test.go",
            "methodName":"wrapErr"
        },
        {
            "lineNumber":31,
            "className":"github.com/codeclysm/raygun_test",
            "fileName":"/opt/workspace/github.com/codeclysm/raygun/raygun_test.go",
            "methodName":"TestFromErr"
        },
        {
            "lineNumber":657,
            "className":"testing",
            "fileName":"/usr/local/go/src/testing/testing.go",
            "methodName":"tRunner"
        },
        {
            "lineNumber":2197,
            "className":"runtime",
            "fileName":"/usr/local/go/src/runtime/asm_amd64.s",
            "methodName":"goexit"
        }
    ]
}
```