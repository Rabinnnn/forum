function handleLike(event) {
    event.preventDefault();
    event.stopPropagation();

    const button = event.currentTarget;
    const postId = button.getAttribute('data-post-id');
    const isLike = button.classList.contains('like-btn');

    console.log('Sending like request:', { postId, isLike });

    fetch('/react', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            post_id: postId,
            like: isLike  // This will send true/false instead of 1/0
        }),
        credentials: 'include'
    })
    .then(response => {
        console.log('Response status:', response.status);
        if (!response.ok) {
            if (response.status === 401) {
                console.log('User not authenticated, redirecting to login');
                window.location.href = '/login';
                throw new Error('Please sign in to like posts');
            }
            throw new Error('Failed to update like');
        }
        return response.json();
    })
    .then(data => {
        console.log('Like operation successful:', data);
        document.getElementById(`likes-${postId}`).textContent = data.likes;
        document.getElementById(`dislikes-${postId}`).textContent = data.dislikes;
    })
    .catch(error => {
        console.error('Error in like operation:', error);
        if (!error.message.includes('Please sign in')) {
            alert(error.message);
        }
    });
}

// Add event listeners to like/dislike buttons
document.querySelectorAll('.like-btn, .dislike-btn').forEach(button => {
    button.addEventListener('click', handleLike);
});
