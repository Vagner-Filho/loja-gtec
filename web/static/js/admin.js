// Admin Panel JavaScript for Product Management

document.addEventListener('DOMContentLoaded', () => {
  loadProducts();
  setupEventListeners();
  setupImagePreviews();
});

// Load all products
async function loadProducts() {
  try {
    const response = await fetch('/api/admin/products');
    if (!response.ok) throw new Error('Failed to load products');
    
    const products = await response.json();
    renderProducts(products);
  } catch (error) {
    console.error('Error loading products:', error);
    showMessage('add-message', 'Failed to load products', 'error');
  }
}

// Render products list
function renderProducts(products) {
  const container = document.getElementById('products-list');
  
  if (!products || products.length === 0) {
    container.innerHTML = '<p class="text-gray-500 text-center py-8">No products found. Add your first product above!</p>';
    return;
  }

  container.innerHTML = products.map(product => `
    <div class="border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow" data-product-id="${product.id}">
      <div class="flex items-start gap-4">
        <img src="${product.image}" alt="${product.name}" class="w-24 h-24 object-contain rounded">
        <div class="flex-1">
          <h3 class="text-lg font-semibold text-gray-800">${product.name}</h3>
          <p class="text-gray-600 mt-1">R$ ${product.price.toFixed(2)}</p>
          <p class="text-sm text-gray-500 mt-1">Category: <span class="font-medium">${formatCategory(product.category)}</span></p>
        </div>
        <div class="flex gap-2">
          <button 
            onclick="editProduct(${product.id})" 
            class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 transition-colors text-sm font-medium"
          >
            Edit
          </button>
          <button 
            onclick="deleteProduct(${product.id})" 
            class="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600 transition-colors text-sm font-medium"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  `).join('');
}

// Format category name
function formatCategory(category) {
  const categories = {
    'bebedouros': 'Bebedouros',
    'purificadores': 'Purificadores',
    'refis': 'Refis',
    'pecas': 'Peças'
  };
  return categories[category] || category;
}

// Setup event listeners
function setupEventListeners() {
  // Add product form
  const addForm = document.getElementById('add-product-form');
  addForm.addEventListener('submit', handleAddProduct);

  // Edit product form
  const editForm = document.getElementById('edit-product-form');
  editForm.addEventListener('submit', handleEditProduct);

  // Modal controls
  document.getElementById('close-modal').addEventListener('click', closeModal);
  document.getElementById('cancel-edit').addEventListener('click', closeModal);
  
  // Close modal on outside click
  document.getElementById('edit-modal').addEventListener('click', (e) => {
    if (e.target.id === 'edit-modal') {
      closeModal();
    }
  });
}

// Setup image preview functionality
function setupImagePreviews() {
  // Add form image preview
  const addImageInput = document.getElementById('image');
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

  // Edit form image preview
  const editImageInput = document.getElementById('edit-image');
  editImageInput.addEventListener('change', (e) => {
    const file = e.target.files[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        document.getElementById('edit-preview-img').src = e.target.result;
        document.getElementById('edit-image-preview').classList.remove('hidden');
      };
      reader.readAsDataURL(file);
    } else {
      document.getElementById('edit-image-preview').classList.add('hidden');
    }
  });
}

// Handle add product
async function handleAddProduct(e) {
  e.preventDefault();
  
  const formData = new FormData(e.target);

  try {
    const response = await fetch('/api/admin/products', {
      method: 'POST',
      body: formData // Send as multipart/form-data
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || 'Failed to add product');
    }

    showMessage('add-message', 'Product added successfully!', 'success');
    e.target.reset();
    document.getElementById('add-image-preview').classList.add('hidden');
    loadProducts();
  } catch (error) {
    console.error('Error adding product:', error);
    showMessage('add-message', `Failed to add product: ${error.message}`, 'error');
  }
}

// Edit product
async function editProduct(id) {
  try {
    const response = await fetch('/api/admin/products');
    if (!response.ok) throw new Error('Failed to load product');
    
    const products = await response.json();
    const product = products.find(p => p.id === id);
    
    if (!product) throw new Error('Product not found');

    // Fill form
    document.getElementById('edit-id').value = product.id;
    document.getElementById('edit-name').value = product.name;
    document.getElementById('edit-price').value = product.price;
    document.getElementById('edit-category').value = product.category;
    document.getElementById('edit-current-image').value = product.image;
    
    // Show current image
    document.getElementById('edit-current-preview').src = product.image;
    
    // Reset file input and hide new preview
    document.getElementById('edit-image').value = '';
    document.getElementById('edit-image-preview').classList.add('hidden');

    // Show modal
    openModal();
  } catch (error) {
    console.error('Error loading product for edit:', error);
    showMessage('add-message', 'Failed to load product', 'error');
  }
}

// Handle edit product
async function handleEditProduct(e) {
  e.preventDefault();
  
  const formData = new FormData(e.target);
  const id = formData.get('id');

  try {
    const response = await fetch(`/api/admin/products/${id}`, {
      method: 'PUT',
      body: formData // Send as multipart/form-data
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || 'Failed to update product');
    }

    showMessage('edit-message', 'Product updated successfully!', 'success');
    setTimeout(() => {
      closeModal();
      loadProducts();
    }, 1000);
  } catch (error) {
    console.error('Error updating product:', error);
    showMessage('edit-message', `Failed to update product: ${error.message}`, 'error');
  }
}

// Delete product
async function deleteProduct(id) {
  if (!confirm('Are you sure you want to delete this product?')) {
    return;
  }

  try {
    const response = await fetch(`/api/admin/products/${id}`, {
      method: 'DELETE'
    });

    if (!response.ok) throw new Error('Failed to delete product');

    showMessage('add-message', 'Product deleted successfully!', 'success');
    loadProducts();
  } catch (error) {
    console.error('Error deleting product:', error);
    showMessage('add-message', 'Failed to delete product', 'error');
  }
}

// Show message
function showMessage(elementId, message, type) {
  const element = document.getElementById(elementId);
  element.className = type === 'success' 
    ? 'bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded'
    : 'bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded';
  element.textContent = message;
  element.classList.remove('hidden');

  setTimeout(() => {
    element.classList.add('hidden');
  }, 3000);
}

// Modal controls
function openModal() {
  const modal = document.getElementById('edit-modal');
  modal.classList.remove('hidden');
  document.body.style.overflow = 'hidden';
}

function closeModal() {
  const modal = document.getElementById('edit-modal');
  modal.classList.add('hidden');
  document.body.style.overflow = 'auto';
  document.getElementById('edit-message').classList.add('hidden');
  document.getElementById('edit-image-preview').classList.add('hidden');
}

// Make functions globally available
window.editProduct = editProduct;
window.deleteProduct = deleteProduct;
