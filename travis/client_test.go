package travis

import (
	"net/http"
	"testing"

	"github.com/shuheiktgw/go-travis"
)

var IsNotFound = isNotFound

func TestIsAlreadySyncing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "already_syncing",
			err: &travis.ErrorResponse{
				Response:  &http.Response{StatusCode: http.StatusConflict},
				ErrorType: "already_syncing",
			},
			want: true,
		},
		{
			name: "not_found",
			err: &travis.ErrorResponse{
				Response:  &http.Response{StatusCode: http.StatusNotFound},
				ErrorType: "not_found",
			},
			want: false,
		},
		{
			name: "nil",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isAlreadySyncing(tt.err); got != tt.want {
				t.Fatalf("isAlreadySyncing() = %v, want %v", got, tt.want)
			}
		})
	}
}
