//let ws;

document.addEventListener("DOMContentLoaded", function () {
  const newPostForm = document.getElementById('new-post-form');
  const newCommentForm = document.getElementById("new-comment-form");
  const postsContainer = document.getElementById('posts-container');
  const commentsContainer = document.getElementById("comments-container");
  const chatListContainer = document.getElementById("private-chat-container");
  if (!newPostForm) {
    console.error("Error: newPostForm is not found in the DOM");
    return;
  }

  function connectWebSocket() {
    //ws = new WebSocket("ws://localhost:8080/ws");

    ws.onopen = () => {
      console.log("WebSocket is connected.");
    };

    ws.onclose = (event) => {
      console.log("WebSocket is closed. Reconnect will be attempted in 1 second.", event.reason);
      setTimeout(() => {
        connectWebSocket();
      }, 1000);
    };

    ws.onerror = (err) => {
      console.error("WebSocket encountered error: ", err.message, "Closing socket");
      ws.close();
    };

    ws.addEventListener('message', function (event) {
      const receivedData = JSON.parse(event.data);
      console.log("Type ===:", receivedData.type);
      switch (receivedData.type) {
        case "new_post":
          displayNewPost(receivedData.data);
          break;
        case "post_list":
          displayAllPost(receivedData.data);
          break;
        case "comment_list":
          displayAllComment(receivedData.data);
          break;
        case "new_comment":
          displayNewComment(receivedData.data);
          break;
        case "online_users":
          displayOnlineUsers(receivedData.data);
          break;
        case "chat_list":
          displayPrivateChat(receivedData.data);
          break;
        case "new_message":
          displayMessage(receivedData.data);
          break;
        case "more_chat_list":
          displayMoreMessage(receivedData.data);
          break;
        case "logout":
          refreshOnlineUsers();
          break;
      }
    });
  }

  function refreshOnlineUsers(){
      const payload = {
        user_id: Number(sessionStorage.getItem("loggedInUserId")),
      };
      sendMessage("get_online_users_sort", payload);
  }
  function displayAllPost(postData) {
    // refresh online users
    const payload = {
      user_id: Number(sessionStorage.getItem("loggedInUserId")),
    };
    sendMessage("get_online_users_sort", payload);

    // Create a new DOM element for the post and add it to postsContainer
    const postsTempContainer = document.createElement("div");
    postsTempContainer.classList.add("post-list");
    // Display the post content, user ID, and timestamp
    postData.forEach((item) => {
      const newPostListElement = document.createElement("div");
      newPostListElement.onclick = function () {
        // Add your desired function or code here
        showView("comments-section");
        dispalyCommentsForm(item);
        console.log("Div clicked!" + item.id);
      };
      newPostListElement.innerHTML = `
            <p>${item.content}</p>
            <p>Posted by: User - ${item.nickname || "Unknown"} </p><br/>`;
      postsTempContainer.appendChild(newPostListElement);
    });
    const firstPostElement = postsContainer.firstChild;
    postsContainer.insertBefore(postsTempContainer, firstPostElement);
    //postsContainer.appendChild(newPostElement);
  }

  function displayAllComment(commentData) {
    // Create a new DOM element for the post and add it to postsContainer
    const commentsTempContainer = document.createElement("div");
    commentsTempContainer.classList.add("comment-list");
    // Display the post content, user ID, and timestamp
    if (!commentData) return;
    commentData.forEach((item) => {
      const newcommentListElement = document.createElement("div");
      newcommentListElement.innerHTML = `
            <p>${item.content}</p>
            <p>commented by: User - ${item.nickname || "Unknown"} </p><br/>`;
      commentsTempContainer.appendChild(newcommentListElement);
    });
    commentsContainer.appendChild(commentsTempContainer);
  }

  function displayNewPost(postData) {
    console.log("Received postData:", postData);
    // Ensure postData contains the expected properties
    if (!postData.content || !postData.category || !postData.user_id || !postData.timestamp) {
      console.error("Invalid post data received:", postData);
      return;
    }

    // Create a new DOM element for the post and add it to postsContainer
    const newPostElement = document.createElement('div');
    newPostElement.classList.add('post');

    newPostElement.onclick = function () {
      // Add your desired function or code here
      showView("comments-section");
      dispalyCommentsForm(postData);
      console.log("Div clicked!" + postData.id);
    };
    // Display the post content, user ID, and timestamp
    newPostElement.innerHTML = `
        <p>${postData.content}</p>
        <p>Posted by: User - ${postData.nickname || 'Unknown'} </p>
        `;
    postsContainer.appendChild(newPostElement);
  }

  function displayOnlineUsers(userList) {
    const userListContainer = document.getElementById("online-users");
    while (userListContainer.lastChild) {
      userListContainer.removeChild(userListContainer.lastChild);
    }
    userList.forEach((item) => {
      console.log(item.online);
      const userElement = document.createElement("div");
      userElement.onclick = function () {
        // Add your desired function or code here
        getChatHistory(item);
        // Remove 'selected-user' class from all user elements
        const userElements = document.getElementsByClassName("user");
        for (let i = 0; i < userElements.length; i++) {
          userElements[i].classList.remove("selected-user");
        }
        userElement.classList.add("selected-user");
      };
      if (sessionStorage.getItem("SlectedUserId") && parseInt(sessionStorage.getItem("SlectedUserId")) === item.id){
        userElement.classList.add("selected-user");
      }
        userElement.innerHTML = `
                    <div style="display: flex;">${
                      item.nickname
                    } <div style="width: 15px;height: 15px;background: ${
          item.online ? "green" : "gray"
        };margin-left: 5px;border-radius: 50px;"></div></div><br/>`;
      userElement.classList.add("user"); // Add 'user' class to the user element
      userListContainer.appendChild(userElement);
    });
  }


  function getChatHistory(user) {
    sessionStorage.setItem("SlectedUserId", user.id);
    const payload = {
      user_id: Number(sessionStorage.getItem("loggedInUserId")),
      receiver_id: Number(sessionStorage.getItem("loggedInUserId")),
      sender_id: user.id,
      offset: 0
    };
    sendMessage("get_chat_list", payload);
  }

  function displayPrivateChat(chatList) {

    if (!chatList) {
      while (chatListContainer.lastChild) {
        chatListContainer.removeChild(chatListContainer.lastChild);
      }
      return;
    }

    while (chatListContainer.lastChild) {
      chatListContainer.removeChild(chatListContainer.lastChild);
    }
    const userId = sessionStorage.getItem("loggedInUserId");

    chatList.forEach((item) => {
      //Todo
      const chatElement = document.createElement("div");
      chatElement.classList.add("chat-message");
      if (item.sender_id === parseInt(userId)) {
        chatElement.classList.add("self-chat");
      } else {
        chatElement.classList.add("friend-chat");
      }
      chatElement.innerHTML = `<p>${item.content}</p>
        <div>
        <span class="message-timestamp">${item.nickname}</span>
        <span class="message-timestamp">${item.sent_at}</span>
        </div>
        `;
      chatListContainer.appendChild(chatElement);
    });
  }

  function displayMoreMessage(chatList) {
    if (!chatList) return;

    const tempChatList = chatList.reverse();
    const userId = sessionStorage.getItem("loggedInUserId");

    tempChatList.forEach((item) => {
      //Todo
      const chatElement = document.createElement("div");
      chatElement.classList.add("chat-message");
      if (item.sender_id === parseInt(userId)) {
        chatElement.classList.add("self-chat");
      } else {
        chatElement.classList.add("friend-chat");
      }
      chatElement.innerHTML = `<p>${item.content}</p>
        <div>
        <span class="message-timestamp">${item.nickname}</span>
        <span class="message-timestamp">${item.sent_at}</span>
        </div>
        `;
      chatListContainer.prepend(chatElement);
    });
  }
  function displayNewComment(commentData) {
    // Ensure postData contains the expected properties
    if (!commentData.content || !commentData.user_id) {
      console.error("Invalid post data received:", commentData);
      return;
    }

    // Create a new DOM element for the post and add it to postsContainer
    const newElement = document.createElement("div");
    newElement.classList.add("post");
    // Display the post content, user ID, and timestamp
    newElement.innerHTML = `
        <p>${commentData.content}</p>
        <p>Posted by: User - ${commentData.nickname || "Unknown"} </p>
        `;
    commentsContainer.appendChild(newElement);
  }
  function dispalyCommentsForm(postData) {
    // Create a new DOM element for the post and add it to postsContainer
    const mainPostContainer = document.getElementById("main-post");
    mainPostContainer.innerHTML = `
                    <h3>${postData.content}</h3>
                    <p>Posted by: User - ${postData.nickname || "Unknown"} </p>`;
    const postIdInput = document.getElementById("post_id");
    postIdInput.value = postData.id;
    //get comments request 
    const payload = {
      user_id: Number(sessionStorage.getItem("loggedInUserId")),
      post_id: postData.id,
    };
    sendMessage("get_comment_list", payload);
  }
  //create new post
  newPostForm.addEventListener('submit', function (event) {
    event.preventDefault();

    const formData = new FormData(newPostForm);
    const data = {};
    formData.forEach((value, key) => {
      data[key] = value;
    });


    const payload = {
      user_id: Number(sessionStorage.getItem('loggedInUserId')),
      category: data['post-category'],
      content: data['post-content']
    };

    sendMessage('create_post', payload);
    newPostForm.querySelector('textarea').value = '';
  });
  //create new comment
  newCommentForm.addEventListener("submit", function (event) {
    event.preventDefault();

    const formData = new FormData(newCommentForm);
    const data = {};
    formData.forEach((value, key) => {
      data[key] = value;
    });

    console.log(
      "Sending this comment data to the server:",
      JSON.stringify(data)
    );

    const payload = {
      user_id: Number(sessionStorage.getItem("loggedInUserId")),
      post_id: parseInt(data["post_id"]),
      content: data["comment_content"],
    };

    sendMessage("create_comment", payload);
    newPostForm.querySelector("textarea").value = "";
  });

  chatListContainer.addEventListener("scroll", function () {
    if (chatListContainer.scrollTop === 0) {
      // Perform actions when scrolled to the top
      console.log("Scrolled to the top");
      if (!sessionStorage.getItem('SlectedUserId')) return
      const childCount = chatListContainer.childElementCount;

      const payload = {
        user_id: Number(sessionStorage.getItem("loggedInUserId")),
        receiver_id: Number(sessionStorage.getItem("loggedInUserId")),
        sender_id: Number(sessionStorage.getItem("SlectedUserId")),
        offset: childCount,
      };

      sendMessage("get_chat_list_scroll", payload);
      // Add code to load more chat history or perform any other action
    }
  });

  function sendMessage(type, data) {
    if (ws.readyState === WebSocket.OPEN) {
      const message = {
        type: type,
        data: data
      };
      ws.send(JSON.stringify(message));
      console.log("Sending message:", JSON.stringify(message));
    } else {
      console.error("Failed to send message: WebSocket connection is not open.");
    }
  }

  connectWebSocket();
  // newPostForm.reset();

});


function connectWebSocket() {
  ws = new WebSocket("ws://localhost:8080/ws");

  ws.onopen = () => {
    console.log("WebSocket is connected.");

  };

  ws.onclose = (event) => {
    console.log("WebSocket is closed. Reconnect will be attempted in 1 second.", event.reason);
    setTimeout(() => {
      connectWebSocket();
    }, 1000);
  };

  ws.onerror = (err) => {
    console.error("WebSocket encountered error: ", err.message, "Closing socket");
    ws.close();
  };

  // Handle incoming WebSocket messages
  ws.addEventListener('message', function (event) {
    console.log("Received data from server:", event.data);
    const receivedData = JSON.parse(event.data);

    switch (receivedData.type) {
      case 'new_post':
        displayNewPost(receivedData.data);
        break;
      default:
        console.error("Unknown message type:", receivedData.type);
    }
  });
}

function sendMessage(type, data) {
  if (ws.readyState === WebSocket.OPEN) {
    const message = {
      type: type,
      data: data
    };
    ws.send(JSON.stringify(message));
    console.log("Sending message:", JSON.stringify(message));
  } else {
    console.error("Failed to send message: WebSocket connection is not open.");
    setTimeout(function () {
      sendMessage(type, data);
    }, 100);
  }
}