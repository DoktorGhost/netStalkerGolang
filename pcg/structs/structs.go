package structs

type Users struct {
	Name_user     string
	Lastname_user string
	Status        bool
	ChatID        int64
	UserID        int
}

type VKUserInfoResponse struct {
	Response []VKUser `json:"response"`
}

type VKUser struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Online    int    `json:"online"`
}
