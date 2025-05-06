# FORUM

## Overview

This project is a web-based forum application that facilitates user communication through posts and comments. It incorporates features such as user authentication, post categorization, liking/disliking of content, and filtering mechanisms. The application is built using Go, utilizes SQLite for data storage, and is containerized using Docker for ease of deployment.
## Features

 ###   User Authentication

- Registration with email, username, and password.

- Login functionality with session management using cookies.

- Encrypted password storage (Bonus).

- Session expiration handling.

 ###   Posts and Comments

- Creation of posts and comments by registered users.

- Association of posts with one or more categories.

- Visibility of posts and comments to all users.

###    Likes and Dislikes

- Registered users can like or dislike posts and comments.

- Display of like and dislike counts to all users.

###    Filtering Mechanism

- Filter posts by categories.

- Filter posts created by the logged-in user.

- Filter posts liked by the logged-in user.

## Technologies Used
- Backend: Go (Golang)

- Database: SQLite

- Authentication: Sessions and cookies 

- Password Encryption: bcrypt 

- Containerization: Docker
   

## Installation and Setup

- **Clone the Repository and navigate to the project directory**
```bash
$  git clone https://learn.zone01kisumu.ke/git/rotieno/forum.git
$  cd forum
```

- **Build the Docker Image**
```bash
$  docker build -t forum .

```

- **Run the Docker Container**
```bash
$  docker run -p 8080:8080 forum

```
- Alternatively, instead of using docker you can run the program directly after cloning the repo by navigating to the directory then using the following command:
```bash
$ go run .
```
- **Access the Application**

    Open your web browser and navigate to http://localhost:8080


## Usage

- Registration: Users can register by providing a unique email, username, and password. The system will return an error if the email is already taken.

- Login: Registered users can log in using their credentials. Upon successful login, a session is created and managed via cookies.

- Creating Posts: Logged-in users can create posts and associate them with one or more categories.

- Commenting: Logged-in users can comment on posts.

- Liking/Disliking: Logged-in users can like or dislike posts and comments. The counts are visible to all users.

- Filtering: Users can filter posts by categories. Logged-in users can additionally filter by their own posts and liked posts.

## Error Handling

The application handles various errors gracefully, including:

- HTTP status errors (e.g., 404 Not Found, 500 Internal Server Error).

- Database errors (e.g., connection issues, query failures).

- Authentication errors (e.g., invalid credentials, session expiration).



## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Contributors
- [Rabin Otieno](https://learn.zone01kisumu.ke/git/rotieno)

- [Kevin Wasonga](https://learn.zone01kisumu.ke/git/kevwasonga)

- [Franklyne Namayi](https://learn.zone01kisumu.ke/git/fnamayi)

- [Granton Onyango](https://learn.zone01kisumu.ke/git/gonyango)