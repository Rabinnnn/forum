package auth

import "database/sql"

// User represents a forum user
type User struct {
    ID         string
    Username   string
    Email      string
    Password   string
    ProfilePic sql.NullString
}
