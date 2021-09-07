package network

type OracleError struct {
	ErrCode int
	ErrMsg  string
}

func (err *OracleError) Error() string {
	return err.ErrMsg
}
