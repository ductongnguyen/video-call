package signaling

import "errors"

var (
    ErrCallExists         = errors.New("call already exists between these users")
    ErrCallNotFound       = errors.New("call not found")
    ErrInvalidTransition  = errors.New("invalid call state transition")
    ErrPermissionDenied   = errors.New("permission denied for this call")
) 