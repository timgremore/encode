package main

import (
  // "flag"
  // "github.com/codegangsta/cli"
  "os"
  "testing"
)

type testpair struct {
  input string
  errorExpected bool
}

var testPaths = []testpair{
  {"/my/test/path", false},
  {"/", false},
  {"does-not-exist", true},
}

var testDestinationPaths = []testpair{
  {"/my-test-go-directory", false},
}

func TestVerifyPath(t *testing.T) {
  for _, pair := range testPaths {
    // set := flag.NewFlagSet("test", 0)
    // set.String("path", pair.input, "--path ~/home")

    // context := cli.NewContext(nil, set, set)

    // Initialize path with value of path flag
    path := pair.input

    // Verify that the path is valid
    path, err := verifyPath(path)

    // If this pair is invalid, expect an error to have been returned from verifyPath
    if pair.errorExpected {
      if err == nil {
        t.Error("Expected an error for path " + pair.input + " but received nil.")
      }
    } else {
      if err != nil {
        t.Error("Did not expect an error but received one for path " + path)
      }
    }
  }
}

func TestCreateDirectory(t *testing.T) {
  for _, pair := range testDestinationPaths {

    // Create the directory but do not force an overwrite
    _, err := createDirectory(pair.input, false)

    if os.IsNotExist(err) {
      t.Error("Expected " + pair.input + " to be created but was not.")
    }

    // Create the directory and force an overwrite if one already exists
    _, err = createDirectory(pair.input, true)

    if os.IsNotExist(err) {
      t.Error("Expected " + pair.input + " to be created but was not.")
    }
  }
}
