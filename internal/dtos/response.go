package dtos

type ResponseDTO struct {
	UserErrs     []error
	ServerErr    error
	Data         map[string]any
	ResponseCode string
	Status       int
}
