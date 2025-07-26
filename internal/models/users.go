package models

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"` // voter or admin
}

var DummyUsers = []User{
	{"1", "user1", "pass1", "user"},
	{"2", "user2", "pass2", "user"},
	{"3", "user3", "pass3", "user"},
	{"4", "user4", "pass4", "user"},
	{"5", "user5", "pass5", "user"},
	{"6", "user6", "pass6", "user"},
	{"7", "user7", "pass7", "user"},
	{"8", "user8", "pass8", "user"},
	{"9", "user9", "pass9", "user"},
	{"10", "admin", "admin123", "admin"},
}
