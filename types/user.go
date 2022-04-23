package types

type User struct {
	ID          string
	FullName    string
	SessionInfo SessionInfo
}

type SessionInfo struct {
	PickedImage string
	Points      int
	SessionID   string
}
