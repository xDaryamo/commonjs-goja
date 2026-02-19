package commonjs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/tliron/commonjs-goja"
	"github.com/tliron/commonjs-goja/api"
	"github.com/tliron/exturl"

	_ "github.com/tliron/commonlog/simple"
)

func TestEnvironment(t *testing.T) {
	urlContext := exturl.NewContext()
	defer urlContext.Release()

	path := filepath.Join(getRoot(t), "examples")

	environment := commonjs.NewEnvironment(urlContext, urlContext.NewFileURL(path))
	defer environment.Release()

	environment.Extensions = api.DefaultExtensions{}.Create()

	testEnvironment(t, environment)
}

func TestUnwrapJavaScriptExceptionInfiniteLoop(t *testing.T) {
	vm := goja.New()
	_, err := vm.RunString(`throw "some string error"`)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}

	// We use a channel to detect the infinite loop via timeout
	done := make(chan bool)
	go func() {
		_ = commonjs.UnwrapJavaScriptException(err)
		done <- true
	}()

	select {
	case <-done:
		// Success (no infinite loop)
		expected := "some string error"
		if unwrapped := commonjs.UnwrapJavaScriptException(err); unwrapped.Error() != expected {
			t.Errorf("Expected error message %q, but got %q", expected, unwrapped.Error())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("UnwrapJavaScriptException timed out, likely an infinite loop")
	}
}

func TestUnwrapJavaScriptExceptionErrorObject(t *testing.T) {
	vm := goja.New()
	_, err := vm.RunString(`throw new Error("js error")`)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}

	expected := "js error"
	if unwrapped := commonjs.UnwrapJavaScriptException(err); !strings.Contains(unwrapped.Error(), expected) {
		t.Errorf("Expected error message to contain %q, but got %q", expected, unwrapped.Error())
	}
}

func testEnvironment(t *testing.T, environment *commonjs.Environment) {
	// Start!
	if _, err := environment.Require("./start", false, nil); err != nil {
		t.Errorf("%s", err)
	}
}

func getRoot(t *testing.T) string {
	var root string
	var ok bool
	if root, ok = os.LookupEnv("COMMONJS_TEST_ROOT"); !ok {
		var err error
		if root, err = os.Getwd(); err != nil {
			t.Errorf("os.Getwd: %s", err.Error())
		}
	}
	return root
}
