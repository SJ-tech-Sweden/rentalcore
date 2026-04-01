/*
 * RentalCore Design System JavaScript
 * Handles interactions for dropdowns, navigation, and components
 */

if (typeof window.RentalCoreDesign === 'undefined') {
class RentalCoreDesign {
    constructor() {
        this.init();
    }

    init() {
        this.initDropdowns();
        this.initMobileNav();
        this.initAnimations();
        this.initTooltips();
        this.handleClickOutside();
        this.initThemeToggle();
        this.initSearchableSelects();
    }

    // Dropdown functionality
    initDropdowns() {
        console.log('Initializing dropdowns...');
        const dropdowns = document.querySelectorAll('.rc-dropdown');
        console.log('Found dropdowns:', dropdowns.length);
        
        dropdowns.forEach((dropdown, index) => {
            // Skip if already initialized
            if (dropdown.dataset.initialized === 'true') {
                console.log('Dropdown', index, 'already initialized, skipping');
                return;
            }
            
            const toggle = dropdown.querySelector('.rc-dropdown-toggle');
            const menu = dropdown.querySelector('.rc-dropdown-menu');
            
            console.log('Dropdown', index, '- Toggle:', !!toggle, 'Menu:', !!menu);
            
            if (toggle && menu) {
                // Mark as initialized
                dropdown.dataset.initialized = 'true';
                
                toggle.addEventListener('click', (e) => {
                    console.log('Dropdown toggle clicked');
                    e.preventDefault();
                    e.stopPropagation();

                    // Check if we're on mobile
                    const isMobile = window.innerWidth <= 768;
                    if (isMobile) {
                        // On mobile, only toggle - don't redirect
                        this.toggleDropdown(dropdown);
                    } else {
                        // On desktop, normal dropdown behavior
                        this.toggleDropdown(dropdown);
                    }
                });

                // Close dropdown when clicking menu items (except headers and dividers)
                const items = menu.querySelectorAll('.rc-dropdown-item');
                console.log('Dropdown', index, 'has', items.length, 'items');
                items.forEach(item => {
                    item.addEventListener('click', (e) => {
                        console.log('Dropdown item clicked:', item.textContent);
                        if (!item.classList.contains('rc-dropdown-header') && 
                            !item.classList.contains('rc-dropdown-divider')) {
                            this.closeDropdown(dropdown);
                            // Don't prevent default - allow navigation to occur
                        }
                    });
                });
            }
        });
    }

    toggleDropdown(dropdown) {
        const isOpen = dropdown.classList.contains('show');
        console.log('Toggling dropdown, currently open:', isOpen);
        
        // Close all other dropdowns
        this.closeAllDropdowns();
        
        if (!isOpen) {
            console.log('Opening dropdown');
            dropdown.classList.add('show');

            // Don't auto-focus items - let users navigate naturally
        } else {
            console.log('Dropdown was already open, now closed by closeAllDropdowns');
        }
    }

    closeDropdown(dropdown) {
        dropdown.classList.remove('show');
    }

    closeAllDropdowns() {
        const dropdowns = document.querySelectorAll('.rc-dropdown.show');
        dropdowns.forEach(dropdown => this.closeDropdown(dropdown));
    }

    // Mobile navigation
    initMobileNav() {
        const toggle = document.querySelector('.rc-navbar-toggle');
        const nav = document.querySelector('.rc-navbar-nav');
        
        if (toggle && nav) {
            toggle.addEventListener('click', (e) => {
                e.stopPropagation();
                nav.classList.toggle('show');
                
                // Update toggle icon
                const icon = toggle.querySelector('i');
                if (icon) {
                    if (nav.classList.contains('show')) {
                        icon.className = 'bi bi-x-lg';
                    } else {
                        icon.className = 'bi bi-list';
                    }
                }
            });

            // Close mobile nav when clicking nav items
            const navItems = nav.querySelectorAll('a');
            navItems.forEach(item => {
                item.addEventListener('click', () => {
                    nav.classList.remove('show');
                    const icon = toggle.querySelector('i');
                    if (icon) {
                        icon.className = 'bi bi-list';
                    }
                });
            });
        }
    }

    // Click outside handler
    handleClickOutside() {
        document.addEventListener('click', (e) => {
            // Close dropdowns when clicking outside
            if (!e.target.closest('.rc-dropdown')) {
                this.closeAllDropdowns();
            }
            
            // Close mobile nav when clicking outside or on backdrop
            const nav = document.querySelector('.rc-navbar-nav');
            const toggle = document.querySelector('.rc-navbar-toggle');
            if (nav && nav.classList.contains('show') &&
                !e.target.closest('.rc-navbar-nav') &&
                !e.target.closest('.rc-navbar-toggle')) {
                nav.classList.remove('show');
                const icon = toggle?.querySelector('i');
                if (icon) {
                    icon.className = 'bi bi-list';
                }
            }
        });
    }

    // Animation utilities
    initAnimations() {
        // Fade in animations on scroll
        const observerOptions = {
            threshold: 0.1,
            rootMargin: '0px 0px -50px 0px'
        };

        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    entry.target.classList.add('rc-animate-fade-in');
                }
            });
        }, observerOptions);

        // Observe cards and sections
        const elements = document.querySelectorAll('.rc-card, .rc-section, .rc-hero');
        elements.forEach(el => observer.observe(el));

        // Button loading states
        this.initButtonStates();
    }

    initButtonStates() {
        const buttons = document.querySelectorAll('.rc-btn');
        
        buttons.forEach(button => {
            button.addEventListener('click', function(e) {
                // Only apply loading state if explicitly requested via data-loading attribute
                if (this.dataset.loading === 'true') {
                    this.classList.add('rc-loading');
                    
                    // Store original content
                    if (!this.dataset.originalContent) {
                        this.dataset.originalContent = this.innerHTML;
                    }
                    
                    // Add loading spinner
                    this.innerHTML = '<i class="bi bi-hourglass-split"></i> Loading...';
                    this.disabled = true;
                    
                    // Auto-restore after 5 seconds if no form submission
                    setTimeout(() => {
                        if (this.classList.contains('rc-loading')) {
                            this.innerHTML = this.dataset.originalContent;
                            this.disabled = false;
                            this.classList.remove('rc-loading');
                        }
                    }, 5000);
                }
                // For submit buttons, don't interfere with normal form submission
                // Let the form handle the submission naturally
            });
        });
    }

    // Tooltip functionality
    initTooltips() {
        const tooltipElements = document.querySelectorAll('[data-tooltip]');
        
        tooltipElements.forEach(element => {
            let tooltip = null;
            
            element.addEventListener('mouseenter', () => {
                const text = element.dataset.tooltip;
                if (text) {
                    tooltip = this.createTooltip(text);
                    document.body.appendChild(tooltip);
                    this.positionTooltip(element, tooltip);
                }
            });
            
            element.addEventListener('mouseleave', () => {
                if (tooltip) {
                    tooltip.remove();
                    tooltip = null;
                }
            });
        });
    }

    createTooltip(text) {
        const tooltip = document.createElement('div');
        tooltip.className = 'rc-tooltip';
        tooltip.textContent = text;
        tooltip.style.cssText = `
            position: absolute;
            background: var(--surface-1);
            color: var(--text-primary);
            padding: var(--space-xs) var(--space-sm);
            border-radius: var(--radius-sm);
            font-size: 0.75rem;
            box-shadow: var(--shadow-md);
            z-index: 1060;
            pointer-events: none;
            white-space: nowrap;
            opacity: 0;
            transition: opacity var(--transition-fast);
            border: 1px solid var(--surface-3);
        `;
        
        // Trigger fade in
        setTimeout(() => tooltip.style.opacity = '1', 10);
        
        return tooltip;
    }

    positionTooltip(element, tooltip) {
        const rect = element.getBoundingClientRect();
        const tooltipRect = tooltip.getBoundingClientRect();
        
        let top = rect.top - tooltipRect.height - 8;
        let left = rect.left + (rect.width / 2) - (tooltipRect.width / 2);
        
        // Adjust if tooltip goes off screen
        if (top < 0) {
            top = rect.bottom + 8;
        }
        if (left < 0) {
            left = 8;
        }
        if (left + tooltipRect.width > window.innerWidth) {
            left = window.innerWidth - tooltipRect.width - 8;
        }
        
        tooltip.style.top = `${top + window.scrollY}px`;
        tooltip.style.left = `${left}px`;
    }

    // Keyboard navigation
    initKeyboardNav() {
        document.addEventListener('keydown', (e) => {
            // ESC to close dropdowns
            if (e.key === 'Escape') {
                this.closeAllDropdowns();
                
                const nav = document.querySelector('.rc-navbar-nav');
                if (nav && nav.classList.contains('show')) {
                    nav.classList.remove('show');
                    const toggle = document.querySelector('.rc-navbar-toggle');
                    const icon = toggle?.querySelector('i');
                    if (icon) {
                        icon.className = 'bi bi-list';
                    }
                }
            }
            
            // Arrow key navigation in dropdowns
            if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
                const openDropdown = document.querySelector('.rc-dropdown.show');
                if (openDropdown) {
                    e.preventDefault();
                    this.navigateDropdown(openDropdown, e.key === 'ArrowDown');
                }
            }
        });
    }

    navigateDropdown(dropdown, down) {
        const items = dropdown.querySelectorAll('.rc-dropdown-item:not(.rc-dropdown-header)');
        const currentFocus = document.activeElement;
        let index = Array.from(items).indexOf(currentFocus);
        
        if (down) {
            index = index < items.length - 1 ? index + 1 : 0;
        } else {
            index = index > 0 ? index - 1 : items.length - 1;
        }
        
        items[index]?.focus();
    }

    // Form enhancements
    initForms() {
        // Floating labels
        const floatingInputs = document.querySelectorAll('.floating-input');
        floatingInputs.forEach(input => {
            // Check initial state
            this.updateFloatingLabel(input);
            
            input.addEventListener('blur', () => this.updateFloatingLabel(input));
            input.addEventListener('focus', () => this.updateFloatingLabel(input));
            input.addEventListener('input', () => this.updateFloatingLabel(input));
        });
    }

    updateFloatingLabel(input) {
        const hasValue = input.value.trim() !== '';
        const label = input.nextElementSibling;
        
        if (label && label.classList.contains('floating-label-text')) {
            if (hasValue || input === document.activeElement) {
                label.style.transform = 'translateY(-150%)';
                label.style.fontSize = '0.75rem';
                label.style.color = 'var(--accent-electric)';
            } else {
                label.style.transform = 'translateY(-50%)';
                label.style.fontSize = '1rem';
                label.style.color = 'var(--text-muted)';
            }
        }
    }

    // Theme toggle functionality
    initThemeToggle() {
        const themeToggle = document.querySelector('[data-theme-toggle]');
        if (themeToggle) {
            themeToggle.addEventListener('click', () => {
                const currentTheme = document.documentElement.getAttribute('data-theme');
                const newTheme = currentTheme === 'light' ? 'dark' : 'light';
                
                document.documentElement.setAttribute('data-theme', newTheme);
                localStorage.setItem('rc-theme', newTheme);
                
                // Update icon
                const icon = themeToggle.querySelector('i');
                if (icon) {
                    icon.className = newTheme === 'dark' ? 'bi bi-sun' : 'bi bi-moon-fill';
                }
            });
        }
        
        // Update icon to match current theme (theme is already set by immediate script)
        if (themeToggle) {
            const currentTheme = document.documentElement.getAttribute('data-theme');
            const icon = themeToggle.querySelector('i');
            if (icon) {
                icon.className = currentTheme === 'dark' ? 'bi bi-sun' : 'bi bi-moon-fill';
            }
        }
    }

    // Searchable select initialization
    initSearchableSelects() {
        document.querySelectorAll('select[data-searchable]').forEach(sel => {
            new RCSearchableSelect(sel);
        });

        // Watch for dynamically added searchable selects (e.g. cloned device rows)
        const observer = new MutationObserver(mutations => {
            mutations.forEach(mutation => {
                mutation.addedNodes.forEach(node => {
                    if (node.nodeType !== 1) return;
                    if (node.matches && node.matches('select[data-searchable]')) {
                        new RCSearchableSelect(node);
                    }
                    if (node.querySelectorAll) {
                        node.querySelectorAll('select[data-searchable]').forEach(sel => {
                            new RCSearchableSelect(sel);
                        });
                    }
                });
            });
        });
        observer.observe(document.body, { childList: true, subtree: true });
    }
}

