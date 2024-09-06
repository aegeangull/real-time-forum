document.addEventListener("DOMContentLoaded", function () {
  const loginForm = document.getElementById('login-form');
  let loggedInUserId = null;

  loginForm.addEventListener('submit', function (event) {
    event.preventDefault();

    const formData = new FormData(loginForm);
    const data = {};
    formData.forEach((value, key) => {
      data[key] = value;
    });
    console.log('Attempting login with data:', data);

    fetch('http://localhost:8080/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    })
      .then(response => {
        if (!response.ok) {
          throw new Error("Failed to login");
        }
        return response.json();
      })
      .then(data => {
        console.log('Token received:', data.token);
        // Store the token in session storage
        sessionStorage.setItem('token', data.token);

        // Store the user ID 
        loggedInUserId = data.user_id;
        
        // Store the userId in session storage to use it after in other views
        sessionStorage.setItem('loggedInUserId', loggedInUserId.toString());

        // Now you can use this token to make requests to protected routes
        fetch('http://localhost:8080/protected', {
          headers: {
            'Authorization': data.token
          }
        })
          .then(response => response.text())
          .then(content => {
            console.log('Protected content:', content);
            console.log("Login was successful!");
            userIsLoggedIn = true;
            updateUIBasedOnLoginStatus();
          });
      })
      .catch(error => console.error('Error:', error));
  });

});
