package secrets

import (
	"context"
	"testing"
)

func TestChain_Resolve(t *testing.T) {
	c := Chain{
		StaticResolver{"a": "1"},
		StaticResolver{"b": "2"},
	}
	got, err := c.Resolve(context.Background(), []string{"a", "b", "missing"})
	if err != nil {
		t.Fatal(err)
	}
	if got["a"] != "1" || got["b"] != "2" {
		t.Fatalf("%#v", got)
	}
	if _, ok := got["missing"]; ok {
		t.Fatal("unexpected missing key")
	}
}
