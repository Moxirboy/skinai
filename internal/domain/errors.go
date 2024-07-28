package domain

const (
	ErrNotExist                     = Err("Phone Number Not Exist")
	ErrPhoneNumberExist             = Err("Phone Number Exist")
	ErrCouldNotScan                 = Err("Cannot Scan from database")
	ErrCouldNotCreateProgram        = Err("Cannot create program type")
	ErrCouldNotRetrieveFromDataBase = Err("Cannot read from database")
	ErrEmptyField=Err("empty space")
)

type Err string

func (e Err) Error() string {
	return string(e)
}
