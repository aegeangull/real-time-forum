var ws = new WebSocket('ws://localhost:8080/ws');

ws.addEventListener('open', function (event) {
    console.log('WebSocket connection opened:', event);
});

ws.addEventListener('message', function (event) {
    console.log('WebSocket message received:', event.data);
    const message = JSON.parse(event.data);

    if (message.type === "online_users") {
        handleOnlineUsers(message.data);
    }
});

function handleOnlineUsers(users) {
    // Clear current users from the list
    const userListContainer = document.getElementById('online-users');
    userListContainer.innerHTML = '';

    // Display each user
    users.forEach(user => {
        displayUser(user);
    });
}

// This will execute once the DOM is fully loaded
document.addEventListener("DOMContentLoaded", function () {
    console.log("Document loaded.");

    const newMessageForm = document.getElementById('new-message-form');

    // Check if newMessageForm is found in the DOM
    if (!newMessageForm) {
        console.error("Error: newMessageForm is not found in the DOM");
        return;
    }
});
