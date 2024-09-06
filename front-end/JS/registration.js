document.addEventListener("DOMContentLoaded", function() {
    const registrationForm = document.getElementById('registration-form');

    registrationForm.addEventListener('submit', function(event) {
        event.preventDefault();

        const formData = new FormData(registrationForm);
        const data = {};
        formData.forEach((value, key) => {
            data[key] = value;
        });

        fetch('http://localhost:8080/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        })
            .then(response => {
                if (!response.ok) {
                    throw new Error("Failed to register");
                }
                return response.json();
            })
            .then(data => {
                console.log('Registration successful:', data);
                // redirect to login page
            })
            .catch(error => console.error('Error:', error));
    });
});
