package structs

type Users struct {
	Name_user     string
	Lastname_user string
	Status        bool
	ChatID        int64
}

type VKUserInfoResponse struct {
	Response []VKUser `json:"response"`
}

type VKUser struct {
	ID        int    `json:"id"`
	Online    int    `json:"online"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
