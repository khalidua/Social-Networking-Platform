package domain

type Notification struct {
	ID        string
	UserID    string
	Type      string
	Message   string
	Read      bool
	CreatedAt string
}
