package styles

const globalCSS = `
/* === CSS Variables === */
:root {
    --bg-primary: #ffffff;
    --bg-sidebar: #f7f6f3;
    --bg-hover: #efefef;
    --bg-active: #e8e7e4;
    --text-primary: #37352f;
    --text-secondary: #9b9a97;
    --text-placeholder: #c4c4c0;
    --accent: #2eaadc;
    --border: #e9e9e7;
    --font-body: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
    --font-mono: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
    --sidebar-width: 240px;
    --editor-max-width: 720px;
    --transition-speed: 200ms;
}

/* === Reset === */
*, *::before, *::after {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
}

html, body {
    height: 100%;
    font-family: var(--font-body);
    font-size: 16px;
    line-height: 1.6;
    color: var(--text-primary);
    background: var(--bg-primary);
    -webkit-font-smoothing: antialiased;
}

/* === App Layout === */
.app-layout {
    display: flex;
    height: 100vh;
    overflow: hidden;
}

/* === Sidebar === */
.sidebar {
    width: var(--sidebar-width);
    min-width: var(--sidebar-width);
    height: 100vh;
    background: var(--bg-sidebar);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    transition: transform var(--transition-speed) ease;
}

.sidebar-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 14px;
    font-size: 14px;
    font-weight: 600;
    color: var(--text-secondary);
    min-height: 44px;
}

.sidebar-header-title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.new-page-btn {
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-secondary);
    font-size: 18px;
    padding: 4px 8px;
    border-radius: 4px;
    min-width: 44px;
    min-height: 44px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background var(--transition-speed);
}

.new-page-btn:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
}

.search-bar {
    padding: 0 14px 8px;
}

.search-input {
    width: 100%;
    padding: 6px 10px;
    border: 1px solid var(--border);
    border-radius: 4px;
    background: var(--bg-primary);
    font-size: 13px;
    font-family: var(--font-body);
    color: var(--text-primary);
    outline: none;
    min-height: 44px;
}

.search-input:focus {
    border-color: var(--accent);
}

.search-input::placeholder {
    color: var(--text-placeholder);
}

/* === Page Tree === */
.page-tree {
    flex: 1;
    overflow-y: auto;
    padding: 4px 0;
}

.page-tree-item {
    display: flex;
    align-items: center;
    padding: 4px 14px;
    cursor: pointer;
    color: var(--text-primary);
    font-size: 14px;
    border-radius: 4px;
    margin: 0 4px;
    transition: background var(--transition-speed);
    min-height: 44px;
    user-select: none;
}

.page-tree-item:hover {
    background: var(--bg-hover);
}

.page-tree-item-active {
    background: var(--bg-active);
}

.page-tree-item-icon {
    margin-right: 6px;
    font-size: 16px;
    flex-shrink: 0;
}

.page-tree-item-title {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
}

/* === Editor Area === */
.editor-area {
    flex: 1;
    overflow-y: auto;
    display: flex;
    justify-content: center;
    padding: 24px 48px;
}

.editor-page {
    width: 100%;
    max-width: var(--editor-max-width);
}

/* === Page Header === */
.page-header {
    margin-bottom: 24px;
}

.page-header-icon {
    font-size: 48px;
    cursor: pointer;
    margin-bottom: 8px;
    display: inline-block;
    padding: 4px;
    border-radius: 4px;
    transition: background var(--transition-speed);
}

.page-header-icon:hover {
    background: var(--bg-hover);
}

.page-header-title {
    font-size: 40px;
    font-weight: 700;
    line-height: 1.2;
    border: none;
    outline: none;
    width: 100%;
    font-family: var(--font-body);
    color: var(--text-primary);
    background: transparent;
    padding: 0;
}

.page-header-title::placeholder {
    color: var(--text-placeholder);
}

/* === Blocks === */
.block-list {
    min-height: 200px;
}

.block {
    padding: 3px 2px;
    border-radius: 3px;
    transition: background var(--transition-speed);
    position: relative;
}

.block:hover {
    background: var(--bg-hover);
}

.block:focus-within {
    background: transparent;
    box-shadow: -2px 0 0 0 var(--accent);
}

.block-content {
    outline: none;
    min-height: 1em;
    word-break: break-word;
}

.block-content:empty::before {
    content: attr(data-placeholder);
    color: var(--text-placeholder);
    pointer-events: none;
}

/* === Text Block === */
.block-text .block-content {
    font-size: 16px;
    line-height: 1.6;
}

/* === Heading Blocks === */
.block-h1 .block-content {
    font-size: 30px;
    font-weight: 700;
    line-height: 1.3;
    margin-top: 24px;
    margin-bottom: 4px;
}

.block-h2 .block-content {
    font-size: 24px;
    font-weight: 600;
    line-height: 1.3;
    margin-top: 20px;
    margin-bottom: 2px;
}

.block-h3 .block-content {
    font-size: 20px;
    font-weight: 600;
    line-height: 1.3;
    margin-top: 16px;
    margin-bottom: 1px;
}

/* === List Blocks === */
.block-bullet {
    display: flex;
    align-items: flex-start;
}

.block-bullet .block-bullet-marker {
    flex-shrink: 0;
    width: 24px;
    text-align: center;
    color: var(--text-primary);
    line-height: 1.6;
    font-size: 16px;
}

.block-numbered {
    display: flex;
    align-items: flex-start;
}

.block-numbered .block-number {
    flex-shrink: 0;
    width: 24px;
    text-align: center;
    color: var(--text-primary);
    line-height: 1.6;
    font-size: 16px;
}

/* === Toggle Block === */
.block-toggle {
    display: flex;
    align-items: flex-start;
}

.block-toggle .toggle-indicator {
    flex-shrink: 0;
    width: 24px;
    text-align: center;
    cursor: pointer;
    color: var(--text-secondary);
    line-height: 1.6;
    font-size: 14px;
    transition: transform var(--transition-speed);
    user-select: none;
    min-width: 44px;
    min-height: 44px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.block-toggle .toggle-indicator.expanded {
    transform: rotate(90deg);
}

.block-toggle .toggle-children {
    padding-left: 24px;
    border-left: 1px solid var(--border);
    margin-left: 12px;
    margin-top: 4px;
}

/* === Code Block === */
.block-code {
    background: #f7f6f3;
    border-radius: 4px;
    padding: 12px 16px;
    margin: 4px 0;
    font-family: var(--font-mono);
    font-size: 14px;
    line-height: 1.5;
    overflow-x: auto;
    tab-size: 2;
}

.block-code .code-language {
    font-size: 12px;
    color: var(--text-secondary);
    margin-bottom: 8px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.block-code textarea,
.block-code .block-content {
    font-family: var(--font-mono);
    font-size: 14px;
    line-height: 1.5;
    color: var(--text-primary);
    background: transparent;
    border: none;
    outline: none;
    width: 100%;
    resize: none;
    white-space: pre;
}

/* === Divider Block === */
.block-divider {
    padding: 12px 0;
}

.block-divider hr {
    border: none;
    border-top: 1px solid var(--border);
}

/* === Slash Command Menu === */
.slash-menu {
    position: absolute;
    z-index: 100;
    background: var(--bg-primary);
    border: 1px solid var(--border);
    border-radius: 6px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    padding: 6px 0;
    min-width: 240px;
    max-height: 320px;
    overflow-y: auto;
}

.slash-menu-item {
    display: flex;
    align-items: center;
    padding: 8px 14px;
    cursor: pointer;
    font-size: 14px;
    color: var(--text-primary);
    transition: background var(--transition-speed);
    min-height: 44px;
}

.slash-menu-item:hover,
.slash-menu-item-active {
    background: var(--bg-hover);
}

.slash-menu-item-icon {
    margin-right: 10px;
    font-size: 18px;
    width: 24px;
    text-align: center;
    color: var(--text-secondary);
}

.slash-menu-item-label {
    font-weight: 500;
}

.slash-menu-item-desc {
    font-size: 12px;
    color: var(--text-secondary);
    margin-left: auto;
    padding-left: 12px;
}

/* === Emoji Picker === */
.emoji-picker {
    position: absolute;
    z-index: 100;
    background: var(--bg-primary);
    border: 1px solid var(--border);
    border-radius: 6px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    padding: 8px;
    display: grid;
    grid-template-columns: repeat(8, 1fr);
    gap: 4px;
    max-width: 280px;
}

.emoji-picker-item {
    font-size: 20px;
    padding: 4px;
    cursor: pointer;
    border-radius: 4px;
    text-align: center;
    transition: background var(--transition-speed);
    min-width: 36px;
    min-height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.emoji-picker-item:hover {
    background: var(--bg-hover);
}

/* === Mobile Sidebar === */
.hamburger-btn {
    display: none;
    position: fixed;
    top: 12px;
    left: 12px;
    z-index: 200;
    background: var(--bg-primary);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 8px 12px;
    cursor: pointer;
    font-size: 18px;
    color: var(--text-primary);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    min-width: 44px;
    min-height: 44px;
    align-items: center;
    justify-content: center;
}

.sidebar-overlay {
    display: none;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.3);
    z-index: 49;
}

/* === Block Placeholder === */
.block-list-empty {
    padding: 8px 2px;
    color: var(--text-placeholder);
    font-size: 16px;
    cursor: text;
}

/* === Mobile Responsive === */
@media (max-width: 768px) {
    .sidebar {
        position: fixed;
        top: 0;
        left: 0;
        z-index: 50;
        transform: translateX(-100%);
        height: 100vh;
        box-shadow: 4px 0 12px rgba(0, 0, 0, 0.1);
    }

    .sidebar.sidebar-open {
        transform: translateX(0);
    }

    .sidebar-overlay.sidebar-overlay-visible {
        display: block;
    }

    .hamburger-btn {
        display: flex;
    }

    .editor-area {
        padding: 16px;
        padding-top: 56px;
    }

    .page-header-title {
        font-size: 28px;
    }

    .page-header-icon {
        font-size: 36px;
    }

    .slash-menu {
        position: fixed;
        bottom: 0;
        left: 0;
        right: 0;
        top: auto;
        border-radius: 12px 12px 0 0;
        max-height: 50vh;
        min-width: 100%;
    }
}
`
