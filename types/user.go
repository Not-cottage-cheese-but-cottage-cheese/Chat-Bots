package types

type User struct {
	ID          string
	SessionInfo SessionInfo
}

type SessionInfo struct {
	PickedImage string
	Points      int
	SessionID   string
}
