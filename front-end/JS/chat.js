
document.addEventListener("DOMContentLoaded", function () {
  console.log("Document is loaded");
  const chatContainer = document.getElementById('chat-container');
  const newMessageForm = document.getElementById('new-message-form');
  const messageContentInput = document.getElementById('message-content');
  fetchUsers().then(users => {
    // Clear existing users
    const userListContainer = document.getElementById('online-users');
    userListContainer.innerHTML = '';

    users.forEach(user => {
      console.log("Displaying user:", user.nickname);
      displayUser(user);
    });
  }).catch(error => {
    console.error("Error fetching users:", error);
  });
  newMessageForm.addEventListener('submit', function (event) {
    event.preventDefault();
    if (!sessionStorage.getItem("SlectedUserId")) {
      alert("Please select user")
      return
    }
    const storedUserId = sessionStorage.getItem('loggedInUserId');
    const receiverId = Number(sessionStorage.getItem("SlectedUserId"));
    const content = messageContentInput.value;
    if (!content) {
      alert("Please type message");
      return;
    }
    if (!content.trim()) {
      console.error("Message content cannot be empty.");
      return;
    }
    const data = {
      sender_id: Number(storedUserId),
      receiver_id: Number(receiverId),
      content: content.trim(),
    };
    sendMessage("new_message", data);
    messageContentInput.value = '';
  });

  //fetchUsers();
});
async function fetchUsers() {
  const response = await fetch('http://localhost:8080/get-users');
  if (!response.ok) {
    throw new Error("Failed to fetch users");
  }
  return await response.json();
}
function displayUser(user) {
  const userElement = document.createElement("div");
  userElement.classList.add("user-item");
  userElement.innerText = user.nickname; 
  const userListContainer = document.getElementById('online-users');

  console.log("Before appending:", user.nickname, userListContainer.innerHTML); 
  userListContainer.appendChild(userElement);

  console.log("After appending:", user.nickname, userListContainer.innerHTML);  
}


function displayMessage(item) {
  console.log("----------display new message -------------")
  const payload = {
    user_id: Number(sessionStorage.getItem("loggedInUserId")),
  };
  sendMessage("get_online_users_sort", payload);


  if (!sessionStorage.getItem("SlectedUserId")) {
    alert("You received new message from " + item.sender_nickname);
    return;
  }

  const chatListContainer = document.getElementById("private-chat-container");
  const chatElement = document.createElement("div");
  chatElement.classList.add("chat-message");
  const userId = sessionStorage.getItem("loggedInUserId");
  console.log(userId, item.sender_id);
  if (item.sender_id === parseInt(userId)) {
    chatElement.classList.add("self-chat");
  } else {
    chatElement.classList.add("friend-chat");
  }
  chatElement.innerHTML = `<p>${item.content}</p>
        <div>
            <span class="message-timestamp">${item.sender_nickname}</span>
            <span class="message-timestamp">${item.sent_at}</span>
        </div>`;
  chatListContainer.appendChild(chatElement);
}
