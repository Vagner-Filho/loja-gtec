// Import cart functions
function getCart() {
  return JSON.parse(localStorage.getItem('cart')) || [];
}

function saveCart(cart) {
  localStorage.setItem('cart', JSON.stringify(cart));
}

function clearCart() {
  localStorage.setItem('cart', JSON.stringify([]));
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

// Validate email
function validateEmail(email) {
  const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return re.test(email);
}

// Validate phone
function validatePhone(phone) {
  const cleaned = phone.replace(/\D/g, '');
  return cleaned.length >= 10;
}

// Validate card number (simple Luhn algorithm)
function validateCardNumber(cardNumber) {
  const cleaned = cardNumber.replace(/\s/g, '');
  if (!/^\d{13,19}$/.test(cleaned)) return false;

  let sum = 0;
  let isEven = false;

  for (let i = cleaned.length - 1; i >= 0; i--) {
    let digit = parseInt(cleaned.charAt(i), 10);

    if (isEven) {
      digit *= 2;
      if (digit > 9) {
        digit -= 9;
      }
    }

    sum += digit;
    isEven = !isEven;
  }

  return sum % 10 === 0;
}

// Validate expiry date
function validateExpiry(expiry) {
  const parts = expiry.split('/');
  if (parts.length !== 2) return false;

  const month = parseInt(parts[0], 10);
  const year = parseInt('20' + parts[1], 10);

  if (month < 1 || month > 12) return false;

  const now = new Date();
  const currentYear = now.getFullYear();
  const currentMonth = now.getMonth() + 1;

  if (year < currentYear) return false;
  if (year === currentYear && month < currentMonth) return false;

  return true;
}

// Validate CVV
function validateCVV(cvv) {
  return /^\d{3,4}$/.test(cvv);
}

// Show error message
function showError(input, message) {
  const parent = input.parentElement;
  let error = parent.querySelector('.error-message');

  if (!error) {
    error = document.createElement('p');
    error.className = 'error-message text-red-500 text-sm mt-1';
    parent.appendChild(error);
  }

  error.textContent = message;
  input.classList.add('border-red-500');
  input.classList.remove('border-gray-300');
}

// Clear error message
function clearError(input) {
  const parent = input.parentElement;
  const error = parent.querySelector('.error-message');

  if (error) {
    error.remove();
  }

  input.classList.remove('border-red-500');
  input.classList.add('border-gray-300');
}

// Render checkout items
function renderCheckoutItems() {
  const cart = getCart();
  const checkoutItemsContainer = document.getElementById('checkout-items');
  const emptyCartMessage = document.getElementById('empty-cart-message');
  const subtotalElement = document.getElementById('subtotal');
  const shippingElement = document.getElementById('shipping');
  const taxElement = document.getElementById('tax');
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
    itemElement.innerHTML = `
      <div class="flex-1">
        <h4 class="font-semibold text-gray-900">${item.name}</h4>
        <p class="text-sm text-gray-500">Quantidade: ${item.quantity}</p>
      </div>
      <p class="font-semibold text-gray-900">R$ ${(item.price * item.quantity).toFixed(2)}</p>
    `;
    checkoutItemsContainer.appendChild(itemElement);
    subtotal += item.price * item.quantity;
  });

  const shipping = cart.length > 0 ? 10.00 : 0;
  const tax = subtotal * 0.08; // 8% tax
  const total = subtotal + shipping + tax;

  subtotalElement.textContent = subtotal.toFixed(2);
  shippingElement.textContent = shipping.toFixed(2);
  taxElement.textContent = tax.toFixed(2);
  totalElement.textContent = total.toFixed(2);
}

// Generate order number
function generateOrderNumber() {
  const timestamp = Date.now().toString(36).toUpperCase();
  const random = Math.random().toString(36).substring(2, 5).toUpperCase();
  return `ORD-${timestamp}-${random}`;
}

