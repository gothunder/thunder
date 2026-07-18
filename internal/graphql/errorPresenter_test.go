package graphql

import (
	"context"
	"errors"
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func TestErrorPresenterReturnsFreshInternalErrors(t *testing.T) {
	first := errorPresenter(context.Background(), errors.New("resolver failed"))
	first.Path = ast.Path{ast.PathName("stale")}

	second := errorPresenter(context.Background(), errors.New("another resolver failed"))

	if first == second {
		t.Fatal("errorPresenter returned a shared error instance")
	}
	if len(second.Path) != 0 {
		t.Fatalf("fresh error inherited stale path: %#v", second.Path)
	}
	if status, ok := second.Extensions["status"].(int); !ok || status != 500 {
		t.Fatalf("unexpected internal error status: %#v", second.Extensions)
	}
}

func TestRecoverFuncReturnsFreshInternalErrors(t *testing.T) {
	first, ok := recoverFunc(context.Background(), errors.New("first panic")).(*gqlerror.Error)
	if !ok {
		t.Fatal("recoverFunc did not return a GraphQL error")
	}
	first.Path = ast.Path{ast.PathName("stale")}

	second, ok := recoverFunc(context.Background(), errors.New("second panic")).(*gqlerror.Error)
	if !ok {
		t.Fatal("recoverFunc did not return a GraphQL error")
	}

	if first == second {
		t.Fatal("recoverFunc returned a shared error instance")
	}
	if len(second.Path) != 0 {
		t.Fatalf("fresh panic error inherited stale path: %#v", second.Path)
	}
}
