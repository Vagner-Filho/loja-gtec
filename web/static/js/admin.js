// Admin Panel JavaScript for Product Management with HTMX

document.addEventListener('DOMContentLoaded', () => {
  setupImagePreviews();
  setupHTMXEventListeners();
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

console.log('hello')
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
