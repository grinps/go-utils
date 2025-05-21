package server

type SessionCache interface {
	New() (sessionId string, createErr error)
	Delete(sessionId string) error
	Has(sessionId string) (sessionExists bool, validationErr error)
	Get(sessionId string) (sessionData []byte, getErr error)
}
