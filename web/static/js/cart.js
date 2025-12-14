function addToCart(productName, price, id) {
  const cart = getCart();
  const productIndex = cart.findIndex(item => item.name === productName);

  if (productIndex > -1) {
    cart[productIndex].quantity += 1;
  } else {
    cart.push({ id: Number(id), name: productName, price: price, quantity: 1 });
  }

  saveCart(cart);
  updateCartBadge();

  const dialogExists = document.querySelector('#cart-container #cart-modal');
  if (dialogExists) {
    showCart();
  } else {
    // Load modal first, then render and show
    loadCartModal().then(() => {
      showCart();
    });
  }
}

function removeFromCart(productName) {
  let cart = getCart();
  cart = cart.filter(item => item.name !== productName);
  saveCart(cart);
  updateCartBadge();

  const dialogExists = document.querySelector('#cart-container #cart-modal');
  if (dialogExists) {
    renderCart();
  }
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
    updateCartBadge();

    const dialogExists = document.querySelector('#cart-container #cart-modal');
    if (dialogExists) {
      renderCart();
    }
  }
}

function clearCart() {
  saveCart([]);
  updateCartBadge();

  const dialogExists = document.querySelector('#cart-container #cart-modal');
  if (dialogExists) {
    renderCart();
  }
}

function getCart() {
  return JSON.parse(localStorage.getItem('cart')) || [];
}

function saveCart(cart) {
  localStorage.setItem('cart', JSON.stringify(cart));
}

const INSTALLATION_SERVICE_NAME = 'Serviço de Instalação';

function renderCart() {
  const cartItemsContainer = document.getElementById('cart-items');
  const cartTotalContainer = document.getElementById('cart-total');
  const cartCountContainer = document.getElementById('cart-count');
  const proceedToCheckoutButton = document.getElementById('proceed-to-checkout');

  // Early return if cart elements don't exist yet
  if (!cartItemsContainer || !cartTotalContainer) {
    return;
  }

  const cart = getCart();
  // Filter out installation service for display and calculations
  const displayCart = cart.filter(item => item.name !== INSTALLATION_SERVICE_NAME);

  cartItemsContainer.innerHTML = '';
  let total = 0;
  let totalItems = 0;

  if (displayCart.length === 0) {
    cartItemsContainer.innerHTML = `
      <div class="text-center py-12">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-24 w-24 mx-auto text-gray-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
        </svg>
        <h3 class="text-lg font-semibold text-gray-700 mb-2">Seu carrinho está vazio</h3>
        <p class="text-gray-500">Adicione alguns produtos para começar!</p>
      </div>
    `;

    // Disable checkout button when cart is empty
    if (proceedToCheckoutButton) {
      proceedToCheckoutButton.disabled = true;
      proceedToCheckoutButton.classList.add('opacity-50', 'cursor-not-allowed', 'hover:scale-100');
      proceedToCheckoutButton.classList.remove('hover:scale-105', 'cursor-pointer');
    }
  } else {
    displayCart.forEach(item => {
      const itemElement = document.createElement('div');
      itemElement.className = 'bg-white border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow duration-200';
      itemElement.innerHTML = `
        <div class="flex justify-between items-start mb-3">
          <div class="flex-1">
            <h4 class="font-semibold text-gray-900 text-lg">${item.name}</h4>
            <p class="text-gray-500 text-sm mt-1">R$ ${item.price.toFixed(2)} cada</p>
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
            <p class="text-lg font-bold text-gray-900">R$ ${(item.price * item.quantity).toFixed(2)}</p>
          </div>
        </div>
      `;
      cartItemsContainer.appendChild(itemElement);
      total += item.price * item.quantity;
      totalItems += item.quantity;
    });

    // Enable checkout button when cart has items
    if (proceedToCheckoutButton) {
      proceedToCheckoutButton.disabled = false;
      proceedToCheckoutButton.classList.remove('opacity-50', 'cursor-not-allowed', 'hover:scale-100');
      proceedToCheckoutButton.classList.add('hover:scale-105', 'cursor-pointer');
    }
  }

  cartTotalContainer.innerText = total.toFixed(2);
  if (cartCountContainer) {
    cartCountContainer.innerText = `${totalItems} ${totalItems === 1 ? 'item' : 'itens'}`;
  }
}

