package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type ContextKey string

const CKisAuthenticated = ContextKey("isAuthenticated")

var ERRCouldNotConvertContextEntry = errors.New("Could not convert context value")

func SetAuthenticatedInContext(r *http.Request, isAuthenticated bool) {

	ctx := r.Context()
	ctx = context.WithValue(ctx, CKisAuthenticated, isAuthenticated)
	r = r.WithContext(ctx)
	fmt.Println(r.Context().Value(CKisAuthenticated).(bool))

}

func GetAuthenticatedFromContext(r *http.Request) (bool, error) {
	isAuthenticated, ok := r.Context().Value(CKisAuthenticated).(bool)
	if !ok {
		return false, ERRCouldNotConvertContextEntry
	}
	return isAuthenticated, nil

}
