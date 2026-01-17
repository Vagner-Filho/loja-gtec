import { saveCart, updateCartBadge } from "./cart.js";

const INSTALLATION_SERVICE_NAME = 'Serviço de Instalação';

// Import cart functions
function getCart() {
  return JSON.parse(localStorage.getItem('cart')) || [];
}

// Format card number with spaces
function formatCardNumber(value) {
  const v = value.replace(/\s+/g, '').replace(/[^0-9]/gi, '');
  const matches = v.match(/\d{4,16}/g);
  const match = (matches && matches[0]) || '';
  const parts = [];

  for (let i = 0, len = match.length; i < len; i += 4) {
    parts.push(match.substring(i, i + 4));
  }

  if (parts.length) {
    return parts.join(' ');
  } else {
    return value;
  }
}

// Format expiry date
function formatExpiry(value) {
  const v = value.replace(/\s+/g, '').replace(/[^0-9]/gi, '');
  if (v.length >= 2) {
    return v.slice(0, 2) + '/' + v.slice(2, 4);
  }
  return v;
}

// Format CPF/CNPJ
function formatCPF(value) {
  const cleaned = value.replace(/\D/g, '');

  if (cleaned.length <= 11) {
    // CPF format: 000.000.000-00
    return cleaned.replace(/(\d{3})(\d{3})(\d{3})(\d{2})/, '$1.$2.$3-$4');
  } else {
    // CNPJ format: 00.000.000/0000-00
    return cleaned.replace(/(\d{2})(\d{3})(\d{3})(\d{4})(\d{2})/, '$1.$2.$3/$4-$5');
  }
}

// Format CEP
function formatCEP(value) {
  const cleaned = value.replace(/\D/g, '');
  if (cleaned.length <= 5) {
    return cleaned;
  }
  return cleaned.slice(0, 5) + '-' + cleaned.slice(5, 8);
}

// Show zip code validation error
function showZipCodeError(message) {
  const zipCodeInput = document.getElementById('zipCode');
  const errorContainer = zipCodeInput.parentElement.querySelector('.error-message-container');

  // Add error styling to input
  zipCodeInput.classList.add('border-red-500', 'bg-red-50');
  zipCodeInput.classList.remove('border-gray-300', 'border-blue-500', 'bg-blue-50');

  // Display error message
  errorContainer.innerHTML = `<p class="error-message text-red-500 text-sm mt-1">${message}</p>`;
}

// Clear zip code validation error
function clearZipCodeError() {
  const zipCodeInput = document.getElementById('zipCode');
  const errorContainer = zipCodeInput.parentElement.querySelector('.error-message-container');

  // Remove error styling from input
  zipCodeInput.classList.remove('border-red-500', 'bg-red-50');
  zipCodeInput.classList.add('border-gray-300');

  // Clear error message
  errorContainer.innerHTML = '';
}