function showCart() {
  loadCartModal().then(() => {
    const dialog = document.querySelector('#cart-container #cart-modal');
    if (!dialog) {
      return;
    }

    if (dialog.open) {
      dialog.close();
    }

    dialog.showModal();
    renderCart();

    setTimeout(() => {
      const content = dialog.querySelector('#cart-modal-content');
      if (content) {
        content.classList.remove('scale-95', 'opacity-0');
        content.classList.add('scale-100', 'opacity-100');
      }
    }, 10);

    cartModalUISetup();
  });
}

function hideCart() {
  const dialog = document.querySelector('#cart-container #cart-modal');
  if (dialog) {
    // Animate out
    const content = dialog.querySelector('#cart-modal-content');
    if (content) {
      content.classList.remove('scale-100', 'opacity-100');
      content.classList.add('scale-95', 'opacity-0');
    }

    setTimeout(() => {
      dialog.close();
    }, 300);
  }
}

function loadInstallationServiceModal() {
  const container = document.getElementById('cart-container');
  if (!container) {
    console.error('Cart container not found');
    return Promise.resolve(false);
  }

  const existingModal = container.querySelector('#installation-modal');
  if (existingModal) {
    return Promise.resolve(true);
  }

  return fetch('/installation-service-modal')
    .then(response => response.text())
    .then(html => {
      const parser = new DOMParser();
      const doc = parser.parseFromString(html, 'text/html');
      const modal = doc.querySelector('#installation-modal');
      const style = doc.querySelector('style');

      if (modal) {
        container.appendChild(modal);
      }

      if (style) {
        container.appendChild(style);
      }

      return true;
    })
    .catch(error => {
      console.error('Failed to load installation service modal:', error);
      return false;
    });
}

function showInstallationServiceModal() {
  return loadInstallationServiceModal().then(() => {
    const dialog = document.querySelector('#cart-container #installation-modal');
    if (!dialog) {
      return;
    }

    if (dialog.open) {
      dialog.close();
    }

    dialog.showModal();

    setTimeout(() => {
      const content = dialog.querySelector('#installation-modal-content');
      if (content) {
        content.classList.remove('scale-95', 'opacity-0');
        content.classList.add('scale-100', 'opacity-100');
      }
    }, 10);

    installationServiceModalUISetup();
  });
}

function hideInstallationServiceModal() {
  const dialog = document.querySelector('#cart-container #installation-modal');
  if (dialog) {
    const content = dialog.querySelector('#installation-modal-content');
    if (content) {
      content.classList.remove('scale-100', 'opacity-100');
      content.classList.add('scale-95', 'opacity-0');
    }

    setTimeout(() => {
      dialog.close();
    }, 300);
  }
}

function loadCartModal() {
  const container = document.getElementById('cart-container');
  if (!container) {
    console.error('Cart container not found');
    return Promise.resolve(false);
  }

  const existingModal = container.querySelector('#cart-modal');
  if (existingModal) {
    return Promise.resolve(true);
  }

  return fetch('/cart-modal')
    .then(response => response.text())
    .then(html => {
      const parser = new DOMParser();
      const doc = parser.parseFromString(html, 'text/html');
      const modal = doc.querySelector('#cart-modal');
      const style = doc.querySelector('style');

      if (modal) {
        container.appendChild(modal);
      }

      if (style) {
        container.appendChild(style);
      }

      return true;
    })
    .catch(error => {
      console.error('Failed to load cart modal:', error);
      return false;
    });
}

function updateCartBadge() {
  const cart = getCart();
  // Exclude installation service from badge count
  const totalItems = cart.reduce((sum, item) =>
    item.name === INSTALLATION_SERVICE_NAME ? sum : sum + item.quantity, 0);

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
  // Don't call renderCart() here - it will be called when modal loads
  updateCartBadge();

  // Cart icon click is now handled by HTMX, but we need to show the dialog after it loads
  const cartIcon = document.getElementById('cart-icon');
  if (cartIcon) {
    cartIcon.addEventListener('click', () => {
      // Check if dialog is already loaded
      const dialog = document.querySelector('#cart-container #cart-modal');
      if (dialog) {
        showCart();
      } else {
        // HTMX will load the dialog, then we show it
        setTimeout(showCart, 100);
      }
    });
  }

  // Event listeners will be handled by the dialog's own script
  // but we keep fallbacks for when the dialog isn't loaded yet

  document.querySelectorAll('.add-to-cart').forEach(button => {
    button.addEventListener('click', (e) => {
      const productCard = e.target.closest('.product-card');
      const productName = productCard.dataset.name;
      const productId = productCard.dataset.id;
      const productPrice = parseFloat(productCard.dataset.price);
      addToCart(productName, productPrice, productId);
    });
  });

  // Event delegation for dynamically created buttons
  // This works both in the main page and in the dialog
  document.addEventListener('click', (e) => {
    const cartItems = document.getElementById('cart-items');
    if (cartItems && cartItems.contains(e.target)) {
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
    }
  });
});

