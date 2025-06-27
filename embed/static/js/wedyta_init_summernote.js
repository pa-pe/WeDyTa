function initSummernote(field, modelName) {
    $('#' + field).summernote({
        height: 300,
        callbacks: {
            onImageUpload: async function (files) {

                const recordId = getRecordIdForModel(modelName);
                const check = await isImageUploadAllowed(field, modelName, recordId);

                if (!check.allowed) {
                    // alert(check.message);
                    showBootstrapPopup(check.message, 'danger');
                    return;
                }

                const data = new FormData();
                data.append("file", files[0]);
                data.append("model", modelName);
                data.append("field", field);
                data.append("record_id", recordId);


                fetch('/wedyta/upload/image', {
                    method: 'POST',
                    body: data
                })
                    .then(async response => {
                        const text = await response.text();

                        if (!response.ok) {
                            let message = "Upload failed.";

                            try {
                                const parsed = JSON.parse(text);
                                if (parsed.error) message = parsed.error;
                            } catch (_) {
                                message = text; // если текст не JSON — покажем как есть
                            }

                            throw new Error(message);
                        }

                        return text; // это URL
                    })
                    .then(url => {
                        console.log("Image URL:", url);
                        $('#' + field).summernote('insertImage', url);
                    })
                    .catch(error => {
                        console.error("Upload failed:", error);
                        showBootstrapPopup(error.message, "danger");
                    });

            }
        }
    });
}

function getRecordIdForModel(modelName) {
    const form = $(`form input[name="modelName"][value="${modelName}"]`).closest("form");
    if (!form.length) return null;

    const idInput = form.find('input[name="id"]');
    return idInput.val() || null;
}

async function isImageUploadAllowed(field, modelName, recordId) {
    if (!recordId) {
        return {
            allowed: false,
            message: "It looks like you are currently in the new post creation mode.\n\n" +
                "To enable uploading, please save the current post and continue editing and uploading the image from the edit mode."
        };
    }

    try {
        const response = await fetch('/wedyta/upload/check', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({field, model: modelName, id: recordId})
        });

        if (!response.ok) {
            return {
                allowed: false,
                message: `Server responded with error ${response.status} (${response.statusText})`
            };
        }

        const result = await response.json();

        if (typeof result.allowed !== "boolean") {
            return {
                allowed: false,
                message: "Server response malformed: missing 'allowed' field."
            };
        }

        return {
            allowed: result.allowed,
            message: result.message || (result.allowed ? "" : "Server denied image upload.")
        };
    } catch (err) {
        console.error("Upload check error:", err);
        return {allowed: false, message: "Unable to verify permission with server. Please try again later."};
    }
}

function showBootstrapPopup(message, type = 'danger') {
    let container = document.getElementById('bootstrapPopupContainer');
    if (!container) {
        container = document.createElement('div');
        container.id = 'bootstrapPopupContainer';
        container.className = 'position-fixed bottom-0 end-0 p-3';
        container.style.zIndex = '9999';
        document.body.appendChild(container);
    }

    const toastId = 'toast-' + Date.now();

    container.insertAdjacentHTML('beforeend', `
    <div id="${toastId}" class="toast align-items-center text-bg-${type} border-0 mb-2" role="alert" aria-live="assertive" aria-atomic="true">
      <div class="d-flex">
        <div class="toast-body">${message}</div>
        <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast" aria-label="Close"></button>
      </div>
    </div>
  `);

    const toastEl = document.getElementById(toastId);
    const toast = new bootstrap.Toast(toastEl, {delay: 5000});
    toast.show();

    // Удаление DOM после скрытия
    toastEl.addEventListener('hidden.bs.toast', () => toastEl.remove());
}
