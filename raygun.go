// Package raygun is an unofficial library to send crash reporting to [Raygun](https://raygun.com/).
//
// Crash reporting to Raygun is extremely simple: [It's an authenticated http POST with json data](https://raygun.com/raygun-providers/rest-json-api).
//
// This library provides the struct and some helper methods to help fill them with data.
package raygun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/pkg/errors"
)

// Endpoint contains the endpoint of the raygun api. You can change it for testing purposes.
var Endpoint = "https://api.raygun.io"

// Post is the full body of a raygun message. See https://raygun.com/raygun-providers/rest-json-api
type Post struct {
	OccuredOn string  `json:"occurredOn,omitempty"` // the time the error occured on, format 2006-01-02T15:04:05Z
	Details   Details `json:"details,omitempty"`    // all the details needed by the API
}

// Details contains the info about the circumstances of the error
type Details struct {
	MachineName    string       `json:"machineName,omitempty"` // the machine's hostname
	Version        string       `json:"version,omitempty"`     // the version from context
	Client         Client       `json:"client,omitempty"`      // information on this client
	Error          Error        `json:"error,omitempty"`       // everything we know about the error itself
	Breadcrumbs    []Breadcrumb `json:"breadcrumbs,omitempty"` // A list of steps the user made before the crash
	Environment    Environment  `json:"environment,omitempty"`
	Tags           []string     `json:"tags,omitempty"`           // the tags from context
	UserCustomData interface{}  `json:"userCustomData,omitempty"` // the custom data from the context
	Request        Request      `json:"request,omitempty"`        // the request from the context
	Response       Response     `json:"response,omitempty"`       // the response from the context
	User           User         `json:"user,omitempty"`           // the user from the context
	Context        Context      `json:"context,omitempty"`        // the identifier from the context
}

// Client contains the info about the app generating the error
type Client struct {
	Name      string `json:"identifier,omitempty"`
	Version   string `json:"version,omitempty"`
	ClientURL string `json:"clientUrl,omitempty"`
}

// Error contains the info about the actual error
type Error struct {
	InnerError string      `json:"innerError,omitempty"` // not really useful in go, but whatever
	Data       interface{} `json:"data,omitempty"`       // could be anything
	ClassName  string      `json:"className,omitempty"`  // not really useful in go, but whatever
	Message    string      `json:"message,omitempty"`    // This is basically err.Error()
	StackTrace StackTrace  `json:"stackTrace,omitempty"` // the stacktrace of the error
}

func (e Error) Error() string {
	return e.Message
}

// StackTrace implements the interface to add a new stack element
type StackTrace []StackTraceElement

// AddEntry adds a new line to the stacktrace
func (s *StackTrace) AddEntry(lineNumber int, packageName, fileName, methodName string) {
	*s = append(*s, StackTraceElement{lineNumber, packageName, fileName, methodName})
}

func (s *StackTrace) String() string {
	out := ""
	for _, line := range *s {
		out = out + fmt.Sprintf("%s/%s:%d\n", line.PackageName, line.FileName, line.LineNumber)
	}
	return out
}

// StackTraceElement is one element of the error's stack trace.
type StackTraceElement struct {
	LineNumber  int    `json:"lineNumber,omitempty"`
	PackageName string `json:"className,omitempty"`
	FileName    string `json:"fileName,omitempty"`
	MethodName  string `json:"methodName,omitempty"`
}

// Breadcrumb is a step that the user did in the application. See https://raygun.com/thinktank/suggestion/4228
type Breadcrumb struct {
	Message    string      `json:"message,omitempty"`
	Category   string      `json:"category,omitempty"`
	CustomData interface{} `json:"customData,omitempty"`
	Timestamp  int         `json:"timestamp,omitempty"`
	Level      int         `json:"level,omitempty"`
	Type       string      `json:"type,omitempty"`
}

