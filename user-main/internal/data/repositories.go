package data

type Repositories struct {
	UserRepo *UserRepo
}

func NewRepositories(dataLayer *Data) *Repositories {
	return &Repositories{
		UserRepo: NewUserRepo(dataLayer),
	}
}
