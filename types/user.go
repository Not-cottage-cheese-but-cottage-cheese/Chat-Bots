package types

type User struct {
	ID          string
	FullName    string
	SessionInfo SessionInfo
}

type SessionInfo struct {
	PickedImage int
	Points      int
	SessionID   string
}
