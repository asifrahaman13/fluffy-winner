package domain

type User struct {
	Email    string `json:"email" bson:"email"`
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
}

type UserName struct {
	Username string `json:"username" bson:"username"`
}

type AccessToken struct {
	Token string `json:"token"`
}

type ResponseData struct {
	Response string `json:"response"`
}