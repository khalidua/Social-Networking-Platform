package domain

type User struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Bio            string `json:"bio"`
	ProfilePicture string `json:"profile_picture"`
}
