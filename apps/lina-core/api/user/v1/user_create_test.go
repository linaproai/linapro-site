// This file verifies administrator CreateReq validation boundaries for user.

package v1

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/util/gvalid"
)

// TestCreateUserStillRequiresNickname verifies create validation still requires nickname.
func TestCreateUserStillRequiresNickname(t *testing.T) {
	t.Parallel()

	req := CreateReq{
		Username: "zhangsan",
		Password: "123456",
	}
	if err := gvalid.New().Data(req).Run(context.Background()); err == nil {
		t.Fatal("expected user creation without nickname to fail validation")
	}
}
