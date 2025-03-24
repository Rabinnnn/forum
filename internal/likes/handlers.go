package likes
import (
    "encoding/json"
    "net/http"
    "forum/internal/auth"
    "log"
)

type LikeHandler struct {
    service *LikeService
}

func NewLikeHandler(service *LikeService) *LikeHandler {
    return &LikeHandler{service: service}
}

type LikeRequest struct {
    PostID string `json:"post_id"`
    Like   bool   `json:"like"`
}

// Custom unmarshaler to handle both boolean and numeric values
func (lr *LikeRequest) UnmarshalJSON(data []byte) error {
    // First try to unmarshal as the original structure
    type Alias LikeRequest
    aux := &struct {
        *Alias
    }{
        Alias: (*Alias)(lr),
    }
    
    if err := json.Unmarshal(data, aux); err == nil {
        return nil
    }

    // If that fails, try with numeric like value
    aux2 := &struct {
        PostID string `json:"post_id"`
        Like   int    `json:"like"`
    }{}
    
    if err := json.Unmarshal(data, aux2); err != nil {
        return err
    }

    // Convert numeric to boolean
    lr.PostID = aux2.PostID
    lr.Like = aux2.Like != 0
    return nil
}

func (h *LikeHandler) HandleLike(w http.ResponseWriter, r *http.Request) {
    log.Printf("HandleLike called with method: %s", r.Method)

    if r.Method != http.MethodPost {
        log.Printf("Method not allowed: %s", r.Method)
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    userID, ok := auth.GetUserIDFromSession(r)
    if !ok {
        log.Printf("Unauthorized: No valid session found")
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var req LikeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("Error decoding request body: %v", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    log.Printf("Received like request for PostID: %s, Like: %v", req.PostID, req.Like)

    response, err := h.service.ToggleLike(userID, req.PostID, req.Like)
    if err != nil {
        log.Printf("Error toggling like: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    log.Printf("Successfully toggled like. Response: %+v", response)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
