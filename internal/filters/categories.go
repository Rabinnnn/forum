package filters

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	//"forum/xerrors"
	"forum/internal/db"
	"forum/internal/posts"
	"forum/internal/xerrors"
)

type CategoryHandler struct{}

func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{}
}

func (ch *CategoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/categories":
		if r.Method == http.MethodGet {
			ch.handleGetCategories(w, r)
		} else if r.Method == http.MethodPost {
			ch.handleCreateCategory(w, r)
		} else {
			xerrors.RenderErrorPage(w, http.StatusMethodNotAllowed, xerrors.ErrMethodNotAllowed)
		}
	case "/category":
		if r.Method == http.MethodGet {
			categoryName := r.URL.Query().Get("name")
			if categoryName == "" {
				xerrors.RenderErrorPage(w, http.StatusBadRequest, xerrors.ErrInvalidForm)
				return
			}

			var exists bool
			err := db.Globaldb.QueryRow("SELECT EXISTS(SELECT 1 FROM categories WHERE name = ?)",
				categoryName).Scan(&exists)
			if err != nil {
				log.Printf("Error checking category existence: %v", err)
				xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer)
				return
			}
			if !exists {
				xerrors.RenderErrorPage(w, http.StatusNotFound, "Category not found")
				return
			}

			ch.handleGetPostsByCategoryName(w, r, categoryName)
		} else {
			xerrors.RenderErrorPage(w, http.StatusMethodNotAllowed, xerrors.ErrMethodNotAllowed)
		}
	default:
		xerrors.RenderErrorPage(w, http.StatusNotFound, xerrors.ErrNotFound)
	}
}

func (ch *CategoryHandler) checkAuthStatus(r *http.Request) bool {
	//cookie, err := r.Cookie("session_token")
	_, err := r.Cookie("session_token")

	if err != nil {
		return false
	}
	// _, err = xerrors.ValidateSession(db.Globaldb, cookie.Value)
	// return err == nil
	return true
}

func (ch *CategoryHandler) handleGetCategories(w http.ResponseWriter, _ *http.Request) {
	categories, err := ch.getAllCategories()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer)
		return
	}

	tmpl, err := template.ParseFiles("templates/category_posts.html")
	if err != nil {
		log.Printf("Error parsing categories template: %v", err)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrTemplateExec)
		return
	}

	if err := tmpl.Execute(w, categories); err != nil {
		log.Printf("Error executing categories template: %v", err)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrTemplateExec)
	}
}

func (ch *CategoryHandler) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		xerrors.RenderErrorPage(w, http.StatusBadRequest, xerrors.ErrInvalidForm)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		xerrors.RenderErrorPage(w, http.StatusBadRequest, xerrors.ErrInvalidForm)
		return
	}

	stmt, err := db.Globaldb.Prepare("INSERT INTO categories (name) VALUES (?)")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(name)
	if err != nil {
		log.Printf("Error executing insert: %v", err)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer)
		return
	}

	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}

func (ch *CategoryHandler) handleGetPostsByCategoryName(w http.ResponseWriter, r *http.Request, categoryName string) {
	posts, err := ch.getPostsByCategoryName(categoryName)
	if err != nil {
		log.Printf("Error fetching posts for category %s: %v", categoryName, err)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer)
		return
	}

	isLoggedIn := ch.checkAuthStatus(r)
	var currentUserID string

	// if cookie, err := r.Cookie("session_token"); err == nil {
	// 	if userID, err := xerrors.ValidateSession(db.Globaldb, cookie.Value); err == nil {
	// 		currentUserID = userID
	// 	}
	// }

	data := struct {
		IsLoggedIn    bool
		Posts         []posts.Post
		CurrentUserID string
	}{
		IsLoggedIn:    isLoggedIn,
		Posts:         posts,
		CurrentUserID: currentUserID,
	}

	tmpl, err := template.ParseFiles("templates/category_posts.html")
	if err != nil {
		log.Printf("Error parsing category posts template: %v", err)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrTemplateExec)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Error executing category posts template: %v", err)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrTemplateExec)
	}
}

func (ch *CategoryHandler) getPostsByCategoryName(categoryName string) ([]posts.Post, error) {
	rows, err := db.Globaldb.Query(`
        SELECT p.id, p.title, p.content, p.imagepath, p.post_at, u.username, u.profile_pic,
               (SELECT COUNT(*) FROM reaction WHERE post_id = p.id AND like = 1) AS Likes,
               (SELECT COUNT(*) FROM reaction WHERE post_id = p.id AND like = 0) AS Dislikes,
               (SELECT COUNT(*) FROM comments WHERE post_id = p.id) AS Comments
        FROM posts p
        JOIN post_categories pc ON p.id = pc.post_id
        JOIN users u ON p.user_id = u.id
        JOIN categories c ON pc.category_id = c.id
        WHERE c.name = ?
    `, categoryName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	postMap := make(map[int]posts.Post)
	for rows.Next() {
		var post posts.Post
		var postTime time.Time
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.ImagePath, &postTime, &post.Username, &post.ProfilePic, &post.Likes, &post.Dislikes, &post.Comments); err != nil {
			log.Printf("Error scanning post: %v", err)
			continue
		}
		post.PostTime = FormatTimeAgo(postTime.Local())
		
		idNum, _ := strconv.Atoi(post.ID)
		postMap[idNum] = post
	}

	var posts []posts.Post
	for _, post := range postMap {
		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func (ch *CategoryHandler) getAllCategories() ([]posts.Category, error) {
	rows, err := db.Globaldb.Query("SELECT id, name FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []posts.Category
	for rows.Next() {
		var category posts.Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			log.Printf("Error scanning category: %v", err)
			continue
		}
		categories = append(categories, category)
	}

	return categories, rows.Err()
}



func FormatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 48*time.Hour:
		return "yesterday"
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	default:
		return t.Format("Jan 2, 2006")
	}
}