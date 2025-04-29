// postReaction handles liking and disliking of posts
function postReaction(event) {
    event.preventDefault();
    event.stopPropagation();

    const button = event.currentTarget; // get the button that was clicked
    const postID = button.getAttribute("data-post-id");
    const action = button.getAttribute("data-action");
    const like = action === "like" ? 1 : 0; // if the action is 'like' assign 1 else 0
   // const userID = button.getAttribute("data-user-id"); // Add userID from the button

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
        console.log("Content-Type:", contentType); // Print the content type

        // if the content type is not json then redirect to login page
        if (!contentType || !contentType.includes("application/json")) {

           // window.location.href = '/login';
            //throw new Error('Session expired. Please sign in again.');
        }
        // if unauthorized redirect to login
        // if (response.status === 401) {
        //     window.location.href = '/login';
        //     throw new Error('Please sign in to react to posts');
        // }
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
            comment_id: commentID,
            like: like,
        }),
        credentials: 'include'
    })
    .then(response => {
        const contentType = response.headers.get("content-type");
        console.log('content type:', contentType);
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

// Comment editing functionality
function editComment(commentId) {
    const contentDiv = document.getElementById(`comment-content-${commentId}`);
    if (!contentDiv) return;
    
    // Check if we're already editing
    if (contentDiv.querySelector('.edit-comment-form')) {
        return;
    }
    
    const currentContent = contentDiv.textContent.trim();
    
    // Create edit form
    const form = document.createElement('form');
    form.method = 'POST';
    form.action = '/editcomment';
    form.className = 'edit-comment-form';
    
    // Create textarea
    const textarea = document.createElement('textarea');
    textarea.name = 'content';
    textarea.value = currentContent;
    textarea.required = true;
    textarea.className = 'comment-input';
    
    // Create hidden input for comment ID
    const hiddenInput = document.createElement('input');
    hiddenInput.type = 'hidden';
    hiddenInput.name = 'comment_id';
    hiddenInput.value = commentId;
    
    // Create buttons container
    const buttonsDiv = document.createElement('div');
    buttonsDiv.className = 'edit-buttons';
    
    // Create save button
    const saveButton = document.createElement('button');
    saveButton.type = 'submit';
    saveButton.className = 'save-btn';
    saveButton.textContent = 'Save';
    
    // Create cancel button
    const cancelButton = document.createElement('button');
    cancelButton.type = 'button';
    cancelButton.className = 'cancel-btn';
    cancelButton.textContent = 'Cancel';
    cancelButton.onclick = (e) => {
        e.preventDefault();
        contentDiv.innerHTML = currentContent;
    };
    
    // Assemble the form
    buttonsDiv.appendChild(saveButton);
    buttonsDiv.appendChild(cancelButton);
    form.appendChild(textarea);
    form.appendChild(hiddenInput);
    form.appendChild(buttonsDiv);
    
    // Replace content with form
    contentDiv.innerHTML = '';
    contentDiv.appendChild(form);
    
    // Focus the textarea
    textarea.focus();
    textarea.setSelectionRange(textarea.value.length, textarea.value.length);
}

// Comment deletion confirmation
function confirmDeleteComment(event, commentId) {
    event.preventDefault();
    
    if (confirm('Are you sure you want to delete this comment? This action cannot be undone.')) {
        const form = event.target.closest('form');
        if (form) {
            form.submit();
        }
    }
}

// Attach event listeners for post reaction buttons
document.querySelectorAll(".like-btn, .dislike-btn").forEach(button => {
    button.addEventListener("click", postReaction);
});

// Attach event listeners for comment reaction buttons
document.querySelectorAll(".comment-like-btn, .comment-dislike-btn").forEach(button => {
    button.addEventListener("click", commentReaction);
});

// Add this function to handle comment submission
// Function to handle comment submission
async function submitComment(event, postId, currentUser) {
    event.preventDefault();
    console.log('CURRENT_USER_IDDDDD:', currentUser);

    const form = event.target;
    const textarea = form.querySelector('textarea');
    const comment = textarea.value.trim();
    
    if (!comment) return;

    try {
        const response = await fetch('/comments', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                post_id: postId,
                content: comment
            }),
            credentials: 'include'
        });

        if (!response.ok) {
            if (response.status === 401) {
                window.location.href = '/login';
                return;
            }
            throw new Error('Failed to post comment');
        }

        // Clear the textarea
        textarea.value = '';

        console.log("Comment submitted! Now loading comments...");

        // Force refresh comments immediately
        await loadComments(postId, currentUser);
        // Attach event listeners for comment reaction buttons
        document.querySelectorAll(".comment-like-btn, .comment-dislike-btn").forEach(button => {
            button.addEventListener("click", commentReaction);
        });

        // Update comment count
        const commentCountElement = document.getElementById(`comments-${postId}`);
        if (commentCountElement) {
            const currentCount = parseInt(commentCountElement.textContent) || 0;
            commentCountElement.textContent = currentCount + 1;
        }

    } catch (error) {
        console.error('Error posting comment:', error);
        alert('Failed to post comment. Please try again.');
    }
}


