package csrf

import (
	"crypto/subtle"
	"net/http"
	"slices"

	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

const (
	headerCsrfName = "x-csrf-token"
	formCsrfName   = "csrf"
	cookieCsrfName = "whmr_xct"
)

var ignoreMethods = []string{"GET", "HEAD", "OPTIONS", "TRACE"}

// 从header或者form表单中校验csrftoken
//
// 校验原则：cookie中存在csrftoken，并且header或form表单中存在相同的csrftoken
func Validate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if slices.Contains(ignoreMethods, r.Method) {
			next(w, r)
			return
		}
		
		cookie, err := r.Cookie(cookieCsrfName)
		if err != nil || len(cookie.Value) != tokenSize {
			xhttp.Error(r, w, xerror.ErrCsrf)
			return
		}

		cookieCsrf := cookie.Value
		reqCsrf := requestToken(r)
		if !compareTokens(reqCsrf, cookieCsrf) {
			xhttp.Error(r, w, xerror.ErrCsrf)
			return
		}

		next(w, r)
	}
}

func requestToken(r *http.Request) string {
	// get from header
	requestCsrf := r.Header.Get(headerCsrfName)
	if requestCsrf == "" {
		requestCsrf = r.PostFormValue(formCsrfName)
	}

	if requestCsrf == "" && r.MultipartForm != nil {
		vals := r.MultipartForm.Value[formCsrfName]
		if len(vals) > 0 {
			requestCsrf = vals[0]
		}
	}

	return requestCsrf
}

func compareTokens(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	return subtle.ConstantTimeCompare(utils.StringToBytes(a), utils.StringToBytes(b)) == 1
}