// Handle form submission
function handleCheckout(e) {
  e.preventDefault();

  // Get form values
  const email = document.getElementById('email').value.trim();
  const phone = document.getElementById('phone').value.trim();
  const firstName = document.getElementById('firstName').value.trim();
  const lastName = document.getElementById('lastName').value.trim();
  const address = document.getElementById('address').value.trim();
  const city = document.getElementById('city').value.trim();
  const state = document.getElementById('state').value.trim();
  const zipCode = document.getElementById('zipCode').value.trim();

  // Get selected payment method
  const paymentMethod = document.querySelector('input[name="paymentMethod"]:checked')?.value;

  // Payment method specific fields
  const cardName = document.getElementById('cardName')?.value.trim() || '';
  const cardNumber = document.getElementById('cardNumber')?.value.trim() || '';
  const expiry = document.getElementById('expiry')?.value.trim() || '';
  const cvv = document.getElementById('cvv')?.value.trim() || '';
  const cpf = document.getElementById('cpf')?.value.trim() || '';
  const pixKey = document.getElementById('pixKey')?.value.trim() || '';

  let isValid = true;

  // Clear all previous errors
  document.querySelectorAll('input').forEach(input => clearError(input));

  // Validate email
  if (!email) {
    showError(document.getElementById('email'), 'Email é obrigatório');
    isValid = false;
  } else if (!validateEmail(email)) {
    showError(document.getElementById('email'), 'Por favor, insira um email válido');
    isValid = false;
  }

  // Validate phone
  if (!phone) {
    showError(document.getElementById('phone'), 'Telefone é obrigatório');
    isValid = false;
  } else if (!validatePhone(phone)) {
    showError(document.getElementById('phone'), 'Por favor, insira um telefone válido');
    isValid = false;
  }

  // Validate name fields
  if (!firstName) {
    showError(document.getElementById('firstName'), 'Nome é obrigatório');
    isValid = false;
  }

  if (!lastName) {
    showError(document.getElementById('lastName'), 'Sobrenome é obrigatório');
    isValid = false;
  }

  // Validate address
  if (!address) {
    showError(document.getElementById('address'), 'Endereço é obrigatório');
    isValid = false;
  }

  if (!city) {
    showError(document.getElementById('city'), 'Cidade é obrigatória');
    isValid = false;
  }

  if (!state) {
    showError(document.getElementById('state'), 'Estado é obrigatório');
    isValid = false;
  }

  if (!zipCode) {
    showError(document.getElementById('zipCode'), 'CEP é obrigatório');
    isValid = false;
  }

  // Validate payment method specific fields
  if (paymentMethod === 'credit_card') {
    if (!cardName) {
      showError(document.getElementById('cardName'), 'Nome no cartão é obrigatório');
      isValid = false;
    }

    if (!cardNumber) {
      showError(document.getElementById('cardNumber'), 'Número do cartão é obrigatório');
      isValid = false;
    } else if (!validateCardNumber(cardNumber)) {
      showError(document.getElementById('cardNumber'), 'Por favor, insira um número de cartão válido');
      isValid = false;
    }

    if (!expiry) {
      showError(document.getElementById('expiry'), 'Data de validade é obrigatória');
      isValid = false;
    } else if (!validateExpiry(expiry)) {
      showError(document.getElementById('expiry'), 'Por favor, insira uma data de validade válida');
      isValid = false;
    }

    if (!cvv) {
      showError(document.getElementById('cvv'), 'CVV é obrigatório');
      isValid = false;
    } else if (!validateCVV(cvv)) {
      showError(document.getElementById('cvv'), 'Por favor, insira um CVV válido');
      isValid = false;
    }
  } else if (paymentMethod === 'boleto') {
    if (!cpf) {
      showError(document.getElementById('cpf'), 'CPF/CNPJ é obrigatório');
      isValid = false;
    } else if (!validateCPF(cpf)) {
      showError(document.getElementById('cpf'), 'Por favor, insira um CPF ou CNPJ válido');
      isValid = false;
    }
  } else if (paymentMethod === 'pix') {
    if (!pixKey) {
      showError(document.getElementById('pixKey'), 'Chave PIX é obrigatória');
      isValid = false;
    }
  }

  if (!isValid) {
    // Scroll to first error
    const firstError = document.querySelector('.error-message');
    if (firstError) {
      firstError.closest('input').scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
    return;
  }

  // Show success message
  const checkoutForm = document.getElementById('checkout-form');
  const successMessage = document.getElementById('success-message');
  const orderNumber = document.getElementById('order-number');

  checkoutForm.classList.add('hidden');
  successMessage.classList.remove('hidden');
  orderNumber.textContent = generateOrderNumber();

  // Clear cart
  clearCart();

  // Scroll to top
  window.scrollTo({ top: 0, behavior: 'smooth' });
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
      paymentForms.forEach(form => form.classList.add('hidden'));

      // Show selected form
      const selectedForm = document.getElementById(`${selectedMethod}-form`);
      if (selectedForm) {
        selectedForm.classList.remove('hidden');
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

// Validate CPF
function validateCPF(cpf) {
  const cleaned = cpf.replace(/\D/g, '');
  return cleaned.length === 11 || cleaned.length === 14;
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

  // Place order button
  const placeOrderBtn = document.getElementById('place-order-btn');
  if (placeOrderBtn) {
    placeOrderBtn.addEventListener('click', handleCheckout);
  }

  // Clear error on input focus
  document.querySelectorAll('input').forEach(input => {
    input.addEventListener('focus', () => clearError(input));
  });
});
