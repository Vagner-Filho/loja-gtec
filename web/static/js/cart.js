function addToCart(productName, price) {
  const cart = getCart();
  const productIndex = cart.findIndex(item => item.name === productName);

  if (productIndex > -1) {
    cart[productIndex].quantity += 1;
  } else {
    cart.push({ name: productName, price: price, quantity: 1 });
  }

  saveCart(cart);
  renderCart();
  updateCartBadge();
  showCart();
}

function removeFromCart(productName) {
  let cart = getCart();
  cart = cart.filter(item => item.name !== productName);
  saveCart(cart);
  renderCart();
  updateCartBadge();
}

function updateQuantity(productName, delta) {
  const cart = getCart();
  const productIndex = cart.findIndex(item => item.name === productName);

  if (productIndex > -1) {
    cart[productIndex].quantity += delta;
    if (cart[productIndex].quantity <= 0) {
      cart.splice(productIndex, 1);
    }
    saveCart(cart);
    renderCart();
    updateCartBadge();
  }
}

function clearCart() {
  saveCart([]);
  renderCart();
  updateCartBadge();
}

function getCart() {
  return JSON.parse(localStorage.getItem('cart')) || [];
}

function saveCart(cart) {
  localStorage.setItem('cart', JSON.stringify(cart));
}

function renderCart() {
  const cartItemsContainer = document.getElementById('cart-items');
  const cartTotalContainer = document.getElementById('cart-total');
  const cartCountContainer = document.getElementById('cart-count');
  const cart = getCart();
  
  cartItemsContainer.innerHTML = '';
  let total = 0;
  let totalItems = 0;

  if (cart.length === 0) {
    cartItemsContainer.innerHTML = `
      <div class="text-center py-12">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-24 w-24 mx-auto text-gray-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
        </svg>
        <h3 class="text-lg font-semibold text-gray-700 mb-2">Your cart is empty</h3>
        <p class="text-gray-500">Add some products to get started!</p>
      </div>
    `;
  } else {
    cart.forEach(item => {
      const itemElement = document.createElement('div');
      itemElement.className = 'bg-white border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow duration-200';
      itemElement.innerHTML = `
        <div class="flex justify-between items-start mb-3">
          <div class="flex-1">
            <h4 class="font-semibold text-gray-900 text-lg">${item.name}</h4>
            <p class="text-gray-500 text-sm mt-1">$${item.price.toFixed(2)} cada</p>
          </div>
          <button class="remove-item text-gray-400 hover:text-red-500 transition-colors duration-200" data-name="${item.name}">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="flex justify-between items-center">
          <div class="flex items-center gap-3 bg-gray-100 rounded-lg p-1">
            <button class="decrease-qty bg-white hover:bg-gray-50 text-gray-700 w-8 h-8 rounded-md flex items-center justify-center transition-colors duration-200 shadow-sm" data-name="${item.name}">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" />
              </svg>
            </button>
            <span class="font-semibold text-gray-900 w-8 text-center">${item.quantity}</span>
            <button class="increase-qty bg-white hover:bg-gray-50 text-gray-700 w-8 h-8 rounded-md flex items-center justify-center transition-colors duration-200 shadow-sm" data-name="${item.name}">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
            </button>
          </div>
          <div class="text-right">
            <p class="text-sm text-gray-500">Subtotal</p>
            <p class="text-lg font-bold text-gray-900">$${(item.price * item.quantity).toFixed(2)}</p>
          </div>
        </div>
      `;
      cartItemsContainer.appendChild(itemElement);
      total += item.price * item.quantity;
      totalItems += item.quantity;
    });
  }

  cartTotalContainer.innerText = total.toFixed(2);
  if (cartCountContainer) {
    cartCountContainer.innerText = `${totalItems} ${totalItems === 1 ? 'item' : 'items'}`;
  }
}

function showCart() {
  const cartModal = document.getElementById('cart-modal');
  const cartModalContent = document.getElementById('cart-modal-content');
  
  cartModal.classList.remove('hidden');
  
  // Trigger animation
  setTimeout(() => {
    cartModal.classList.remove('opacity-0');
    cartModalContent.classList.remove('scale-95', 'opacity-0');
    cartModalContent.classList.add('scale-100', 'opacity-100');
  }, 10);
}

function hideCart() {
  const cartModal = document.getElementById('cart-modal');
  const cartModalContent = document.getElementById('cart-modal-content');
  
  // Animate out
  cartModal.classList.add('opacity-0');
  cartModalContent.classList.remove('scale-100', 'opacity-100');
  cartModalContent.classList.add('scale-95', 'opacity-0');
  
  setTimeout(() => {
    cartModal.classList.add('hidden');
  }, 300);
}

function updateCartBadge() {
  const cart = getCart();
  const totalItems = cart.reduce((sum, item) => sum + item.quantity, 0);
  
  let badge = document.getElementById('cart-badge');
  if (!badge) {
    badge = document.createElement('span');
    badge.id = 'cart-badge';
    badge.className = 'absolute -top-2 -right-2 bg-red-500 text-white text-xs rounded-full h-5 w-5 flex items-center justify-center';
    const cartIcon = document.getElementById('cart-icon');
    if (cartIcon) {
      cartIcon.parentElement.style.position = 'relative';
      cartIcon.parentElement.appendChild(badge);
    }
  }
  
  if (totalItems > 0) {
    badge.textContent = totalItems;
    badge.classList.remove('hidden');
  } else {
    badge.classList.add('hidden');
  }
}

document.addEventListener('DOMContentLoaded', () => {
  renderCart();
  updateCartBadge();

  const cartIcon = document.getElementById('cart-icon');
  if (cartIcon) {
    cartIcon.addEventListener('click', showCart);
  }

  const closeCartButton = document.getElementById('close-cart');
  if (closeCartButton) {
    closeCartButton.addEventListener('click', hideCart);
  }

  const clearCartButton = document.getElementById('clear-cart');
  if (clearCartButton) {
    clearCartButton.addEventListener('click', () => {
      if (confirm('Are you sure you want to clear your cart?')) {
        clearCart();
      }
    });
  }

  // Close modal when clicking on backdrop
  const cartModal = document.getElementById('cart-modal');
  if (cartModal) {
    cartModal.addEventListener('click', (e) => {
      if (e.target === cartModal) {
        hideCart();
      }
    });
  }

  // Close modal on Escape key
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && !cartModal.classList.contains('hidden')) {
      hideCart();
    }
  });

  document.querySelectorAll('.add-to-cart').forEach(button => {
    button.addEventListener('click', (e) => {
      const productCard = e.target.closest('.product-card');
      const productName = productCard.dataset.name;
      const productPrice = parseFloat(productCard.dataset.price);
      addToCart(productName, productPrice);
    });
  });

  // Event delegation for dynamically created buttons
  document.getElementById('cart-items').addEventListener('click', (e) => {
    const removeBtn = e.target.closest('.remove-item');
    const increaseBtn = e.target.closest('.increase-qty');
    const decreaseBtn = e.target.closest('.decrease-qty');

    if (removeBtn) {
      const productName = removeBtn.dataset.name;
      removeFromCart(productName);
    } else if (increaseBtn) {
      const productName = increaseBtn.dataset.name;
      updateQuantity(productName, 1);
    } else if (decreaseBtn) {
      const productName = decreaseBtn.dataset.name;
      updateQuantity(productName, -1);
    }
  });
});

export { addToCart, removeFromCart, updateQuantity, clearCart, renderCart, showCart, hideCart, updateCartBadge };
