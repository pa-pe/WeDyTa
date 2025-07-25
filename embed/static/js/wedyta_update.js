const anim_time = 500;
const iconSuccess = '<i class="bi-check-circle" style="color: green;"></i>';
const iconLoading = '<i class="bi-arrow-repeat" style="color: blue;"></i>';
const iconFail = '<i class="bi-x-circle" style="color: red;"></i>';

function send_update_data(data) {
    let query_result = -1;

    // fetch('/update_model', {
    fetch('/wedyta/update', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                query_result = 1;
//            statusContainer.animShow(iconSuccess);
//alert("update ok");
//                 location.reload();
                window.location.href = window.location.pathname + window.location.search + window.location.hash;
            } else {
                query_result = 0;
//            statusContainer.animShow(iconFail);
                alert('Failed to update: ' + (data.error || 'Unknown error'));
            }
        })
        .catch(error => {
            query_result = 0;
//        statusContainer.animShow(iconFail);
            alert('Error: ' + error);
        });

    let success_update = query_result > 0;
    return success_update;
}

function showQueuedAnim(name, tagCont, doFunc) {
    if ($(tagCont).hasClass("change_animation_progress")) {
        setTimeout(showQueuedAnim, 100, name, tagCont, doFunc);
        return;
    }
    doFunc();
}

function animHideElement(element, animTime, tagCont, callback) {
    $(tagCont).addClass("change_animation_progress");
    $(element).animate({opacity: 0}, animTime, function () {
        element.style.display = "none";
        $(tagCont).removeClass("change_animation_progress");
        if (callback) callback();
    });
}

function animShowElement(element, animTime, tagCont) {
    showQueuedAnim("animShow", tagCont, function () {
        element.style.display = "block";
        $(tagCont).addClass("change_animation_progress");
        $(element).animate({opacity: 1}, animTime, function () {
            $(tagCont).removeClass("change_animation_progress");
        });
    });
}

function animHideStatusContainer(statusContainer, animTime, tagCont) {
    if (statusContainer.style.display !== "none") {
        showQueuedAnim("hide-cont", tagCont, function () {
            $(tagCont).addClass("change_animation_progress");
            $(statusContainer).animate({opacity: 0}, animTime, function () {
                statusContainer.style.display = "none";
                statusContainer.innerHTML = '';
                $(tagCont).removeClass("change_animation_progress");
            });
        });
    }
}

function animShowStatusContainer(statusContainer, iconTag, animTime, tagCont, dropdown) {
    animHideStatusContainer(statusContainer, animTime, tagCont);

    showQueuedAnim("show-success", tagCont, function () {
        $(tagCont).addClass("change_animation_progress");
        statusContainer.style.opacity = "0";
        statusContainer.innerHTML = iconTag;
        statusContainer.style.display = "block";
        $(statusContainer).animate({opacity: 1}, animTime, function () {
            $(tagCont).removeClass("change_animation_progress");

            if (iconTag === iconSuccess) {
                animHideStatusContainer(statusContainer, animTime, tagCont);
                animShowElement(dropdown, animTime, tagCont);
            }
        });
    });
}


$(document).ready(function () {
    // for /wedyta/model/id/update
    if ($('#editForm').length === 1) {
        bindSaveButton();
    }

    var currentTd;

    // Function to create the modal if it doesn't exist
    function createModal(title, contentHtml) {
        // remove parent model if exist
        $('#editModal').remove();

        // adding new one
        $('body').append(`
        <div class="modal fade" id="editModal" tabindex="-1" aria-labelledby="editModalLabel" aria-hidden="true">
            <div class="modal-dialog">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title" id="editModalLabel">${title}</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                    </div>
                    <div class="modal-body">
                        ${contentHtml}
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
                        <button type="button" class="btn btn-primary" id="saveButton">Save</button>
                    </div>
                </div>
            </div>
        </div>
    `);

        // Bind events after creating the modal
        bindModalEvents();
    }

    function buildRecordUpdateForm(modelName, recordId, fieldName, content, isTextarea) {
        return `
        <form id="editForm">
        <input type="hidden" name="modelName" value="${modelName}">
        <input type="hidden" name="id" value="${recordId}">
        ${isTextarea
            ? `<textarea class="form-control" name="${fieldName}" rows="5">${content}</textarea>`
            : `<input class="form-control" type="text" name="${fieldName}" value="${content}">`
        }
        </form>
    `;
    }


    // Function to bind events to the modal
    function bindModalEvents() {
        bindSaveButton();

        // Close the modal on ESC key press
        $(document).on('keydown', function (e) {
            if (e.key === 'Escape') {
                $('#editModal').modal('hide');
            }
        });
    }

    function bindSaveButton() {
        const form = $('#editForm');
        $('#saveButton').on('click', function () {
            // var formData = form.serialize();
            // console.log(formData);
            let formDataJson = serializeFormToJson(form);
            // console.log(formDataJson);
            let success_update = send_update_data(formDataJson);
            if (success_update) {
                let newContent = $('#editTextarea').val();
                currentTd.text(newContent);
            }
            $('#editModal').modal('hide');
        });
    }

    $('.editable-textarea, .editable-input').on('dblclick', function () {
        const currentTd = $(this);
        const modelName = currentTd.closest('table').attr("model");
        const fieldName = currentTd.attr('fieldName');
        let title = $("#header_of_" + fieldName).text();
        const content = currentTd.text();
        const isTextarea = currentTd.hasClass('editable-textarea');

        let recordId = currentTd.closest('table').attr("record_id");
        let isTableMode = false
        if (!recordId) {
            isTableMode = true;
            recordId = currentTd.closest('tr').find('.rec_id').text();
        }

        if (isTableMode) {
            title = "#" + recordId + " " + title;
        }

        const formHtml = buildRecordUpdateForm(modelName, recordId, fieldName, content, isTextarea);
        createModal(title, formHtml);

        $('#editModal').modal('show').on('shown.bs.modal', function () {
            $('#editForm').find(isTextarea ? 'textarea' : 'input[type="text"]').focus();
        });
    });
});

function serializeFormToJson(form) {
    let formDataArray = form.serializeArray(); // Получаем массив объектов {name: 'key', value: 'value'}
    let formDataJson = {};

    $.each(formDataArray, function () {
        formDataJson[this.name] = this.value; // Заполняем объект JSON
    });

    return formDataJson; // Возвращаем JSON-объект
}