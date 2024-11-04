// +build !online

package testtag

import (
	"git.grassecon.net/urdt/ussd/remote"
	accountservice "git.grassecon.net/urdt/ussd/internal/testutil/testservice"
)

var (
	AccountService remote.AccountServiceInterface = &accountservice.TestAccountService{}
)
