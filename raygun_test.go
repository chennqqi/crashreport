package raygun_test

import (
	"errors"
	"testing"

	"github.com/codeclysm/raygun"
	"github.com/davecgh/go-spew/spew"
	jujuerr "github.com/juju/errors"
	pkerr "github.com/pkg/errors"
)

type custErr string

func (e custErr) Error() string {
	return string(e)
}

func (e custErr) Cause() error {
	return errors.New("cause of: " + e.Error())
}

func TestFromErr(t *testing.T) {
	var err error
	var rayErr raygun.Error

	// pkg/errors
	err = pkerr.New("new error")
	rayErr = raygun.FromErr(wrapErr(err))

	if rayErr.Message != "wrapped err: new error" {
		t.Error("rayErr.Message should be 'wrapped err: new error'")
	}
	if rayErr.InnerError != "new error" {
		t.Error("rayErr.InnerError should be 'new error'")
	}

	if len(rayErr.StackTrace) != 4 {
		t.Error("rayErr.StackTrace should be 4 elements long")
	}

	spew.Dump(rayErr.StackTrace)

	// juju/errors
	err = jujuerr.New("new error")
	rayErr = raygun.FromErr(annotateErr(err))

	if rayErr.Message != "wrapped err: new error" {
		t.Error("rayErr.Message should be 'wrapped err: new error'")
	}
	if rayErr.InnerError != "new error" {
		t.Error("rayErr.InnerError should be 'new error'")
	}

	if len(rayErr.StackTrace) != 2 {
		t.Error("rayErr.StackTrace should be 2 elements long")
	}

	// standard error
	rayErr = raygun.FromErr(errors.New("new error"))

	if rayErr.Message != "new error" {
		t.Error("rayErr.Message should be 'new error'")
	}

	if len(rayErr.StackTrace) != 3 {
		t.Error("rayErr.StackTrace should be 4 elements long")
	}
}

func wrapErr(err error) error {
	return pkerr.Wrap(err, "wrapped err")
}

func annotateErr(err error) error {
	return jujuerr.Annotate(err, "wrapped err")
}
