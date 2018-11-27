package handler

import "github.com/pkg/errors"

var ErrNYI = errors.Errorf("NYI")

var MaxGRPCMessageSize = 1024 * 1024 * 16 // 16MB

const refNameMaster = "refs/heads/master"
