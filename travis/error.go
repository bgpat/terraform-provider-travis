package travis

import "github.com/shuheiktgw/go-travis"

func IsNotFound(err error) bool {
	errResp, ok := err.(*travis.ErrorResponse)
	return ok && errResp.ErrorType == "not_found"
}
