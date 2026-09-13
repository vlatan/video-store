// ==========================================================================
// Form
// ==========================================================================

const mainForm = document.getElementById("main-form");
const formSubmit = mainForm?.querySelector(".form-button");
const formSpinner = mainForm?.querySelector(".submit-spinner");

formSubmit?.addEventListener("click", () => {
    const formInputs = mainForm?.querySelectorAll(".form-input");
    if (!formInputs) return;

    // Check if all required inputs have values
    let ok = true;
    for (const inputElement of formInputs) {
        if (!(inputElement instanceof HTMLInputElement)) continue;
        if (inputElement.required && inputElement.value.trim() === "") {
            ok = false;
            break;
        }
    }

    if (ok) {
        if (formSpinner) {
            formSpinner.classList.add("show");
        }
    }
});

// ==========================================================================
// Directors functionality
// ==========================================================================

const list = document.getElementById("directors-list");

// Add new row
document.getElementById("add-director-btn").addEventListener("click", () => {
    const row = document.createElement("div");
    row.className = "director-row";
    row.innerHTML = `
        <input class="form-input" name="directors" placeholder="Director's name..." 
            required type="text" maxlength="256">
        <button type="button" class="remove-director-btn">&times;</button>`;
    list.appendChild(row);
});

// Delete row (handles both original and newly added rows)
list.addEventListener("click", (e) => {
    if (
        e.target instanceof Element &&
        e.target.classList.contains("remove-director-btn")
    ) {
        e.target.closest(".director-row")?.remove();
    }
});
