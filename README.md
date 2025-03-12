#Run application

#step 1: Create  an executable 
go build -o forum ./cmd

#step 2:  run with the  followiing command 
go run ./cmd/main.go






```md
# Forum Project

This project is a web-based discussion forum built using **Go** and **SQLite**. It allows users to communicate, create posts, comment, like/dislike posts, and filter content.

---
```

## 🏗 Project Structure

```
forum/
│── cmd/                      # Entry point of the application
│   └── main.go                # Initializes the application
│── internal/
│   ├── auth/                  # Authentication (Login, Signup, Sessions)
│   │   ├── handlers.go         # HTTP handlers for authentication
│   │   ├── service.go          # Business logic for authentication
│   │   ├── model.go            # User model (struct)
│   ├── posts/                 # Posts and comments module
│   │   ├── handlers.go         # HTTP handlers for posts
│   │   ├── service.go          # Business logic for posts/comments
│   │   ├── model.go            # Post and Comment models
│   ├── likes/                 # Likes and dislikes module
│   │   ├── handlers.go         # HTTP handlers for likes/dislikes
│   │   ├── service.go          # Business logic for likes
│   │   ├── model.go            # Like model
│   ├── db/                    # Database connection
│   │   ├── sqlite.go           # SQLite initialization
│   │   ├── migrations/         # SQL migration files
│── web/                       # Static assets
│   ├── templates/              # HTML templates
│   ├── static/                 # CSS/JS files
│── config/                    # Configuration files (env, etc.)
│── Dockerfile                 # Docker configuration
│── go.mod                     # Go module dependencies
│── README.md                  # Documentation
```
```


## 🚀 Features

- ✅ **User Authentication** (Login, Signup, Sessions)
- ✅ **Create, Edit, and Delete Posts & Comments**
- ✅ **Like/Dislike Posts & Comments**
- ✅ **Filter Posts by Category**
- ✅ **SQLite Database for Storage**
- ✅ **Docker Support for Easy Deployment**
- ✅ **Secure Password Storage using bcrypt**
- ✅ **Minimalist UI using HTML, CSS, and Go Templates**

---

## 🛠 Setup Instructions

### 1️⃣ Clone the Repository
```sh
git clone https://github.com/Rabinnnn/forum.git
cd forum
```

### 2️⃣ Install Dependencies
```sh
go mod tidy
```

### 3️⃣ Run the Application
```sh
go run ./cmd
```

### 4️⃣ Open in Browser  
Visit `http://localhost:8080` in your browser.

---

## 🗃 Database Schema
The application uses an SQLite database with the following tables:

- **Users**: Stores user details
- **Posts**: Stores user-generated posts
- **Comments**: Stores comments on posts
- **Likes**: Tracks likes/dislikes for posts and comments

---

## 🐳 Docker Support

### 1️⃣ Build the Docker Image
```sh
docker build -t forum-app .
```

### 2️⃣ Run the Container
```sh
docker run -p 8080:8080 forum-app
```

---

## 📜 License
This project is licensed under the **MIT License**.

---

## 📩 Contact
For any issues or contributions, feel free to submit a pull request or open an issue.
```

---

### 🛠 Key Improvements:
1. **Added Headers** (`## Features`, `## Setup Instructions`)  
2. **Formatted the File Tree** (inside a proper Markdown code block)  
3. **Added Code Blocks** for commands (`git clone`, `go run`)  
4. **Added Emojis** for visual appeal  
5. **Improved Readability** with bullet points  

 🚀 