// Searchable Select Component
class RCSearchableSelect {
    // Fix 4: shared state for a single document-level click handler
    static _instances = new Set();
    static _docListenerRegistered = false;
    static _idCounter = 0;

    constructor(selectEl) {
        if (selectEl._rcSearchable) return;
        selectEl._rcSearchable = this;
        this.select = selectEl;
        this.isOpen = false;
        this._ignoreChange = false;
        this._build();
        this._observe();
    }

    _build() {
        const select = this.select;

        this.wrapper = document.createElement('div');
        this.wrapper.className = 'rc-searchable-select';
        // Inherit inline width/max-width from the original select
        if (select.style.maxWidth) this.wrapper.style.maxWidth = select.style.maxWidth;
        if (select.style.width) this.wrapper.style.width = select.style.width;

        this.display = document.createElement('button');
        this.display.type = 'button';
        this.display.className = 'rc-searchable-select__display';
        this.display.setAttribute('aria-haspopup', 'listbox');
        this.display.setAttribute('aria-expanded', 'false');

        // Fix 2 + Fix 3: re-associate any <label for="select.id"> with the custom button
        if (select.id) {
            const btnId = select.id + '-rcss-btn';
            this.display.id = btnId;
            try {
                document.querySelectorAll(`label[for="${CSS.escape(select.id)}"]`).forEach(label => {
                    label.setAttribute('for', btnId);
                });
            } catch (_) { /* CSS.escape not available in very old browsers */ }
        }
        if (select.getAttribute('aria-label')) {
            this.display.setAttribute('aria-label', select.getAttribute('aria-label'));
        }

        this.dropdown = document.createElement('div');
        this.dropdown.className = 'rc-searchable-select__dropdown';
        // Note: no role here — the listbox role belongs on the options list only

        this.searchInput = document.createElement('input');
        this.searchInput.type = 'text';
        this.searchInput.className = 'rc-searchable-select__search';
        this.searchInput.placeholder = 'Search...';
        this.searchInput.setAttribute('autocomplete', 'off');
        this.searchInput.setAttribute('aria-label', 'Search options');

        // Fix 3: role="listbox" on the options container, not the enclosing dropdown div
        const listboxId = 'rc-ss-lb-' + (++RCSearchableSelect._idCounter);
        this.optionsList = document.createElement('div');
        this.optionsList.className = 'rc-searchable-select__options';
        this.optionsList.id = listboxId;
        this.optionsList.setAttribute('role', 'listbox');
        this.display.setAttribute('aria-controls', listboxId);

        this.dropdown.appendChild(this.searchInput);
        this.dropdown.appendChild(this.optionsList);
        this.wrapper.appendChild(this.display);
        this.wrapper.appendChild(this.dropdown);

        // Fix 2: visually hide the native select (not display:none) so label[for] clicks still work
        select.style.position = 'absolute';
        select.style.opacity = '0';
        select.style.width = '1px';
        select.style.height = '1px';
        select.style.margin = '0';
        select.style.padding = '0';
        select.style.border = '0';
        select.style.clip = 'rect(0 0 0 0)';
        select.style.clipPath = 'inset(50%)';
        select.style.overflow = 'hidden';

        select.parentNode.insertBefore(this.wrapper, select);
        this.wrapper.appendChild(select);

        this._renderOptions();
        this._updateDisplay();
        this._bindEvents();
    }

