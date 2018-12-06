package consts

import "github.com/pkg/errors"

var ErrNYI = errors.Errorf("NYI")

var MaxGRPCMessageSize = 1024 * 1024 * 16 // 16MB

const RefNameMaster = "refs/heads/master"
