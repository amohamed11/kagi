package kagi

type Error string

func (e Error) Error() string { return string(e) }

const (
	TRUE               = 1
	FALSE              = 0
	KEY_NOT_FOUND      = Error("key not found in database.")
	KEY_ALREADY_EXISTS = Error("key already exists in database.")
	ERROR_WRITING      = Error("error writing to database")
)

func Check(e error) {
	if e != nil {
		panic(e)
	}
}
