package ports

import "github.com/asifrahaman13/bhagabad_gita/internal/core/domain"

type UserService interface {
	Signup(user domain.User) (string, error)
	Login(domain.User) (domain.AccessToken, error)
	GetLLMResponse(string)(string, error)
}

type UserRepository interface {
	BaseRepository[domain.User]
}
