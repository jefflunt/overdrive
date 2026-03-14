---
title: Implement PWA Support for overdrive
status: done
type: feature
---

# Feature Overview
The goal is to enable Progressive Web App (PWA) capabilities for the overdrive application. This will allow users to install the app on their mobile or desktop home screens, providing a more native-like experience.

### User Story
As a developer using overdrive, I want to install the app on my device so that I can quickly access my build jobs and project settings without opening a browser tab.

### Out of Scope
- Full offline data synchronization for jobs.
- Native push notifications (can be a separate feature).

---

# Architecture & Design

### New Files
- `static/manifest.json`: Web app manifest for PWA identification.
- `static/sw.js`: Service worker for offline capability and installation requirements.
- `static/icons/icon-192.png`: App icon for splash screens and home screen (192x192).
- `static/icons/icon-512.png`: App icon for splash screens and home screen (512x512).

### Modified Files
- `api/main.go`: Added routes to serve the manifest, service worker, and static icons.
- `api/templates/layout.html`: Added PWA meta tags, manifest link, service worker registration, and "Add to Home Screen" logic.

---

# Step-by-Step Implementation Plan

## Step 0: Safety & Setup
1. Create a new branch: `git checkout -b feature/pwa-support`.

## Phase 1: Foundation (Backend & Assets)

### Step 1.1: Create Static Directory and Manifest
Create the `static/` directory and the `manifest.json` file.
- **File**: `static/manifest.json`
- **Content**:
```json
{
  "name": "overdrive",
  "short_name": "overdrive",
  "description": "Go + HTMX Build Automation Tool",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#0A0A0B",
  "theme_color": "#00A36C",
  "icons": [
    {
      "src": "/static/icons/icon-192.png",
      "sizes": "192x192",
      "type": "image/png"
    },
    {
      "src": "/static/icons/icon-512.png",
      "sizes": "512x512",
      "type": "image/png"
    }
  ]
}
```
- **Verification**: `ls static/manifest.json`

### Step 1.2: Implement Service Worker
Create a basic service worker that satisfies PWA criteria.
- **File**: `static/sw.js`
- **Logic**: 
  - Install event: Pre-cache the offline fallback or main assets.
  - Fetch event: Network-first strategy with cache fallback.
- **Verification**: `cat static/sw.js`

### Step 1.3: Configure Go Server to Serve Static Files
Update the Go server to serve the `static` directory and specifically expose `manifest.json` and `sw.js` at the root.
- **File**: `api/main.go`
- **Action**: 
  - Near line 50 (after `mux := http.NewServeMux()`), add:
    ```go
    mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    mux.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "static/manifest.json")
    })
    mux.HandleFunc("/sw.js", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "static/sw.js")
    })
    ```
- **Verification**: Run `go run ./api` and visit `http://localhost:3281/manifest.json` and `http://localhost:3281/sw.js`.

### Step 1.4: Add Placeholder Icons
Create the icons directory and place temporary icons (or instructions to add them).
- **Files**: `static/icons/icon-192.png`, `static/icons/icon-512.png`
- **Note**: Since binary files cannot be easily created via this interface, use a placeholder or a simple Go script to generate a colored PNG.
- **Verification**: `ls static/icons/`

## Phase 2: Frontend Integration

### Step 2.1: Update Layout with PWA Meta Tags
Add necessary tags to the `<head>` of the main layout.
- **File**: `api/templates/layout.html`
- **Action**: Add the following inside `<head>`:
  ```html
  <link rel="manifest" href="/manifest.json">
  <meta name="theme-color" content="#00A36C">
  <meta name="apple-mobile-web-app-capable" content="yes">
  <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
  <link rel="apple-touch-icon" href="/static/icons/icon-192.png">
  ```
- **Verification**: Refresh the page and inspect the `<head>` in DevTools.

### Step 2.2: Register Service Worker
Add the registration script to the bottom of the layout.
- **File**: `api/templates/layout.html`
- **Action**: Add before `</body>`:
  ```html
  <script>
    if ('serviceWorker' in navigator) {
      window.addEventListener('load', () => {
        navigator.serviceWorker.register('/sw.js')
          .then(reg => console.log('SW registered:', reg))
          .catch(err => console.error('SW registration failed:', err));
      });
    }
  </script>
  ```
- **Verification**: Check Console in browser DevTools for "SW registered".

### Step 2.3: Implement "Add to Home Screen" UI
Add a button to trigger the install prompt and the logic to handle the event.
- **File**: `api/templates/layout.html`
- **Action**:
  1. Add an "Install" button in the sidebar (hidden by default):
     ```html
     <button id="pwa-install-btn" class="hidden text-slate-400 hover:text-primary transition-colors" title="Install App">
         <span class="material-symbols-outlined">download</span>
     </button>
     ```
  2. Add the JS logic to handle `beforeinstallprompt`:
     ```javascript
     let deferredPrompt;
     const installBtn = document.getElementById('pwa-install-btn');

     window.addEventListener('beforeinstallprompt', (e) => {
       e.preventDefault();
       deferredPrompt = e;
       installBtn.classList.remove('hidden');
     });

     installBtn.addEventListener('click', async () => {
       if (!deferredPrompt) return;
       deferredPrompt.prompt();
       const { outcome } = await deferredPrompt.userChoice;
       if (outcome === 'accepted') {
         console.log('User accepted the A2HS prompt');
       }
       deferredPrompt = null;
       installBtn.classList.add('hidden');
     });
     ```
- **Verification**: Use Chrome DevTools (Application -> Manifest -> Installability) to verify.

---

# Critical Thinking & Edge Cases

- **HTTPS Requirement**: PWAs require HTTPS (except for localhost). Ensure the documentation mentions that production deployments must use TLS.
- **Service Worker Scope**: The SW is placed at the root (`/sw.js`) to ensure it has control over the entire application.
- **Cache Invalidation**: The `sw.js` should implement a strategy to avoid stale HTML. A "Network First" approach is recommended for the main pages.
- **Maskable Icons**: For better integration on Android, consider adding `"purpose": "maskable"` to one of the icons in `manifest.json`.

---

# Final Comprehensive Verification Plan

### Automated Checks
1. **Lighthouse Audit**: Run a Lighthouse report in Chrome and ensure the "PWA" section passes.
2. **Server Check**: Verify `/manifest.json` and `/sw.js` return 200 OK with correct `Content-Type`.

### Manual Walkthrough
1. **Desktop Chrome**:
   - Open the app on `localhost:3281`.
   - Look for the "Install" icon in the address bar or the custom install button in the sidebar.
   - Click it and verify the app installs as a standalone window.
2. **Mobile (Simulated or Real)**:
   - Access the app.
   - Verify the "Add to Home Screen" prompt can be triggered.
   - Verify the theme color (`#00A36C`) is applied to the browser's status bar.
3. **Offline Check**:
   - Turn off the server or use "Offline" mode in DevTools.
   - Verify that the app still loads a basic shell or the cached pages.
