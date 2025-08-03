const anim_time = 500;
const iconSuccess = '<i class="bi-check-circle" style="color: green;"></i>';
const iconLoading = '<i class="bi-arrow-repeat" style="color: blue;"></i>';
const iconFail = '<i class="bi-x-circle" style="color: red;"></i>';

async function send_update_data(data, pageRefresh = true) {
    try {
        const response = await fetch('/wedyta/update', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        });

        const result = await response.json();

        if (result.success) {
            if (pageRefresh) {
                window.location.href = window.location.pathname + window.location.search + window.location.hash;
            }
            return true;
        } else {
            alert('Failed to update: ' + (result.error || 'Unknown error'));
            return false;
        }
    } catch (error) {
        alert('Error: ' + error);
        return false;
    }
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

var currentTd;

// Function to create the modal if it doesn't exist
function createModal(title, contentHtml) {
    // remove previous model if it exist
    $('#editModal').remove();

    // Create new modal
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

    // Bind events after creating the modal
    bindModalEvents();
}

function bindSaveButton() {
    const form = $('#editForm');
    $('#saveButton').on('click', async function () {
        // var formData = form.serialize();
        // console.log(formData);
        let formDataJson = serializeFormToJson(form);
        // console.log(formDataJson);

        let success_update = await send_update_data(formDataJson);
        if (success_update) {
            let newContent = form.find('textarea, input[type="text"]').first().val(); // take first textarea or text input
            if (currentTd){
                currentTd.text(newContent);
            }
        }
        $('#editModal').modal('hide');

        currentTd = null;
    });
}

function urlParamsToHiddenInputs() {
    const params = new URLSearchParams(window.location.search);
    let inputs = '';

    for (const [key, value] of params.entries()) {
        inputs += `<input type="hidden" name="${key}" value="${value}">\n`;
    }

    return inputs;
}

function buildRecordUpdateForm(modelName, recordId, fieldName, content, isTextarea) {
    const hiddenInputs = urlParamsToHiddenInputs();

    return `
        <form id="editForm">
        <input type="hidden" name="modelName" value="${modelName}">
        <input type="hidden" name="id" value="${recordId}">
        ${hiddenInputs}
    ${isTextarea
        ? `<textarea class="form-control" name="${fieldName}" rows="5">${content}</textarea>`
        : `<input class="form-control" type="text" name="${fieldName}" value="${content}">`
    }
        </form>
`;
}

function showConfirmModal(title, messageHtml, onConfirm) {
    // Remove previous modal if it exists
    $('#editModal').remove();

    // Create new modal
    $('body').append(`
        <div class="modal fade" id="editModal" tabindex="-1" aria-hidden="true">
            <div class="modal-dialog modal-dialog-centered">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">${title}</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Закрыть"></button>
                    </div>
                    <div class="modal-body">${messageHtml}</div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Отмена</button>
                        <button type="button" class="btn btn-primary" id="confirmModalBtn">Подтвердить</button>
                    </div>
                </div>
            </div>
        </div>
    `);

    const modal = new bootstrap.Modal(document.getElementById('editModal'));
    modal.show();

    $('#confirmModalBtn').on('click', function () {
        modal.hide();
        onConfirm(); // Call the provided confirmation handler
    });
}

let $pendingCheckbox = null;

function handleSwitchChange($checkbox) {
    const checked = $checkbox.prop('checked');
    const row = $checkbox.closest('tr');
    $pendingCheckbox = $checkbox;

    showConfirmModal(
        'Confirm Change',
        checked
            ? 'Are you sure you want to <strong>enable</strong> this record?'
            : 'Are you sure you want to <strong>disable</strong> this record?',
        function () {
            applySwitchChangeConfirmed($pendingCheckbox, row, checked);
        }
    );
}

function applySwitchChangeConfirmed($checkbox, $row, isChecked) {
    const recId = $checkbox.attr('rec_id');
    const modelName = $checkbox.closest('table').attr("model") || 'unknown_model';
    const fieldName = $checkbox.closest('td').attr('fieldname');

    const data = {
        modelName: modelName,
        id: recId,
        [fieldName]: isChecked ? 1 : 0
    };

    send_update_data(data, false).then(success => {
        if (success) {
            if (isChecked) {
                $row.removeClass('disabled');
            } else {
                $row.addClass('disabled');
            }
        } else {
            // rollback checkbox state on failure
            $checkbox.prop('checked', !isChecked);
        }
    });

    $pendingCheckbox = null;
}

function restoreSwitchOnCancel() {
    if ($pendingCheckbox) {
        // Revert checkbox if modal was dismissed
        $pendingCheckbox.prop('checked', !$pendingCheckbox.prop('checked'));
        $pendingCheckbox = null;
    }
}

$(document).ready(function () {
    // for /wedyta/model/id/update
    if ($('#editForm').length === 1) {
        bindSaveButton();
    }

    $('.editable-textarea, .editable-input').on('dblclick', function () {
        currentTd = $(this);
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

    $(document).on('change', '.table-model-record .editable-bs5switch .form-check-input', function () {
        handleSwitchChange($(this));
    });

    $(document).on('change', '.table-model-records .editable-bs5switch .form-check-input', function () {
        handleSwitchChange($(this));
    });

    // On cancel - return the checkbox to its original state
    $(document).on('hidden.bs.modal', '#editModal', restoreSwitchOnCancel);
});

function serializeFormToJson(form) {
    let formDataJson = {};
    // This code skips disabled checkboxes
    // let formDataArray = form.serializeArray(); // Получаем массив объектов {name: 'key', value: 'value'}
    //
    // $.each(formDataArray, function () {
    //     formDataJson[this.name] = this.value; // Заполняем объект JSON
    // });

    form.find('input, select, textarea').each(function () {
        let name = $(this).attr('name');
        if (!name) return;

        if ($(this).attr('type') === 'checkbox') {
            formDataJson[name] = $(this).is(':checked') ? 1 : 0;
        } else if ($(this).attr('type') === 'radio') {
            if ($(this).is(':checked')) {
                formDataJson[name] = $(this).val();
            }
        } else {
            formDataJson[name] = $(this).val();
        }
    });

    // console.log(formDataJson);

    return formDataJson; // Возвращаем JSON-объект
}