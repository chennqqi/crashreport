package raygun

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// Endpoint contains the endpoint of the raygun api
var Endpoint = "https://api.raygun.io"

// Post is the body of a raygun message
type Post struct {
	OccuredOn string  `json:"occurredOn"` // the time the error occured on, format 2006-01-02T15:04:05Z
	Details   Details `json:"details"`    // all the details needed by the API
}

// SetOccurredOn formats the date accordingly
func (p *Post) SetOccurredOn(t time.Time) {
	p.OccuredOn = t.Format("2006-01-02T15:04:05Z")
}

// Submit sends the error to raygun
func Submit(post *Post, key string, client *http.Client) error {
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

// Details contains the info about the circumstances of the error
type Details struct {
	MachineName    string       `json:"machineName"` // the machine's hostname
	Version        string       `json:"version"`     // the version from context
	Client         Client       `json:"client"`      // information on this client
	Error          Error        `json:"error"`       // everything we know about the error itself
	Breadcrumbs    []Breadcrumb `json:"breadcrumbs"` // A list of steps the user made before the crash
	Environment    Environment  `json:"environment"`
	Tags           []string     `json:"tags"`           // the tags from context
	UserCustomData interface{}  `json:"userCustomData"` // the custom data from the context
	Request        Request      `json:"request"`        // the request from the context
	Response       Response     `json:"response"`       // the response from the context
	User           User         `json:"user"`           // the user from the context
	Context        Context      `json:"context"`        // the identifier from the context
}

// Client contains the info about the app generating the error
type Client struct {
	Name      string `json:"identifier"`
	Version   string `json:"version"`
	ClientURL string `json:"clientUrl"`
}

// Error contains the info about the actual error
type Error struct {
	InnerError string              `json:"innerError"` // not really useful in go, but whatever
	Data       interface{}         `json:"data"`       // could be anything
	ClassName  string              `json:"className"`  // not really useful in go, but whatever
	Message    string              `json:"message"`    // This is basically err.Error()
	StackTrace []StackTraceElement `json:"stackTrace"` // the stacktrace of the error
}

// StackTraceElement is one element of the error's stack trace.
type StackTraceElement struct {
	LineNumber  int    `json:"lineNumber"`
	PackageName string `json:"className"`
	FileName    string `json:"fileName"`
	MethodName  string `json:"methodName"`
}

// Breadcrumb is a step
type Breadcrumb struct {
	Message    string      `json:"message"`
	Category   string      `json:"category"`
	CustomData interface{} `json:"customData"`
	Timestamp  int         `json:"timestamp"`
	Level      int         `json:"level"`
	Type       string      `json:"type"`
}

// Environment contains the info about the machine
type Environment struct {
	ProcessorCount          int    `json:"processorCount"`
	OsVersion               string `json:"osVersion"`
	WindowBoundsWidth       int    `json:"windowBoundsWidth"`
	WindowBoundsHeight      int    `json:"windowBoundsHeight"`
	ResolutionScale         string `json:"resolutionScale"`
	CurrentOrientation      string `json:"currentOrientation"`
	CPU                     string `json:"cpu"`
	PackageVersion          string `json:"packageVersion"`
	Architecture            string `json:"architecture"`
	TotalPhysicalMemory     int    `json:"totalPhysicalMemory"`
	AvailablePhysicalMemory int    `json:"availablePhysicalMemory"`
	TotalVirtualMemory      int    `json:"totalVirtualMemory"`
	AvailableVirtualMemory  int    `json:"availableVirtualMemory"`
	DiskSpaceFree           []int  `json:"diskSpaceFree"`
	DeviceName              string `json:"deviceName"`
	Locale                  string `json:"locale"`
}

// Request holds all information on the request from the context
type Request struct {
	HostName    string            `json:"hostName"`
	URL         string            `json:"url"`
	HTTPMethod  string            `json:"httpMethod"`
	IPAddress   string            `json:"ipAddress"`
	QueryString map[string]string `json:"queryString"` // key-value-pairs from the URI parameters
	Form        map[string]string `json:"form"`        // key-value-pairs from a given form (POST)
	Headers     map[string]string `json:"headers"`     // key-value-pairs from the header
	RawData     interface{}       `json:"rawData"`
}

// Response contains the status code
type Response struct {
	StatusCode int `json:"statusCode"`
}

// User holds information on the affected user.
type User struct {
	Identifier string `json:"identifier"`
}

// Context holds information on the program context.
type Context struct {
	Identifier string `json:"identifier"`
}
