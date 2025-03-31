package filters

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	userPosts, err := fetchUserPostsForPosts(userID)
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}

	if err := renderCreatedTemplateForPosts(w, userPosts, userID); err != nil {
		log.Printf("Error rendering template: %v", err)
		return
	}
}

func HandleLikedPosts(w http.ResponseWriter, r *http.Request) {
	userID, err := validateUserSession(w, r)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	posts, err := fetchUserPostsForLikes(userID)
	if err != nil {
		log.Printf("Error fetching liked posts: %v", err)
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}

	if err := renderLikedTemplate(w, posts, userID); err != nil {
		log.Printf("Error rendering liked posts template: %v", err)
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
        com.id, com.user_id, com.content, com.created_at,
    	(SELECT COUNT(*) FROM comments WHERE post_id = p.id) AS total_comments  -- Correct count of all comments per post
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
		//var commentUsername sql.NullString
		//var commentProfilePic sql.NullString
		var profilePic sql.NullString
		var postTime time.Time
		var totalCount int

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
			&totalCount,
		//	&commentUsername,
		//	&commentProfilePic,
		)
		if err != nil {
			return nil, err
		}

		post.CommentCount = totalCount
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

		// If post already exists in the map, retrieve it
		existingPost, exists := postMap[post.ID]
		if exists {
			post = existingPost
		}

		// Handle NULL comments
		if commentID.Valid && commentUserID.Valid && commentContent.Valid && commentCreatedAt.Valid {
			comment := posts.Comment{
				ID:        commentID.String,
				UserID:    commentUserID.String,
				Content:   commentContent.String,
				CreatedAt: commentCreatedAt.Time,
				//	Username:   commentUsername.String,
				//	ProfilePic: commentProfilePic,
			}
			post.Comments = append(post.Comments, comment)
		}

		postMap[post.ID] = post
	}

	// Convert map to slice
	var postsList []posts.Post
	for _, post := range postMap {
		postsList = append(postsList, post)
	}

	return postsList, nil
}

func fetchUserPostsForLikes(userID string) ([]posts.Post, error) {
	rows, err := db.Globaldb.Query(`
        SELECT 
            p.id, 
            p.user_id, 
            p.title, 
            p.content, 
            p.image_path, 
            p.created_at,
            COALESCE((SELECT COUNT(*) FROM likes WHERE post_id = p.id AND like = 1), 0) as likes,
            COALESCE((SELECT COUNT(*) FROM likes WHERE post_id = p.id AND like = 0), 0) as dislikes,
            u.username, 
            u.profile_pic,
            GROUP_CONCAT(DISTINCT c.id) AS category_ids, 
            GROUP_CONCAT(DISTINCT c.name) AS category_names,
            (SELECT COUNT(*) FROM comments WHERE post_id = p.id) AS comment_count,
            EXISTS(SELECT 1 FROM likes WHERE post_id = p.id AND user_id = ? AND like = 1) as user_liked,
            EXISTS(SELECT 1 FROM likes WHERE post_id = p.id AND user_id = ? AND like = 0) as user_disliked
        FROM posts p
        JOIN users u ON p.user_id = u.id
        LEFT JOIN post_categories pc ON p.id = pc.post_id
        LEFT JOIN categories c ON pc.category_id = c.id
        JOIN likes l ON p.id = l.post_id
        WHERE l.user_id = ? AND l.like = 1
        GROUP BY p.id
        ORDER BY p.created_at DESC
    `, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userPosts []posts.Post
	for rows.Next() {
		var post posts.Post
		var categoryIDs, categoryNames sql.NullString
		var profilePic sql.NullString
		var postTime time.Time
		var userLiked, userDisliked bool

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
			&post.CommentCount,
			&userLiked,
			&userDisliked,
		)
		if err != nil {
			return nil, err
		}

		post.PostTime = FormatTimeAgo(postTime.Local())
		post.ProfilePic = profilePic
		post.UserLiked = userLiked
		post.UserDisliked = userDisliked

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

		userPosts = append(userPosts, post)
	}

	return userPosts, nil
}

func renderCreatedTemplateForPosts(w http.ResponseWriter, userPosts []posts.Post, userID string) error {
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
		Posts:  userPosts,
		UserID: userID,
	}

	return tmpl.Execute(w, data)
}

func renderLikedTemplate(w http.ResponseWriter, userPosts []posts.Post, userID string) error {
	basePath, _ := os.Getwd()
	templatePath := filepath.Join(basePath, "internal", "web", "templates", "liked.html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return err
	}

	data := struct {
		Posts  []posts.Post
		UserID string
	}{
		Posts:  userPosts,
		UserID: userID,
	}

	return tmpl.Execute(w, data)
}
