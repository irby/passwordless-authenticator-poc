package dto

type LoginMethod int32

const (
	Password LoginMethod = 0
	Passcode LoginMethod = 1
	Webauthn LoginMethod = 2
)

func LoginMethodToValue(method LoginMethod) int {
	switch method {
	case Password:
		return 0
	case Passcode:
		return 1
	case Webauthn:
		return 2
	}
	return -1
}
