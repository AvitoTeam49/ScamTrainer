package domain

import "errors"

var (
	ErrChatNotFound        = errors.New("chat not found")
	ErrChatFinished        = errors.New("chat is already finished")
	ErrChatAccessDenied    = errors.New("access to chat denied")
	ErrMessageEmpty        = errors.New("message content is empty")
	ErrInvalidSenderType   = errors.New("invalid sender type")
	ErrInvalidChatStatus   = errors.New("invalid chat status")
	ErrInvalidIncidentKind = errors.New("invalid incident kind")
	ErrInvalidCursor       = errors.New("invalid cursor")
	ErrScenarioNotFound    = errors.New("scenario not found")
)
