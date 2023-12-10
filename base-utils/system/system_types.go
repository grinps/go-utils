package system

import "context"

type System interface {
	System() System
}

type RegistrySystem interface {
	GetService(context context.Context, serviceName string, serviceTypeName string, options ...GetOption) (serviceInstance any, err error)
	RegisterService(context context.Context, serviceIdentifier string, newService any, options ...RegistrationOption) (err error)
}

type GetOption func(context context.Context, system System, serviceIdentifier string, serviceTypeIdentifier string, options []GetOption, retrievedService any, previousApplicableValue any, err *error) (applicableValue any)

type GetOptionHandler interface {
	ChangeGetOptions(getOptions ...GetOption) (err error)
	CurrentGetOptions() (getOptions []GetOption)
}

type RegistrationOption func(context context.Context, system System, serviceIdentifier string, newService any, options []RegistrationOption, previouslyRegisteredService any, previousApplicableValue any, err *error) (applicableValue any)

type RegistrationOptionHandler interface {
	ChangeRegistrationOptions(registrationOptions ...RegistrationOption) (err error)
	CurrentRegistrationOptions() (registrationOptions []RegistrationOption)
}
