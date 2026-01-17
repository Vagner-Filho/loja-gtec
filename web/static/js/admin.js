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
