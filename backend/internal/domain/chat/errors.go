package chatdomain

import "errors"

var (
	ErrChatNotFound      = errors.New("chat not found")
	ErrChatFinished      = errors.New("chat is already finished")
	ErrChatAccessDenied  = errors.New("access to chat denied")
	ErrOwnerNotFound     = errors.New("chat owner not found: create the user profile first")
	ErrMessageEmpty      = errors.New("message content is empty")
	ErrMessageTooLong    = errors.New("message content is too long")
	ErrTitleTooLong      = errors.New("chat title is too long")
	ErrInvalidSenderType = errors.New("invalid sender type")
	ErrInvalidChatStatus = errors.New("invalid chat status")
	ErrInvalidTransition = errors.New("invalid scenario transition")
	ErrInvalidCursor     = errors.New("invalid cursor")
	ErrScenarioNotFound  = errors.New("scenario not found")
)
