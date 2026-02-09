package hello

type Usecase interface {
	Hello() string
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (*Service) Hello() string {
	return "Hello from backend"
}
