document.addEventListener("DOMContentLoaded", function () {
    const addForm = document.getElementById("addForm");
    if (addForm) {
        addForm.addEventListener("submit", function (event) {
            event.preventDefault();

            const formData = new FormData(addForm);
            const formObject = {};
            formData.forEach((value, key) => {
                formObject[key] = value;
            });

            fetch("/wedyta/add/", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify(formObject)
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        alert("Record added successfully");
                        // location.reload();
                        window.location.href = window.location.pathname + window.location.search + window.location.hash;
                    } else {
                        alert("Failed to add record: " + (data.error || "Unknown error"));
                    }
                })
                .catch(error => {
                    alert("Error: " + error);
                });
        });
    }
});
