package filters

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"strings"

	"forum/internal/auth"
	"forum/internal/db"
	"forum/internal/posts"
)

func CreatedPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Check session
	userID, err := validateUserSession(w, r)
	if err != nil {
		return
	}

	// Fetch posts
	posts, err := fetchUserPostsForPosts(userID)
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}

	if err := renderCreatedTemplateForPosts(w, posts, userID); err != nil {
		log.Printf("Error rendering template: %v", err)
		return
	}
}

func LikedPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Check session
	userID, err := validateUserSession(w, r)
	if err != nil {
		return
	}

	// Fetch posts
	posts, err := fetchUserPostsForLikes(userID)
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}

	if err := renderCreatedTemplateForLikes(w, posts, userID); err != nil {
		log.Printf("Error rendering template: %v", err)
		return
	}
}

func validateUserSession(w http.ResponseWriter, r *http.Request) (string, error) {
	// cookie, err := r.Cookie("session_token")
	// if err != nil {
	// 	http.Redirect(w, r, "/signin", http.StatusSeeOther)
	// 	return "", err
	// }
	userID, _ := auth.GetUserIDFromSession(r)

	//userID, err := utils.ValidateSession(db.Globaldb, cookie.Value)
	// if err != nil {
	// 	http.Redirect(w, r, "/signin", http.StatusSeeOther)
	// 	return "", err
	// }

	return userID, nil
}

func fetchUserPostsForPosts(userID string) ([]posts.Post, error) {
	rows, err := db.Globaldb.Query(`
        SELECT p.id, p.user_id, p.title, p.content, p.image_path, p.created_at, p.likes, p.dislikes,
        u.username, u.profile_pic,
        GROUP_CONCAT(c.id) AS category_ids, GROUP_CONCAT(c.name) AS category_names,
        com.id, com.user_id, com.content, com.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		LEFT JOIN post_categories pc ON p.id = pc.post_id
		LEFT JOIN categories c ON pc.category_id = c.id
		LEFT JOIN comments com ON p.id = com.post_id
		WHERE p.user_id = ?
		GROUP BY p.id, com.id
        ORDER BY p.created_at DESC
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	postMap := make(map[string]posts.Post)

	for rows.Next() {
		var post posts.Post
		var categoryIDs sql.NullString
		var categoryNames sql.NullString
		var commentID sql.NullString
		var commentUserID sql.NullString
		var commentContent sql.NullString
		var commentCreatedAt sql.NullTime
		var profilePic sql.NullString
		var postTime time.Time

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.ImagePath,
			&postTime,
			&post.Likes,
			&post.Dislikes,
			&post.Username,
			&profilePic,
			&categoryIDs,
			&categoryNames,
			&commentID,
			&commentUserID,
			&commentContent,
			&commentCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Convert profile picture safely
		post.ProfilePic = profilePic

		// Format time
		post.PostTime = FormatTimeAgo(postTime.Local())

		// Handle NULL categories
		if categoryIDs.Valid && categoryNames.Valid {
			idList := strings.Split(categoryIDs.String, ",")
			nameList := strings.Split(categoryNames.String, ",")
			for i := range idList {
				id, err := strconv.Atoi(strings.TrimSpace(idList[i]))
				if err == nil && i < len(nameList) {
					post.Categories = append(post.Categories, posts.Category{
						ID:   id,
						Name: strings.TrimSpace(nameList[i]),
					})
				}
			}
		}

		// If post already exists in the map, append the new comment
		existingPost, exists := postMap[post.ID]
		if exists {
			post = existingPost
		}

		// Handle NULL comments
		// if commentID.Valid && commentUserID.Valid && commentContent.Valid && commentCreatedAt.Valid {
		// 	comment := posts.CommentData{
		// 		ID:        commentID.String,
		// 		UserID:    commentUserID.String,
		// 		Content:   commentContent.String,
		// 		CreatedAt: commentCreatedAt.Time,
		// 	}
		// 	post.Comments = append(post.Comments, comment)
		// }

		postMap[post.ID] = post
	}

	// Convert map to slice
	var posts []posts.Post
	for _, post := range postMap {
		posts = append(posts, post)
	}

	return posts, nil
}


func fetchUserPostsForLikes(userID string) ([]posts.Post, error) {
	rows, err := db.Globaldb.Query(`
        SELECT p.id, p.user_id, p.title, p.content, p.imagepath, p.post_at, p.likes, p.dislikes, p.comments,
               u.username, u.profile_pic, c.id AS category_id, c.name AS category_name
        FROM posts p
        JOIN users u ON p.user_id = u.id
        LEFT JOIN post_categories pc ON p.id = pc.post_id
        LEFT JOIN categories c ON pc.category_id = c.id
        JOIN reaction r ON p.id = r.post_id
        WHERE r.user_id = ? AND r.like = 1
        ORDER BY p.post_at DESC
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	postMap := make(map[int]posts.Post)
	var postTime time.Time
	for rows.Next() {
		var post posts.Post
		var categoryID sql.NullInt64
		var categoryName sql.NullString
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.ImagePath,
			&postTime,
			&post.Likes,
			&post.Dislikes,
			&post.Comments,
			&post.Username,
			&post.ProfilePic,
			&categoryID,
			&categoryName,
		)
		if err != nil {
			return nil, err
		}
		post.PostTime = FormatTimeAgo(postTime.Local())
		idNum, _ := strconv.Atoi(post.ID)
		postMap[idNum] = post

	}

	var posts []posts.Post
	for _, post := range postMap {
		posts = append(posts, post)
	}

	return posts, nil
}

func renderCreatedTemplateForPosts(w http.ResponseWriter, postss []posts.Post, userID string) error {
	basePath, _ := os.Getwd() // Gets the root directory where the app runs
	templatePath := filepath.Join(basePath, "internal", "web", "templates", "created.html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return err
	}

	data := struct {
		Posts  []posts.Post
		UserID string
	}{
		Posts:  postss,
		UserID: userID,
	}

	return tmpl.Execute(w, data)
}

func renderCreatedTemplateForLikes(w http.ResponseWriter, postss []posts.Post, userID string) error {
	tmpl, err := template.ParseFiles("templates/liked.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return err
	}

	data := struct {
		Posts  []posts.Post
		UserID string
	}{
		Posts:  postss,
		UserID: userID,
	}

	return tmpl.Execute(w, data)
}
