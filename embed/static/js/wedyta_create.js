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

            fetch("/wedyta/create", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify(formObject)
            })
                .then(response => response.json())
                // .then(response => {
                //     if (!response.ok) {
                //         throw new Error("HTTP error " + response.status);
                //     }
                //     return response.json();
                // })
                .then(data => {
                    if (data.success) {
                        alert("Record created successfully");
                        if (data.successfullyCreatedDestination === "refresh_page") {
                            // location.reload();
                            window.location.href = window.location.pathname + window.location.search + window.location.hash;
                        } else if (data.successfullyCreatedDestination.startsWith("/")) {
                            window.location.href = data.successfullyCreatedDestination;
                        }
                    } else {
                        alert("Failed to create record: " + (data.error || "Unknown error"));
                    }
                })
                .catch(error => {
                    alert("Error: " + error);
                });
        });
    }
});
