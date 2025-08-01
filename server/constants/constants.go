package constants

import "time"

const (
	CacheTTL                 = 30 * time.Second
	WorkerQueueSize          = 200
	NumOfAttemptsToConnectDB = 10
	PauseToWaitDBConnection  = 2 * time.Second
)