    _renderOptions(filter = '') {
        this.optionsList.innerHTML = '';
        const lower = filter.toLowerCase();

        Array.from(this.select.options).forEach((opt, i) => {
            if (filter && !opt.text.toLowerCase().includes(lower)) return;

            const item = document.createElement('div');
            item.className = 'rc-searchable-select__option';
            item.setAttribute('role', 'option');
            item.setAttribute('tabindex', '-1');
            // Fix 3: proper aria-selected for screen readers
            item.setAttribute('aria-selected', opt.selected ? 'true' : 'false');
            item.dataset.index = i;
            if (opt.selected) item.classList.add('selected');

            if (filter) {
                const text = opt.text;
                const idx = text.toLowerCase().indexOf(lower);
                if (idx >= 0) {
                    item.innerHTML =
                        this._escapeHtml(text.substring(0, idx)) +
                        '<mark>' + this._escapeHtml(text.substring(idx, idx + filter.length)) + '</mark>' +
                        this._escapeHtml(text.substring(idx + filter.length));
                } else {
                    item.textContent = text;
                }
            } else {
                item.textContent = opt.text;
            }

            item.addEventListener('mousedown', (e) => { e.preventDefault(); this._selectOption(i); });
            this.optionsList.appendChild(item);
        });
    }

    _escapeHtml(str) {
        return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
    }