// Environment contains the info about the machine
type Environment struct {
	ProcessorCount          int    `json:"processorCount,omitempty"`
	OsVersion               string `json:"osVersion,omitempty"`
	WindowBoundsWidth       int    `json:"windowBoundsWidth,omitempty"`
	WindowBoundsHeight      int    `json:"windowBoundsHeight,omitempty"`
	ResolutionScale         string `json:"resolutionScale,omitempty"`
	CurrentOrientation      string `json:"currentOrientation,omitempty"`
	CPU                     string `json:"cpu,omitempty"`
	PackageVersion          string `json:"packageVersion,omitempty"`
	Architecture            string `json:"architecture,omitempty"`
	TotalPhysicalMemory     int    `json:"totalPhysicalMemory,omitempty"`
	AvailablePhysicalMemory int    `json:"availablePhysicalMemory,omitempty"`
	TotalVirtualMemory      int    `json:"totalVirtualMemory,omitempty"`
	AvailableVirtualMemory  int    `json:"availableVirtualMemory,omitempty"`
	DiskSpaceFree           []int  `json:"diskSpaceFree,omitempty"`
	DeviceName              string `json:"deviceName,omitempty"`
	Locale                  string `json:"locale,omitempty"`
}

// Request holds all information on the request from the context
type Request struct {
	HostName    string            `json:"hostName,omitempty"`
	URL         string            `json:"url,omitempty"`
	HTTPMethod  string            `json:"httpMethod,omitempty"`
	IPAddress   string            `json:"ipAddress,omitempty"`
	QueryString map[string]string `json:"queryString,omitempty"` // key-value-pairs from the URI parameters
	Form        map[string]string `json:"form,omitempty"`        // key-value-pairs from a given form (POST)
	Headers     map[string]string `json:"headers,omitempty"`     // key-value-pairs from the header
	RawData     interface{}       `json:"rawData,omitempty"`
}

// Response contains the status code
type Response struct {
	StatusCode int `json:"statusCode,omitempty"`
}

// User holds information on the affected user.
type User struct {
	Identifier string `json:"identifier,omitempty"`
}

// Context holds information on the program context.
type Context struct {
	Identifier string `json:"identifier,omitempty"`
}

// NewPost creates a new post by collecting data about the system, such as the current timestamp, os version and architecture, and number of cpus
func NewPost() Post {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "not available"
	}

	post := Post{
		OccuredOn: time.Now().Format("2006-01-02T15:04:05Z"),
		Details: Details{
			MachineName: hostname,
			Environment: Environment{
				ProcessorCount: runtime.NumCPU(),
				OsVersion:      runtime.GOOS,
				Architecture:   runtime.GOARCH,
			},
		},
	}

	return post
}

// FromErr creates an error struct from an error
// If the error satisfies the interfaces `Class() string` and/or `Data() interface{}` it will use them to construct the
// Error struct.
// FromErr also constructs a stacktrace. It the error satisfies the interface `Stacktrace() []string` it will use that.
// Otherwise it will use the runtime package to retrieve the goroutine stacktrace
func FromErr(err error) Error {
	// If it's already a raygun error, don't do anything
	if e, ok := err.(Error); ok {
		return e
	}

	rayerr := Error{
		// InnerError: cause(err).Error(),  // Disabled because it causes strange things
		Message:    err.Error(),
		ClassName:  class(err),
		Data:       data(err),
		StackTrace: stacktrace(err),
	}

	return rayerr
}

// FromReq returns a Request struct from a http request. Rawdata is set to the content of Body
func FromReq(req *http.Request) Request {
	body, _ := ioutil.ReadAll(req.Body)

	request := Request{
		HostName:    req.Host,
		URL:         req.URL.String(),
		HTTPMethod:  req.Method,
		IPAddress:   req.RemoteAddr,
		QueryString: arrayMapToStringMap(req.URL.Query()),
		Form:        arrayMapToStringMap(req.PostForm),
		Headers:     arrayMapToStringMap(req.Header),
		RawData:     body,
	}

	return request
}

// Submit sends the error to raygun. If the client is nil it will use a default one with a 5s timeout
func Submit(post Post, key string, client *http.Client) error {
	json, err := json.Marshal(post)
	if err != nil {
		return errors.Wrapf(err, "convert to json")
	}

	r, err := http.NewRequest("POST", Endpoint+"/entries", bytes.NewBuffer(json))
	if err != nil {
		return errors.Wrapf(err, "create req")
	}
	r.Header.Add("X-ApiKey", key)

	// Default client has 5s timeout
	if client == nil {
		client = &http.Client{
			Timeout: 5 * time.Second,
		}
	}

	resp, err := client.Do(r)
	if err != nil {
		return errors.Wrapf(err, "execute req")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 202 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			body = []byte("no body")
		}

		return errors.New("unexpected answer '" + resp.Status + "' from Raygun: " + string(body))
	}

	return nil
}
