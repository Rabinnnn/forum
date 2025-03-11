// postReaction handles liking and disliking of posts
function postReaction(event) {
    event.preventDefault();
    event.stopPropagation();

    const button = event.currentTarget; // get the button that was clicked
    const postID = button.getAttribute("data-post-id");
    const action = button.getAttribute("data-action");
    const like = action === "like" ? 1 : 0; // if the action is 'like' assign 1 else 0

    fetch("/react", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            post_id: parseInt(postID),
            like: like,
        }),
        credentials: 'include'
    })
    .then(response => {
        const contentType = response.headers.get("content-type");
        // if the content type is not json then redirect to login page
        if (!contentType || !contentType.includes("application/json")) {
            window.location.href = '/login';
            throw new Error('Session expired. Please sign in again.');
        }
        // if unauthorized redirect to login
        if (response.status === 401) {
            window.location.href = '/login';
            throw new Error('Please sign in to react to posts');
        }
        return response.json();
    })
    .then(data => {
        if (data.error) {
            throw new Error(data.error);
        }
        //update the counters for likes and dislikes
        const likesElement = document.getElementById(`likes-${postID}`);
        const dislikesElement = document.getElementById(`dislikes-${postID}`);
        if (likesElement && dislikesElement) {
            likesElement.textContent = data.likes;
            dislikesElement.textContent = data.dislikes;
        }
    })
    .catch(error => {
        console.error("Error:", error);
        if (!error.message.includes("Please sign in")) {
            alert(error.message);
        }
    });
}

//commentReaction handles liking and disliking of a comment
function commentReaction(event) {
    event.preventDefault();
    event.stopPropagation();

    const button = event.currentTarget;
    const commentID = button.getAttribute("data-comment-id");
    const action = button.getAttribute("data-action");
    const like = action === "like" ? 1 : 0;

    fetch("/commentreact", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            comment_id: parseInt(commentID),
            like: like,
        }),
        credentials: 'include'
    })
    .then(response => {
        const contentType = response.headers.get("content-type");
        if (!contentType || !contentType.includes("application/json")) {
            window.location.href = '/login';
            throw new Error('Session expired. Please sign in again.');
        }

        if (response.status === 401) {
            window.location.href = '/login';
            throw new Error('Please sign in to react to comments');
        }
        return response.json();
    })
    .then(data => {
        if (data.error) {
            throw new Error(data.error);
        }
        const likesElement = document.getElementById(`comment-likes-${commentID}`);
        const dislikesElement = document.getElementById(`comment-dislikes-${commentID}`);
        
        if (likesElement && dislikesElement) {
            likesElement.textContent = data.likes;
            dislikesElement.textContent = data.dislikes;
        }
    })
    .catch(error => {
        console.error("Error:", error);
        if (!error.message.includes("Please sign in")) {
            alert(error.message);
        }
    });
}


// Attach event listeners for post reaction buttons
document.querySelectorAll(".like-btn, .dislike-btn").forEach(button => {
    button.addEventListener("click", postReaction);
});

// Attach event listeners for comment reaction buttons
document.querySelectorAll(".comment-like-btn, .comment-dislike-btn").forEach(button => {
    button.addEventListener("click", commentReaction);
});