// Function to load comments
async function loadComments(postId, currentUser) {
    try {
        console.log(`Fetching comments for post ${postId}...`);

        const response = await fetch(`/comments/post?post_id=${postId}&t=${Date.now()}`, {
            headers: { "Cache-Control": "no-cache" },
        });

        if (!response.ok) throw new Error(`Failed to fetch comments (Status: ${response.status})`);

        const data = await response.json();
        console.log("API Response Data:", data);

        if (!data.comments || !Array.isArray(data.comments)) {
            throw new Error("Invalid response format: Expected an object with a 'comments' array.");
        }

        const comments = data.comments;
        //const currentUserId = data.current_user_id; // The logged-in user's ID
        const comment_list = document.getElementById("comments-list-{{.ID}}");
        const CURRENT_USER_ID = comment_list ? comment_list.getAttribute("data-user-id") : null;


        const commentsList = document.getElementById(`comments-list-${postId}`);
        if (!commentsList) {
            console.error(`Element #comments-list-${postId} not found!`);
            return;
        }

        // Update the comments list
        commentsList.innerHTML = comments.map(comment => {
            const isAuthor = comment.user_id === currentUser; // Ensure correct key matching
            console.log('comment.user_id:', comment.user_id);
            console.log('CURRENT_USER_ID:', currentUser);

            return `
            <div class="comment">
                <div class="comment-header">
                    <span class="comment-author">${comment.username}</span>
                    <span class="comment-time">${new Date(comment.created_at).toLocaleString()}</span>
                </div>
                <p class="comment-content" id="comment-content-${comment.id}">${comment.content}</p>

                ${isAuthor ? `
                <div class="comment-actions">
                    <button onclick="editComment('${comment.id}')" class="edit-btn">
                        <i class="fas fa-edit"></i> Edit
                    </button>
                    <form class="delete-comment-form" method="POST" action="/deletecomment" style="display: inline;">
                        <input type="hidden" name="comment_id" value="${comment.id}">
                        <button type="submit" class="delete-btn">
                            <i class="fas fa-trash"></i> Delete
                        </button>
                    </form>
                </div>
                ` : ''}

                 <div class="comment-reaction-buttons">
                <div class="action-container">
                    <button class="action-btn comment-like-btn" data-comment-id="${comment.id}" data-action="like">
                        <i class="fas fa-heart"></i>
                        <span class="count" id="comment-likes-${comment.id}">${comment.likes}</span>
                    </button>
                </div>
                <div class="action-container">
                    <button class="action-btn comment-dislike-btn" data-comment-id="${comment.id}" data-action="dislike">
                        <i class="fas fa-thumbs-down"></i>
                        <span class="count" id="comment-dislikes-${comment.id}">${comment.dislikes}</span>
                    </button>
                </div>
            </div>
            </div>
            `;
        }).join('');

        console.log(`Loaded ${comments.length} comments.`);
    } catch (error) {
        console.error("Error loading comments:", error);
    }
}



// Add event listeners to toggle comment section and load comments
document.addEventListener("DOMContentLoaded", function () {
    document.querySelectorAll(".comment-btn").forEach(button => {
        button.addEventListener("click", async function () {
            const postID = this.getAttribute("data-post-id");
            const commentSection = document.getElementById(`comments-section-${postID}`);
            
            if (commentSection.style.display === "none" || commentSection.style.display === "") {
                commentSection.style.display = "block";
               // await loadComments(postID); // Load comments only when opening
            } else {
                commentSection.style.display = "none";
            }
        });
    });
});




