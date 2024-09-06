let userIsLoggedIn = sessionStorage.getItem('loggedInUserId') ? true : false;  // Variable to track login status

// Function to update the UI based on login status
function updateUIBasedOnLoginStatus() {
  const chatButton = document.getElementById('chat-button');
  const postsButton = document.getElementById('posts-button');
  const logoutButton = document.getElementById('logout-button');
  const loginButton = document.getElementById('login-button');
  const registrationButton = document.getElementById('registration-button');
  sessionStorage.removeItem("SlectedUserId");
  if (userIsLoggedIn) {
    // Hide login and registration sections
    document.getElementById('login-section').style.display = 'none';
    document.getElementById('registration-section').style.display = 'none';

    // Hide login and registration buttons
    loginButton.style.display = 'none';
    registrationButton.style.display = 'none';

    // Hide posts and chat sections
    document.getElementById('posts-section').style.display = 'none';
    document.getElementById('chat-section').style.display = 'none';

    // Show chat, posts, and logout buttons
    chatButton.style.display = 'block';
    postsButton.style.display = 'block';
    logoutButton.style.display = 'block';

    // Redirect to posts by default
    showView('posts-section');
  } else {
    // Hide posts and chat sections
    document.getElementById('posts-section').style.display = 'none';
    document.getElementById('chat-section').style.display = 'none';

    // Hide chat and posts buttons
    chatButton.style.display = 'none';
    postsButton.style.display = 'none';

    // Show login and registration sections
    document.getElementById('login-section').style.display = 'block';
    document.getElementById('registration-section').style.display = 'block';

    // Show login and registration buttons
    loginButton.style.display = 'block';
    registrationButton.style.display = 'block';

    // Hide logout button
    logoutButton.style.display = 'none';

  }
}

function showView(viewId) {
  console.log("showView function called with ID: " + viewId);
  // Hide all views
  const allViews = document.querySelectorAll('.view');
  allViews.forEach(view => view.style.display = 'none');

  // Show the selected view
  const selectedView = document.getElementById(viewId);
  selectedView.style.display = 'block';
  if (viewId === "posts-section") {
    const payload = {
      user_id: Number(sessionStorage.getItem("loggedInUserId")),
    };
    sendMessage("get_post_list", payload);
  }
  console.log("chat-section", viewId);
  if (viewId === "chat-section") {
    const payload = {
      user_id: Number(sessionStorage.getItem("loggedInUserId")),
    };
    sendMessage("get_online_users_sort", payload);
  } else {
    sessionStorage.removeItem("SlectedUserId");
  }
}

// Function - logout
function logout() {
  const payload = {
    user_id: Number(sessionStorage.getItem("loggedInUserId")),
  };
  sendMessage("logout", payload);

  userIsLoggedIn = false;
  sessionStorage.removeItem('loggedInUserId');
  sessionStorage.removeItem('token');
  sessionStorage.clear();
  updateUIBasedOnLoginStatus();
}

// Show the login view by default when the page loads
window.addEventListener('DOMContentLoaded', () => {
  console.log("DOM fully loaded and parsed");
  updateUIBasedOnLoginStatus();  // Update the UI based on login status
});
