package domain

type Cursor struct {
	Limit   int
	AfterID int64
}

func (c Cursor) Validate() error {
	if c.AfterID < 0 || c.Limit <= 0 {
		return ErrInvalidCursor
	}
	return nil
}