    _selectOption(index) {
        // Fix 1: guard flag so the external 'change' listener skips during internal selection
        this._ignoreChange = true;
        this.select.selectedIndex = index;
        this.select.dispatchEvent(new Event('change', { bubbles: true }));
        this._ignoreChange = false;
        this._updateDisplay();
        this._close();
    }

    _updateDisplay() {
        const sel = this.select.options[this.select.selectedIndex];
        const text = sel ? sel.text : '';
        const hasValue = sel && sel.value !== '';
        this.display.innerHTML =
            '<span class="rc-searchable-select__label' + (hasValue ? '' : ' placeholder') + '">' +
            this._escapeHtml(text) + '</span>' +
            '<i class="bi bi-chevron-down rc-searchable-select__chevron"></i>';
    }

    _open() {
        if (this.isOpen) return;
        // Close any other open searchable selects
        document.querySelectorAll('.rc-searchable-select.open').forEach(w => {
            if (w !== this.wrapper && w._rcss) w._rcss._close();
        });
        this.isOpen = true;
        this.wrapper.classList.add('open');
        this.display.setAttribute('aria-expanded', 'true');
        this.searchInput.value = '';
        this._renderOptions();
        requestAnimationFrame(() => {
            this.searchInput.focus();
            const selected = this.optionsList.querySelector('.selected');
            if (selected) selected.scrollIntoView({ block: 'nearest' });
        });
    }

