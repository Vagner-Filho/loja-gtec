// Admin Panel JavaScript for Product Management with HTMX

document.addEventListener('DOMContentLoaded', () => {
  setupImagePreviews();
  setupHTMXEventListeners();
  setupCompatibilityToggle();
  setupSKUListeners();
});

// Setup image preview functionality
function setupImagePreviews() {
  // Add form image preview
  const addImageInput = document.getElementById('image');
  if (addImageInput) {
    addImageInput.addEventListener('change', (e) => {
      const file = e.target.files[0];
      if (file) {
        const reader = new FileReader();
        reader.onload = (e) => {
          document.getElementById('add-preview-img').src = e.target.result;
          document.getElementById('add-image-preview').classList.remove('hidden');
        };
        reader.readAsDataURL(file);
      } else {
        document.getElementById('add-image-preview').classList.add('hidden');
      }
    });
  }
}

// Setup HTMX event listeners
function setupHTMXEventListeners() {
  // Handle form reset after successful submission
  document.body.addEventListener('htmx:afterRequest', function(evt) {
    if (evt.detail.successful) {
      const target = evt.target;

      // Reset add form if it was the target
      if (target.id === 'add-product-form') {
        target.reset();
        document.getElementById('add-image-preview').classList.add('hidden');
      }
    }
  });

  document.body.addEventListener('htmx:afterSwap', function(evt) {
    if (evt.detail.successful) {
      if (evt.detail.requestConfig.path === "/admin/brands/new") {
        const brandDialog = document.querySelector("dialog#brand-modal");
        if (brandDialog) {
          brandDialog.showModal();
        }
      }
      if (evt.detail.requestConfig.path === "/admin/categories/new" || evt.detail.requestConfig.path.includes("/admin/categories/edit/")) {
        const categoryDialog = document.querySelector("dialog#category-modal");
        if (categoryDialog) {
          categoryDialog.showModal();
        }
      }
    }
  });

  // Handle edit form image preview (for dynamically loaded forms)
  document.body.addEventListener('change', function(evt) {
    if (evt.target.id === 'edit-image') {
      const file = evt.target.files[0];
      if (file) {
        const reader = new FileReader();
        reader.onload = (e) => {
          const previewImg = document.getElementById('edit-preview-img');
          const previewDiv = document.getElementById('edit-image-preview');
          if (previewImg && previewDiv) {
            previewImg.src = e.target.result;
            previewDiv.classList.remove('hidden');
          }
        };
        reader.readAsDataURL(file);
      } else {
        const previewDiv = document.getElementById('edit-image-preview');
        if (previewDiv) {
          previewDiv.classList.add('hidden');
        }
      }
    }
  });
}

// Setup compatibility input toggle based on category selection
function setupCompatibilityToggle() {
  function bindToggle(categorySelectId, compatSelectId, compatWrapperId, partsSelectId, partsWrapperId) {
    const categorySelect = document.getElementById(categorySelectId);
    const compatSelect = document.getElementById(compatSelectId);
    const compatWrapper = document.getElementById(compatWrapperId);
    const partsSelect = document.getElementById(partsSelectId);
    const partsWrapper = document.getElementById(partsWrapperId);

    if (!categorySelect) return;

    function disableSelect(select, wrapper) {
      if (select) {
        select.disabled = true;
        select.classList.add('bg-gray-100', 'cursor-not-allowed');
      }
      if (wrapper) wrapper.classList.add('opacity-50');
    }

    function enableSelect(select, wrapper) {
      if (select) {
        select.disabled = false;
        select.classList.remove('bg-gray-100', 'cursor-not-allowed');
      }
      if (wrapper) wrapper.classList.remove('opacity-50');
    }

    function updateState() {
      const selectedOption = categorySelect.options[categorySelect.selectedIndex];
      const hasSelection = categorySelect.value !== '';
      const allowsCompatibility = selectedOption && selectedOption.getAttribute('data-allows-compatibility') === 'true';

      if (!hasSelection) {
        disableSelect(compatSelect, compatWrapper);
        disableSelect(partsSelect, partsWrapper);
      } else if (allowsCompatibility) {
        enableSelect(compatSelect, compatWrapper);
        disableSelect(partsSelect, partsWrapper);
      } else {
        disableSelect(compatSelect, compatWrapper);
        enableSelect(partsSelect, partsWrapper);
      }
    }

    categorySelect.addEventListener('change', updateState);
    categorySelect.addEventListener('htmx:afterSwap', updateState);
    updateState();
  }

  bindToggle('category_id', 'fits_product_ids', 'compat-wrapper', 'part_product_ids', 'parts-wrapper');
  bindToggle('edit-category', 'edit-fits_product_ids', 'edit-compat-wrapper', 'edit-part_product_ids', 'edit-parts-wrapper');
}

