const anim_time = 500;
const iconSuccess = '<i class="bi-check-circle" style="color: green;"></i>';
const iconLoading = '<i class="bi-arrow-repeat" style="color: blue;"></i>';
const iconFail = '<i class="bi-x-circle" style="color: red;"></i>';

function send_update_data(data){
    let query_result = -1;

    // fetch('/update_model', {
    fetch('/render_table/update/', {
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
            location.reload();
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
    $(element).animate({ opacity: 0 }, animTime, function() {
        element.style.display = "none";
        $(tagCont).removeClass("change_animation_progress");
        if (callback) callback();
    });
}

function animShowElement(element, animTime, tagCont) {
    showQueuedAnim("animShow", tagCont, function() {
        element.style.display = "block";
        $(tagCont).addClass("change_animation_progress");
        $(element).animate({ opacity: 1 }, animTime, function() {
            $(tagCont).removeClass("change_animation_progress");
        });
    });
}

function animHideStatusContainer(statusContainer, animTime, tagCont) {
    if (statusContainer.style.display !== "none") {
        showQueuedAnim("hide-cont", tagCont, function() {
            $(tagCont).addClass("change_animation_progress");
            $(statusContainer).animate({ opacity: 0 }, animTime, function() {
                statusContainer.style.display = "none";
                statusContainer.innerHTML = '';
                $(tagCont).removeClass("change_animation_progress");
            });
        });
    }
}

function animShowStatusContainer(statusContainer, iconTag, animTime, tagCont, dropdown) {
    animHideStatusContainer(statusContainer, animTime, tagCont);

    showQueuedAnim("show-success", tagCont, function() {
        $(tagCont).addClass("change_animation_progress");
        statusContainer.style.opacity = "0";
        statusContainer.innerHTML = iconTag;
        statusContainer.style.display = "block";
        $(statusContainer).animate({ opacity: 1 }, animTime, function() {
            $(tagCont).removeClass("change_animation_progress");

            if (iconTag === iconSuccess) {
                animHideStatusContainer(statusContainer, animTime, tagCont);
                animShowElement(dropdown, animTime, tagCont);
            }
        });
    });
}



$(document).ready(function () {
    var currentTd;

    // Function to create the modal if it doesn't exist
    function createModal() {
        if ($('#editModal').length === 0) {
            $('body').append(`
                <div class="modal fade" id="editModal" tabindex="-1" aria-labelledby="editModalLabel" aria-hidden="true">
                    <div class="modal-dialog">
                        <div class="modal-content">
                            <div class="modal-header">
                                <h5 class="modal-title" id="editModalLabel">Edit Content</h5>
                                <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                            </div>
                            <div class="modal-body">
                                <form id="editForm">
                                    <input type="hidden" name="modelName">
                                    <input type="hidden" name="id">
                                    <textarea class="form-control" id="editTextarea" rows="5"></textarea>
                                </form> 
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
    }

    // Function to bind events to the modal
    function bindModalEvents() {
        // Save button functionality
        var form = $('#editForm');
        $('#saveButton').on('click', function () {
            // var formData = form.serialize();
            // console.log(formData);
            let formDataJson = serializeFormToJson(form);
            // console.log(formDataJson);
            send_update_data(formDataJson);
            let newContent = $('#editTextarea').val();
            currentTd.text(newContent);
            $('#editModal').modal('hide');
        });

        // Close the modal on ESC key press
        $(document).on('keydown', function (e) {
            if (e.key === 'Escape') {
                $('#editModal').modal('hide');
            }
        });
    }

    // Open the modal on click
    $('.editable-textarea').on('dblclick', function () {
        currentTd = $(this);
        let modelName = currentTd.closest('table').attr("model");
        let content = currentTd.text();
        let recordId = currentTd.closest('tr').find('.rec_id').text(); // Get the value from td with class rec_id in the same row
        let fieldName = currentTd.attr('fieldName');
        createModal();
        let form = $('#editForm');
        form.find('input[name="modelName"]').val(modelName);
        form.find('input[name="id"]').val(recordId);
        form.find('textarea').attr("name", fieldName).val(content);
        //$('#editModal').modal('show');
        $('#editModal').modal('show').on('shown.bs.modal', function () {
            form.find('textarea').focus();
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