    _close() {
        if (!this.isOpen) return;
        this.isOpen = false;
        this.wrapper.classList.remove('open');
        this.display.setAttribute('aria-expanded', 'false');
    }

    _bindEvents() {
        this.wrapper._rcss = this;

        this.display.addEventListener('click', (e) => {
            e.stopPropagation();
            this.isOpen ? this._close() : this._open();
        });

        this.searchInput.addEventListener('input', () => {
            this._renderOptions(this.searchInput.value);
        });

        this.searchInput.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') { this._close(); this.display.focus(); }
            if (e.key === 'ArrowDown') {
                e.preventDefault();
                const first = this.optionsList.querySelector('.rc-searchable-select__option');
                if (first) first.focus();
            }
        });

        this.optionsList.addEventListener('keydown', (e) => {
            const focused = document.activeElement;
            if (!focused.classList.contains('rc-searchable-select__option')) return;
            if (e.key === 'ArrowDown') {
                e.preventDefault();
                const next = focused.nextElementSibling;
                if (next) next.focus();
            } else if (e.key === 'ArrowUp') {
                e.preventDefault();
                const prev = focused.previousElementSibling;
                if (prev) prev.focus(); else this.searchInput.focus();
            } else if (e.key === 'Enter') {
                e.preventDefault();
                const idx = parseInt(focused.dataset.index, 10);
                this._selectOption(idx);
            } else if (e.key === 'Escape') {
                this._close();
                this.display.focus();
            }
        });

        // Fix 2: forward focus from the visually-hidden native select to the custom button
        this.select.addEventListener('focus', () => {
            this.display.focus();
        });

        // Fix 1: sync display and option highlighting when external code changes the select value
        this.select.addEventListener('change', () => {
            if (this._ignoreChange) return;
            this._updateDisplay();
            if (this.isOpen) this._renderOptions(this.searchInput.value);
        });

        // Fix 4: register this instance in the shared set; ensure only one document handler exists
        RCSearchableSelect._instances.add(this);
        if (!RCSearchableSelect._docListenerRegistered) {
            RCSearchableSelect._docListenerRegistered = true;
            document.addEventListener('click', (e) => {
                RCSearchableSelect._instances.forEach(inst => {
                    // Clean up instances whose wrapper has been removed from the DOM
                    if (!inst.wrapper.isConnected) {
                        RCSearchableSelect._instances.delete(inst);
                        return;
                    }
                    if (!inst.wrapper.contains(e.target)) inst._close();
                });
            });
        }
    }

    _observe() {
        // Re-render when options are dynamically updated
        this._observer = new MutationObserver(() => {
            this._renderOptions(this.searchInput ? this.searchInput.value : '');
            this._updateDisplay();
        });
        this._observer.observe(this.select, {
            childList: true,
            subtree: true,
            attributes: true,
            attributeFilter: ['selected', 'disabled']
        });
    }
}