// Fetch address by CEP using ViaCEP API
async function fetchAddressByCEP(cep) {
  const zipCodeInput = document.getElementById('zipCode');
  const addressInput = document.getElementById('address');
  const neighborhoodInput = document.getElementById('neighborhood');
  const cityInput = document.getElementById('city');
  const stateInput = document.getElementById('state');

  // Clear any previous errors
  clearZipCodeError();

  // Show loading state
  zipCodeInput.classList.add('border-blue-500', 'bg-blue-50');
  zipCodeInput.classList.remove('border-gray-300');

  try {
    const response = await fetch(`https://viacep.com.br/ws/${cep}/json/`);
    const data = await response.json();

    // Clear loading state
    zipCodeInput.classList.remove('border-blue-500', 'bg-blue-50');
    zipCodeInput.classList.add('border-gray-300');

    // Check if CEP was found
    if (data.erro) {
      // Clear address fields but keep defaults for city/state
      addressInput.value = '';
      neighborhoodInput.value = '';
      cityInput.value = 'Campo Grande';
      stateInput.value = 'MS';
      return;
    }

    // Validate location - only serve Campo Grande, MS
    if (data.localidade !== 'Campo Grande' || data.uf !== 'MS') {
      showZipCodeError('Atendemos exclusivamente clientes em Campo Grande, MS');

      // Clear address fields but keep defaults for city/state
      addressInput.value = '';
      neighborhoodInput.value = '';
      cityInput.value = 'Campo Grande';
      stateInput.value = 'MS';
      return;
    } else {
      showInstallationServiceModal()
    }

    // Populate address fields
    if (data.logradouro) {
      addressInput.value = data.logradouro;
    }

    if (data.bairro) {
      neighborhoodInput.value = data.bairro;
    }

    // Keep city and state as Campo Grande, MS regardless of ViaCEP response
    cityInput.value = 'Campo Grande';
    stateInput.value = 'MS';

  } catch (error) {
    // Clear loading state
    zipCodeInput.classList.remove('border-blue-500', 'bg-blue-50');
    zipCodeInput.classList.add('border-gray-300');

    console.error('ViaCEP API error:', error);

    // Clear address fields on error but keep defaults for city/state
    addressInput.value = '';
    neighborhoodInput.value = '';
    cityInput.value = 'Campo Grande';
    stateInput.value = 'MS';
  }
}

// Payment method switching
function setupPaymentMethodSwitching() {
  const paymentRadios = document.querySelectorAll('input[name="paymentMethod"]');
  const paymentForms = document.querySelectorAll('.payment-form');
  const paymentOptionCards = document.querySelectorAll('.payment-option-card');

  paymentRadios.forEach(radio => {
    radio.addEventListener('change', (e) => {
      const selectedMethod = e.target.value;

      // Hide all forms
      paymentForms.forEach(form => {
        form.classList.add('hidden');
        form.setAttribute('disabled', 'true');
      });

      // Show selected form
      const selectedForm = document.getElementById(`${selectedMethod}-form`);
      if (selectedForm) {
        selectedForm.classList.remove('hidden');
        selectedForm.removeAttribute('disabled');
      }

      // Update card styling
      paymentOptionCards.forEach(card => {
        card.classList.remove('border-blue-500', 'bg-blue-50');
        card.classList.add('border-gray-200');
      });

      // Highlight selected card
      const selectedCard = e.target.closest('.payment-option').querySelector('.payment-option-card');
      if (selectedCard) {
        selectedCard.classList.remove('border-gray-200');
        selectedCard.classList.add('border-blue-500', 'bg-blue-50');
      }
    });
  });

  // Set initial state
  const defaultRadio = document.querySelector('input[name="paymentMethod"]:checked');
  if (defaultRadio) {
    defaultRadio.dispatchEvent(new Event('change'));
  }
}

// Render checkout items
function renderCheckoutItems() {
  const cart = getCart();
  const checkoutItemsContainer = document.getElementById('checkout-items');
  const emptyCartMessage = document.getElementById('empty-cart-message');
  const totalElement = document.getElementById('total');
  const placeOrderBtn = document.getElementById('place-order-btn');

  if (cart.length === 0) {
    checkoutItemsContainer.classList.add('hidden');
    emptyCartMessage.classList.remove('hidden');
    placeOrderBtn.disabled = true;
    return;
  }

  checkoutItemsContainer.classList.remove('hidden');
  emptyCartMessage.classList.add('hidden');
  placeOrderBtn.disabled = false;

  checkoutItemsContainer.innerHTML = '';
  let subtotal = 0;

  cart.forEach(item => {
    const itemElement = document.createElement('div');
    itemElement.className = 'flex justify-between items-start pb-4 border-b border-gray-200';
    if (item.name !== INSTALLATION_SERVICE_NAME) {
      itemElement.innerHTML = `
        <div class="flex-1">
          <h4 class="font-semibold text-gray-900">${item.name}</h4>
          <p class="text-sm text-gray-500">Quantidade: ${item.quantity}</p>
        </div>
        <p class="font-semibold text-gray-900">R$ ${(item.price * item.quantity).toFixed(2)}</p>
      `
    } else {
      itemElement.innerHTML = `
        <div class="flex-1">
          <h4 class="font-semibold text-gray-900">${item.name}</h4>
        </div>
        <p class="font-semibold text-gray-900">R$ ${item.price.toFixed(2)}</p>
      `;
    }
    checkoutItemsContainer.appendChild(itemElement);
    subtotal += item.price * item.quantity;
  });

  totalElement.textContent = subtotal.toFixed(2);
}