function cartModalUISetup() {
  const modal = document.getElementById('cart-modal');
  if (!modal || modal.dataset.bound === 'true') {
    return;
  }

  modal.dataset.bound = 'true';

  const clearCartButton = document.getElementById('clear-cart');
  const closeButton = document.getElementById('close-cart');
  const proceedToCheckoutButton = document.getElementById('proceed-to-checkout');

  function handleCartClear() {
    if (confirm('Tem certeza que deseja limpar seu carrinho?')) {
      localStorage.setItem('cart', JSON.stringify([]));

      const badge = document.getElementById('cart-badge');
      if (badge) {
        badge.classList.add('hidden');
      }

      // Use renderCart() to handle all display updates including button styling
      renderCart();
    }
  }

  function handleCloseCart(event) {
    event?.preventDefault();
    hideCart();
  }

  function handleCartBackdrop(event) {
    if (event.target === event.currentTarget) {
      hideCart();
    }
  }

  function handleCartCancel(event) {
    event.preventDefault();
    hideCart();
  }

  function handleProceedToCheckout(event) {
    event.preventDefault();
    const cart = getCart();
    const displayCart = cart.filter(item => item.name !== INSTALLATION_SERVICE_NAME);

    if (displayCart.length === 0) {
      return; // Don't proceed if cart is empty
    }

    hideCart();
    showInstallationServiceModal();
  }

  clearCartButton?.addEventListener('click', handleCartClear);
  closeButton?.addEventListener('click', handleCloseCart);
  modal.addEventListener('click', handleCartBackdrop);
  modal.addEventListener('cancel', handleCartCancel);
  proceedToCheckoutButton?.addEventListener('click', handleProceedToCheckout);
}

function installationServiceModalUISetup() {
  const modal = document.getElementById('installation-modal');
  if (!modal || modal.dataset.bound === 'true') {
    return;
  }

  modal.dataset.bound = 'true';

  const addInstallationBtn = document.getElementById('add-installation');
  const skipInstallationBtn = document.getElementById('skip-installation');
  const closeBtn = document.getElementById('close-installation');
  let hasChosen = false;

  function proceed(includeInstallation) {
    if (hasChosen) {
      return;
    }
    hasChosen = true;

    if (includeInstallation) {
      const cart = getCart();
      const existingInstallation = cart.findIndex(item => item.name === INSTALLATION_SERVICE_NAME);

      if (existingInstallation === -1) {
        cart.push({ name: INSTALLATION_SERVICE_NAME, price: 120.00, quantity: 1 });
      } else {
        cart[existingInstallation].price = 120.00;
        cart[existingInstallation].quantity = 1;
      }

      saveCart(cart);
      updateCartBadge();
    } else {
      const cart = getCart();
      const filteredCart = cart.filter(item => item.name !== INSTALLATION_SERVICE_NAME);

      if (filteredCart.length !== cart.length) {
        saveCart(filteredCart);
        updateCartBadge();
      }
    }

    hideInstallationServiceModal();
    setTimeout(() => {
      window.location.href = '/checkout';
    }, 350);
  }

  addInstallationBtn?.addEventListener('click', (event) => {
    event.preventDefault();
    proceed(true);
  });

  skipInstallationBtn?.addEventListener('click', (event) => {
    event.preventDefault();
    proceed(false);
  });

  closeBtn?.addEventListener('click', (event) => {
    event.preventDefault();
    proceed(false);
  });

  modal.addEventListener('cancel', (event) => {
    event.preventDefault();
    proceed(false);
  });

  modal.addEventListener('click', (event) => {
    if (event.target === event.currentTarget) {
      proceed(false);
    }
  });
}

export { addToCart, removeFromCart, updateQuantity, clearCart, renderCart, showCart, hideCart, updateCartBadge, cartModalUISetup, loadInstallationServiceModal, showInstallationServiceModal, hideInstallationServiceModal, installationServiceModalUISetup };