// Expose class to the global scope so subsequent loads can detect it
window.RentalCoreDesign = RentalCoreDesign;

// Initialize when DOM is ready (only once)
document.addEventListener('DOMContentLoaded', () => {
    if (!window.rcDesign) {
        // Use the global class reference so it's safe even if the file
        // was concatenated or loaded multiple times.
        window.rcDesign = new (window.RentalCoreDesign)();
    }
});

// Utility functions for external use
window.RentalCore = {
    showNotification: function(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `rc-notification rc-notification-${type}`;
        notification.innerHTML = `
            <div class="rc-notification-content">
                <i class="bi bi-${type === 'error' ? 'exclamation-triangle' : type === 'success' ? 'check-circle' : 'info-circle'}"></i>
                <span>${message}</span>
            </div>
            <button class="rc-notification-close">
                <i class="bi bi-x"></i>
            </button>
        `;
        
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            background: var(--surface-1);
            border: 1px solid var(--surface-3);
            border-radius: var(--radius-md);
            padding: var(--space-md);
            box-shadow: var(--shadow-lg);
            z-index: 1070;
            min-width: 300px;
            opacity: 0;
            transform: translateX(100%);
            transition: var(--transition-normal);
        `;
        
        document.body.appendChild(notification);
        
        // Trigger animation
        setTimeout(() => {
            notification.style.opacity = '1';
            notification.style.transform = 'translateX(0)';
        }, 10);
        
        // Auto-remove after 5 seconds
        setTimeout(() => {
            notification.style.opacity = '0';
            notification.style.transform = 'translateX(100%)';
            setTimeout(() => notification.remove(), 250);
        }, 5000);
        
        // Close button
        const closeBtn = notification.querySelector('.rc-notification-close');
        closeBtn.addEventListener('click', () => {
            notification.style.opacity = '0';
            notification.style.transform = 'translateX(100%)';
            setTimeout(() => notification.remove(), 250);
        });
    },
    
    openModal: function(content, title = '') {
        // Modal implementation would go here
        console.log('Modal:', title, content);
    },
    
    showLoader: function(element) {
        element.classList.add('rc-loading');
    },
    
    hideLoader: function(element) {
        element.classList.remove('rc-loading');
    }
};
}
