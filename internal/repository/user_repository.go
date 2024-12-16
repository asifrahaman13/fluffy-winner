package repository

import (
	"github.com/asifrahaman13/bhagabad_gita/internal/core/domain"
	"go.mongodb.org/mongo-driver/mongo"
)

var UserRepo *UserRepository

type UserRepository struct {
	*repository[domain.User]
}

func (r *UserRepository) Initialize(db *mongo.Client) *UserRepository {
	UserRepo = &UserRepository{
		repository: &repository[domain.User]{db: db},
	}
	return UserRepo
}
