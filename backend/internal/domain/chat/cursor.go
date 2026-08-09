package chatdomain

const MaxCursorLimit = 100

type Cursor struct {
	Limit   int
	AfterID int64
}

func (c Cursor) Validate() error {
	if c.AfterID < 0 || c.Limit <= 0 || c.Limit > MaxCursorLimit {
		return ErrInvalidCursor
	}
	return nil
}
