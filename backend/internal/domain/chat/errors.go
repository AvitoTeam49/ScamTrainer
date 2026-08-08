package chatdomain

import "errors"

var (
	ErrChatNotFound     = errors.New("chat not found")
	ErrChatFinished     = errors.New("chat is already finished")
	ErrChatAccessDenied = errors.New("access to chat denied")
	// ErrOwnerNotFound — у аутентифицированного пользователя ещё нет профиля в users.
	ErrOwnerNotFound     = errors.New("chat owner not found: create the user profile first")
	ErrMessageEmpty      = errors.New("message content is empty")
	ErrInvalidSenderType = errors.New("invalid sender type")
	ErrInvalidChatStatus = errors.New("invalid chat status")
	ErrInvalidTransition = errors.New("invalid scenario transition")
	ErrInvalidCursor     = errors.New("invalid cursor")
	ErrScenarioNotFound  = errors.New("scenario not found")
)
