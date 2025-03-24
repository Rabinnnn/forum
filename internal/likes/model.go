package likes

import "time"

type Like struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    PostID    string    `json:"post_id"`
    IsLike    bool      `json:"is_like"`
    CreatedAt time.Time `json:"created_at"`
}

type LikeResponse struct {
    Likes    int    `json:"likes"`
    Dislikes int    `json:"dislikes"`
    Error    string `json:"error,omitempty"`
}
