// Toast Notification System for HTMX

(function() {
  'use strict';

  const ToastManager = {
    container: null,
    toasts: [],
    maxToasts: 5,
    defaultDuration: 4000,

    init: function() {
      this.createContainer();
      this.setupHTMXListeners();
    },

    createContainer: function() {
      if (this.container) return;
      this.container = document.createElement('div');
      this.container.id = 'toast-container';
      this.container.className = 'fixed top-4 right-4 z-[9999] flex flex-col gap-2 pointer-events-none';
      document.body.appendChild(this.container);
    },

    setupHTMXListeners: function() {
      document.body.addEventListener('htmx:afterRequest', (evt) => {
        const xhr = evt.detail.xhr;
        const message = xhr.getResponseHeader('X-Toast-Message');
        if (!message) return;

        let type = xhr.getResponseHeader('X-Toast-Type');
        if (!type) {
          const status = xhr.status;
          type = status >= 400 ? 'error' : 'success';
        }

        this.show(message, type);
      });
    },

    show: function(message, type = 'success') {
      const toast = this.createToast(message, type);
      this.toasts.push(toast);

      if (this.toasts.length > this.maxToasts) {
        const oldest = this.toasts.shift();
        this.removeToast(oldest);
      }

      requestAnimationFrame(() => {
        toast.classList.add('toast-enter');
      });

      setTimeout(() => {
        this.dismiss(toast);
      }, this.defaultDuration);

      return toast;
    },

    createToast: function(message, type) {
      const toast = document.createElement('div');
      toast.className = 'toast pointer-events-auto flex items-center gap-3 px-4 py-3 rounded-lg shadow-lg transform translate-x-full opacity-0 transition-all duration-300 ease-in-out';

      const colors = {
        success: 'bg-green-700 text-white',
        error: 'bg-red-500 text-white',
        warning: 'bg-yellow-500 text-gray-900'
      };

      const icons = {
        success: '<svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/></svg>',
        error: '<svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>',
        warning: '<svg class="w-5 h-5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/></svg>'
      };

      toast.className += ' ' + (colors[type] || colors.success);
      toast.innerHTML = `
        ${icons[type] || icons.success}
        <span class="font-semibold">${this.escapeHtml(message)}</span>
      `;

      toast.addEventListener('click', () => this.dismiss(toast));

      this.container.appendChild(toast);
      return toast;
    },

    dismiss: function(toast) {
      if (!toast || !toast.parentNode) return;

      toast.classList.remove('toast-enter');
      toast.classList.add('toast-exit');

      setTimeout(() => {
        this.removeToast(toast);
      }, 300);
    },

    removeToast: function(toast) {
      if (!toast || !toast.parentNode) return;
      toast.parentNode.removeChild(toast);
      const index = this.toasts.indexOf(toast);
      if (index > -1) {
        this.toasts.splice(index, 1);
      }
    },

    escapeHtml: function(text) {
      const div = document.createElement('div');
      div.textContent = text;
      return div.innerHTML;
    }
  };

  document.addEventListener('DOMContentLoaded', () => {
    ToastManager.init();
  });

  window.ToastManager = ToastManager;
  window.showToast = function(message, type) {
    ToastManager.show(message, type);
  };
})();