function setupSKUListeners() {
  const catEl = document.querySelector('select[name="category_id"]')
  const brandEl = document.querySelector('select[name="brand_ids"]')
  const priceEl = document.querySelector('input[name="price"]')
  const nameEl = document.querySelector('input[name="name"]')
  const skuInput = document.querySelector('input[name="sku"]')

  async function handleCategorySKU() {
    let sku = '';
    let appendedIdentifiers = 0;
    if (catEl && catEl.selectedIndex > 0) {
      sku = catEl.options[catEl.selectedIndex].text.slice(0, 3).toUpperCase();
      appendedIdentifiers++;
    }
    if (brandEl) {
      const selectedBrands = brandEl.selectedOptions;
      if (selectedBrands && selectedBrands.length > 0) {
        for (const selectedBrand of selectedBrands) {
          sku += '-' + selectedBrand.text.slice(0, 3).toUpperCase();
        }
        appendedIdentifiers++;
      }
    }
    if (nameEl && nameEl.value !== "") {
      sku += '-' + nameEl.value.slice(0, 3).toUpperCase();
      appendedIdentifiers++;
    }
    if (priceEl && priceEl.value) {
      sku += '-' + priceEl.value.slice(0, 3).toUpperCase();
      appendedIdentifiers++;
    }

    if (appendedIdentifiers === 4) {
      sku += '-' + (new Date().getTime() / .7).toString().split('').reverse().join('').slice(1, 5);
    }

    skuInput.value = sku;
  }

  if (skuInput) {
    catEl?.addEventListener('input', handleCategorySKU)
    brandEl?.addEventListener('input', handleCategorySKU)
    nameEl?.addEventListener('input', handleCategorySKU)
    priceEl?.addEventListener('input', handleCategorySKU)
  }
}

// Modal controls for HTMX
function openEditModal() {
  const modal = document.getElementById('edit-modal');
  modal.classList.remove('hidden');
  document.body.style.overflow = 'hidden';
}

function closeEditModal() {
  const modal = document.getElementById('edit-modal');
  modal.classList.add('hidden');
  document.body.style.overflow = 'auto';

  // Clear the modal body
  const modalBody = document.getElementById('edit-modal-body');
  if (modalBody) {
    modalBody.innerHTML = '';
  }
}

function openOrderModal() {
  const modal = document.getElementById('order-modal');
  if (!modal) {
    return;
  }
  modal.classList.remove('hidden');
  document.body.style.overflow = 'hidden';
}

function closeOrderModal() {
  const modal = document.getElementById('order-modal');
  if (!modal) {
    return;
  }
  modal.classList.add('hidden');
  document.body.style.overflow = 'auto';

  const modalBody = document.getElementById('order-modal-body');
  if (modalBody) {
    modalBody.innerHTML = '';
  }
}

// Make functions globally available
window.openEditModal = openEditModal;
window.closeEditModal = closeEditModal;
window.openOrderModal = openOrderModal;
window.closeOrderModal = closeOrderModal;

// Status Modal Functions
function openStatusModal() {
  const modal = document.getElementById('status-modal');
  if (!modal) {
    return;
  }
  modal.classList.remove('hidden');
  document.body.style.overflow = 'hidden';
}

function closeStatusModal() {
  const modal = document.getElementById('status-modal');
  if (!modal) {
    return;
  }
  modal.classList.add('hidden');
  document.body.style.overflow = 'auto';

  const modalBody = document.getElementById('status-modal-body');
  if (modalBody) {
    modalBody.innerHTML = '';
  }
}

function showStatusConfirmModal() {
  const orderId = document.getElementById('order-id').value;
  const statusSelect = document.getElementById('status-select');
  const status = statusSelect.value;
  const statusLabel = statusSelect.options[statusSelect.selectedIndex].text;

  // Load confirmation modal content via HTMX
  const modalBody = document.getElementById('status-modal-body');
  if (modalBody) {
    modalBody.innerHTML = `
      <div class="mb-4">
        <p class="text-lg text-gray-800">
          Mudar o status da ordem para <strong class="font-semibold">${statusLabel}</strong>?
        </p>
        <input type="hidden" id="confirm-order-id" value="${orderId}">
        <input type="hidden" id="confirm-status" value="${status}">
      </div>
      <div class="flex justify-end gap-3">
        <button
          type="button"
          onclick="window.backToStatusSelectModal()"
          class="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
        >
          Voltar
        </button>
        <button
          type="button"
          onclick="window.submitStatusUpdate()"
          class="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
        >
          Confirmar
        </button>
      </div>
    `;
  }
}

function backToStatusSelectModal() {
  const orderId = document.getElementById('confirm-order-id').value;
  const currentStatus = document.getElementById('confirm-status').value;

  // Reload the select modal via HTMX
  const modalBody = document.getElementById('status-modal-body');
  if (modalBody) {
    htmx.ajax('GET', `/api/admin/orders/${orderId}/status-modal`, {
      target: '#status-modal-body',
      swap: 'innerHTML',
      onAfterSwap: function() {
        // Restore the previously selected status
        const statusSelect = document.getElementById('status-select');
        if (statusSelect) {
          statusSelect.value = currentStatus;
        }
      }
    });
  }
}

function submitStatusUpdate() {
  const orderId = document.getElementById('confirm-order-id').value;
  const status = document.getElementById('confirm-status').value;

  // Submit the status update via HTMX
  htmx.ajax('POST', `/api/admin/orders/${orderId}/status`, {
    values: { status: status },
    target: `#order-${orderId}`,
    swap: 'outerHTML'
  }).then(() => {
    closeStatusModal();
  }).catch(() => {
    // Error handling - modal stays open
    console.error('Failed to update order status');
  });
}

// Make status modal functions globally available
window.openStatusModal = openStatusModal;
window.closeStatusModal = closeStatusModal;
window.showStatusConfirmModal = showStatusConfirmModal;
window.backToStatusSelectModal = backToStatusSelectModal;
window.submitStatusUpdate = submitStatusUpdate;