// Initialize
document.addEventListener('DOMContentLoaded', () => {
  renderCheckoutItems();
  setupPaymentMethodSwitching();

  // Card number formatting
  const cardNumberInput = document.getElementById('cardNumber');
  if (cardNumberInput) {
    cardNumberInput.addEventListener('input', (e) => {
      e.target.value = formatCardNumber(e.target.value);
    });
  }

  // Expiry formatting
  const expiryInput = document.getElementById('expiry');
  if (expiryInput) {
    expiryInput.addEventListener('input', (e) => {
      e.target.value = formatExpiry(e.target.value);
    });
  }

  // CVV - numbers only
  const cvvInput = document.getElementById('cvv');
  if (cvvInput) {
    cvvInput.addEventListener('input', (e) => {
      e.target.value = e.target.value.replace(/\D/g, '');
    });
  }

  // CPF formatting
  const cpfInput = document.getElementById('cpf');
  if (cpfInput) {
    cpfInput.addEventListener('input', (e) => {
      e.target.value = formatCPF(e.target.value);
    });
  }

  // CEP formatting and ViaCEP integration
  const zipCodeInput = document.getElementById('zipCode');
  if (zipCodeInput) {
    zipCodeInput.addEventListener('input', (e) => {
      const originalValue = e.target.value;
      const cleanedCEP = originalValue.replace(/\D/g, '');

      // Format CEP in real-time
      e.target.value = formatCEP(originalValue);

      // Clear any previous zip code errors when user starts typing
      if (cleanedCEP.length < 8) {
        clearZipCodeError();
      }

      // Trigger API call when CEP reaches 8 digits
      if (cleanedCEP.length === 8) {
        fetchAddressByCEP(cleanedCEP);
      } else if (cleanedCEP.length < 8) {
        // Clear address fields if CEP becomes incomplete
        const addressInput = document.getElementById('address');
        const neighborhoodInput = document.getElementById('neighborhood');
        const cityInput = document.getElementById('city');
        const stateInput = document.getElementById('state');

        // Only clear if the fields were previously filled by ViaCEP
        // (check if they're currently filled and user is now editing CEP)
        if (addressInput.value && neighborhoodInput.value) {
          addressInput.value = '';
          neighborhoodInput.value = '';
        }
        // Always reset city/state to defaults
        cityInput.value = 'Campo Grande';
        stateInput.value = 'MS';
      }
    });
  }

  // Prevent manual editing of city and state fields
  const cityInput = document.getElementById('city');
  const stateInput = document.getElementById('state');

  if (cityInput) {
    cityInput.addEventListener('input', (e) => {
      e.target.value = 'Campo Grande';
    });

    cityInput.addEventListener('focus', (e) => {
      e.target.select();
    });
  }

  if (stateInput) {
    stateInput.addEventListener('input', (e) => {
      e.target.value = 'MS';
    });

    stateInput.addEventListener('focus', (e) => {
      e.target.select();
    });
  }
});

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

function installationServiceModalUISetup() {
  const modal = document.getElementById('installation-modal');
  if (!modal || modal.dataset.bound === 'true') {
    return;
  }

  modal.dataset.bound = 'true';

  const addInstallationBtn = document.getElementById('add-installation');
  const skipInstallationBtn = document.getElementById('skip-installation');
  const closeBtn = document.getElementById('close-installation');
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
  renderCheckoutItems();
  hideInstallationServiceModal();
  /*setTimeout(() => {
    window.location.href = '/checkout';
  }, 350);*/
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
