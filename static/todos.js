
class TodoTree {
    constructor(containerId, projectId) {
        this.container = document.getElementById(containerId);
        this.projectId = projectId;
        this.todos = [];
        this.collapsed = new Set();
        this.project = null;
        this.setupEventListeners();
        this.init();
    }

    async init() {
        await this.loadProject();
        await this.load();
    }

    async loadProject() {
        try {
            const res = await fetch(`/projects/${this.projectId}/config`);
            if (!res.ok) throw new Error('Failed to load project config');
            this.project = await res.json();
        } catch (err) {
            console.error('Error loading project config:', err);
        }
    }

    debounce(func, wait) {
        let timeout;
        return (...args) => {
            const later = () => {
                clearTimeout(timeout);
                func.apply(this, args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    setupEventListeners() {
        this.container.addEventListener('mousedown', (e) => {
            const btn = e.target.closest('button, a');
            if (!btn) return;

            const action = btn.dataset.action;
            if (action === 'toggle-expand' || action === 'add-child' || action === 'delete' || action === 'submit' || action === 'add-root') {
                // Prevent focus theft for these actions to avoid premature onblur
                e.preventDefault();
            }
        });

        this.container.addEventListener('click', (e) => {
            const btn = e.target.closest('button, a');
            if (!btn) return;

            const action = btn.dataset.action;
            const id = btn.dataset.id;

            if (!action) return;

            e.stopPropagation();

            switch (action) {
                case 'toggle-expand':
                    this.toggleExpand(id);
                    break;
                case 'add-root':
                    this.addRootTodo();
                    break;
                case 'delete':
                    this.deleteTodo(id);
                    break;
                case 'add-child':
                    this.addChild(id);
                    break;
                case 'submit':
                    this.submitTodo(id);
                    break;
            }
        });

        this.container.addEventListener('focusout', (e) => {
            if (e.target.hasAttribute('contenteditable')) {
                const node = e.target.closest('.todo-node');
                if (!node) return;
                const id = node.id.replace('todo-', '');
                const field = e.target.classList.contains('todo-title') ? 'title' : 'description';
                this.updateTodo(id, field, e.target.innerText);
            }
        });

        this.container.addEventListener('keydown', (e) => {
            if (e.target.hasAttribute('contenteditable') && e.key === 'Enter') {
                if (e.target.classList.contains('todo-title')) {
                    e.preventDefault();
                    e.target.blur();
                }
            }
        });

        this.container.addEventListener('input', this.debounce((e) => {
            if (e.target.hasAttribute('contenteditable')) {
                const node = e.target.closest('.todo-node');
                if (!node) return;
                const id = node.id.replace('todo-', '');
                const field = e.target.classList.contains('todo-title') ? 'title' : 'description';
                this.updateTodo(id, field, e.target.innerText);
            }
        }, 1000));
    }

    async load(skipRender = false) {
        try {
            const res = await fetch(`/projects/${this.projectId}/todos`, {
                headers: { 'Accept': 'application/json' }
            });
            if (!res.ok) throw new Error('Failed to load todos');
            this.todos = await res.json();
            if (!skipRender) {
                this.render();
            }
        } catch (err) {
            console.error(err);
            if (!skipRender) {
                this.container.innerHTML = `<div class="text-red-500">Error loading todos: ${err.message}</div>`;
            }
        }
    }

    render() {
        this.container.innerHTML = '';
        const rootList = document.createElement('div');
        rootList.className = 'todo-tree';
        
        const isJira = this.project && this.project.todo_provider === 'jira';
        const isGitHub = this.project && this.project.todo_provider === 'github';
        const isRemote = isJira || isGitHub;

        // Add Header
        const header = document.createElement('div');
        header.className = 'flex items-center justify-between mb-4 md:mb-8';
        header.innerHTML = `
            <div class="flex flex-col">
                <h2 class="text-2xl font-bold font-display text-slate-900 dark:text-slate-100">
                    ${isJira ? 'Jira Issues' : (isGitHub ? 'GitHub Issues' : 'Project Todos')}
                </h2>
                ${isJira ? `<p class="text-xs text-slate-500 mt-1">Jira Project: <span class="font-mono font-bold">${this.project.jira.project_key}</span></p>` : ''}
                ${isGitHub ? `<p class="text-xs text-slate-500 mt-1">GitHub Repo: <span class="font-mono font-bold">${this.project.github.repo}</span></p>` : ''}
            </div>
            ${!isRemote ? `
            <button class="bg-primary dark:bg-primary-dark hover:bg-primary/90 text-black px-4 py-2 rounded text-sm font-bold shadow-lg transition-transform hover:scale-105 flex items-center gap-2" data-action="add-root">
                <span class="material-symbols-outlined text-lg">add</span> New Feature
            </button>
            ` : ''}
        `;
        this.container.appendChild(header);

        if (!this.todos || this.todos.length === 0) {
            const empty = document.createElement('div');
            empty.className = 'text-center p-6 md:p-12 border-2 border-dashed border-slate-200 dark:border-border-dark rounded-lg';
            empty.innerHTML = `
                <div class="text-slate-400 mb-4 text-4xl">${isRemote ? '🔍' : '📝'}</div>
                <h3 class="text-lg font-medium text-slate-900 dark:text-slate-200 mb-2">No ${isRemote ? 'issues' : 'todos'} found</h3>
                <p class="text-slate-500 mb-6">${isJira ? 'Check your Jira project and configuration.' : (isGitHub ? 'Check your GitHub repository and configuration.' : 'Create a hierarchical chat for your project.')}</p>
                ${!isRemote ? '<button class="text-primary hover:underline font-medium" data-action="add-root">Create first item</button>' : ''}
            `;
            this.container.appendChild(empty);
            return;
        }

        this.todos.forEach(todo => {
            rootList.appendChild(this.createTodoNode(todo, true));
        });

        this.container.appendChild(rootList);
    }

    createTodoNode(todo, isRoot = false) {
        const node = document.createElement('div');
        node.className = 'todo-node';
        node.id = `todo-${todo.id}`;

        const isJira = this.project && this.project.todo_provider === 'jira';
        const isGitHub = this.project && this.project.todo_provider === 'github';
        const isRemote = isJira || isGitHub;
        const isLeaf = !todo.children || todo.children.length === 0;
        const hasChildren = todo.children && todo.children.length > 0;
        const isExpanded = !this.collapsed.has(todo.id);

        const content = document.createElement('div');
        content.className = `todo-content group ${todo.status}`;
        
        let statusBadge = '';
        let statusClass = '';
        
        if (isRemote) {
            statusBadge = `${isJira ? 'JIRA' : 'GH'}: ${todo.status.toUpperCase()}`;
            statusClass = 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400';
            
            const lowerStatus = todo.status.toLowerCase();
            if (lowerStatus.includes('progress') || lowerStatus.includes('doing') || lowerStatus.includes('open')) {
                statusClass = 'bg-blue-50 text-blue-600 dark:bg-blue-900/20 dark:text-blue-400';
            } else if (lowerStatus.includes('done') || lowerStatus.includes('complete') || lowerStatus.includes('resolved') || lowerStatus.includes('closed')) {
                statusClass = 'bg-green-50 text-green-600 dark:bg-green-900/20 dark:text-green-400';
            } else if (lowerStatus.includes('fail') || lowerStatus.includes('error')) {
                statusClass = 'bg-red-50 text-red-600 dark:bg-red-900/20 dark:text-red-400';
            }
        } else {
            if (todo.status === 'submitted') {
                statusBadge = '🔒 SUBMITTED';
                statusClass = 'bg-blue-50 text-blue-600 dark:bg-blue-900/20 dark:text-blue-400';
            } else if (todo.status === 'completed') {
                statusBadge = '✅ COMPLETED';
                statusClass = 'bg-green-50 text-green-600 dark:bg-green-900/20 dark:text-green-400';
            } else if (todo.status === 'crashed') {
                statusBadge = '❌ CRASHED';
                statusClass = 'bg-red-50 text-red-600 dark:bg-red-900/20 dark:text-red-400';
            }
        }

        // Expand/Collapse Button
        const expandBtnHtml = `
            <button class="todo-expand ${hasChildren ? '' : 'invisible'} ${isExpanded ? '' : 'collapsed'}" 
                    data-action="toggle-expand" data-id="${todo.id}">
                <span class="material-symbols-outlined text-lg">expand_more</span>
            </button>
        `;

        // Actions
        let actions = '';

        if (!isRemote) {
            // Always include delete icon for local todos
            actions += `
                <button class="p-1 hover:bg-slate-200 dark:hover:bg-slate-700 rounded text-slate-400 hover:text-red-500 transition-colors" data-action="delete" data-id="${todo.id}" title="Delete/Discard">
                    <span class="material-symbols-outlined text-lg">delete</span>
                </button>
            `;

            if (todo.status === 'draft' || todo.status === 'crashed') {
                const isNew = !todo.title || todo.title === 'New Feature' || todo.title === 'New Subtask';
                
                if (!isNew) {
                    actions += `
                        <button class="p-1 hover:bg-slate-200 dark:hover:bg-slate-700 rounded text-slate-400 hover:text-primary transition-colors" data-action="add-child" data-id="${todo.id}" title="Add Subtask">
                            <span class="material-symbols-outlined text-lg">add_circle</span>
                        </button>
                    `;
                    
                    if (isLeaf) {
                        actions += `
                            <button class="p-1 hover:bg-slate-200 dark:hover:bg-slate-700 rounded text-slate-400 hover:text-blue-500 transition-colors" data-action="submit" data-id="${todo.id}" title="Submit to Job Queue">
                                <span class="material-symbols-outlined text-lg">rocket_launch</span>
                            </button>
                        `;
                    }
                }
            } else if (todo.status === 'submitted' && todo.jobId) {
                actions += `
                    <a href="/projects/${this.projectId}/jobs?q=${todo.jobId}" class="p-1 hover:bg-slate-200 dark:hover:bg-slate-700 rounded text-slate-400 hover:text-blue-500 transition-colors" title="View Job">
                        <span class="material-symbols-outlined text-lg">visibility</span>
                    </a>
                `;
            }
        } else {
            // Remote Actions
            if (isLeaf) {
                const lowerStatus = todo.status.toLowerCase();
                const canPickup = !lowerStatus.includes('done') && !lowerStatus.includes('complete') && !lowerStatus.includes('resolved') && !lowerStatus.includes('progress') && !lowerStatus.includes('closed');
                
                if (canPickup) {
                    actions += `
                        <button class="flex items-center gap-1 px-2 py-1 bg-primary dark:bg-primary-dark text-black rounded text-[10px] font-bold shadow transition-transform hover:scale-105" data-action="submit" data-id="${todo.id}">
                            <span class="material-symbols-outlined text-xs">play_arrow</span> Pick up
                        </button>
                    `;
                }
            }
            
            // Link to Remote
            const remoteUrl = isJira ? `${this.project.jira.instance}/browse/${todo.id}` : `https://github.com/${this.project.github.repo}/issues/${todo.id}`;
            actions += `
                <a href="${remoteUrl}" target="_blank" class="p-1 hover:bg-slate-200 dark:hover:bg-slate-700 rounded text-slate-400 hover:text-blue-500 transition-colors" title="View in ${isJira ? 'Jira' : 'GitHub'}">
                    <span class="material-symbols-outlined text-lg">open_in_new</span>
                </a>
            `;
        }

        const isEditable = !isRemote && (todo.status === 'draft' || todo.status === 'crashed');
        const placeholder = isRoot ? 'New Feature' : 'New Subtask';

        content.innerHTML = `
            <div class="flex items-start gap-3 w-full">
                ${expandBtnHtml}
                <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2 mb-1">
                        ${statusBadge ? `<span class="whitespace-nowrap text-[10px] font-bold px-2 py-0.5 rounded ${statusClass} tracking-wider">${statusBadge}</span>` : ''}
                        ${isRemote ? `<span class="whitespace-nowrap text-[10px] font-mono text-slate-400 bg-slate-100 dark:bg-slate-800 px-1 rounded">${todo.id}</span>` : ''}
                        <div class="todo-title flex-1 text-sm font-semibold text-slate-900 dark:text-slate-100 p-1 rounded ${isEditable ? 'hover:bg-slate-100 dark:hover:bg-slate-900 cursor-text' : ''} transition-colors" 
                            contenteditable="${isEditable}" data-placeholder="${placeholder}">
                            ${todo.title || ''}
                        </div>
                        <div class="flex items-center gap-1 transition-opacity">
                            ${actions}
                        </div>
                    </div>
                    ${(todo.description || isEditable) ? `
                    <div class="todo-description text-xs text-slate-500 dark:text-slate-400 p-1 rounded ${isEditable ? 'hover:bg-slate-100 dark:hover:bg-slate-900 cursor-text' : ''} mt-1" 
                        contenteditable="${isEditable}" data-placeholder="Add description...">
                        ${todo.description || ''}
                    </div>` : ''}
                </div>
            </div>
        `;

        node.appendChild(content);

        // Children Container
        if (hasChildren) {
            const childrenContainer = document.createElement('div');
            childrenContainer.className = `todo-children border-l border-slate-200 dark:border-slate-800 ${isExpanded ? '' : 'hidden'}`;
            
            todo.children.forEach(child => {
                childrenContainer.appendChild(this.createTodoNode(child, false));
            });
            node.appendChild(childrenContainer);
        }

        return node;
    }

    toggleExpand(id) {
        const node = document.getElementById(`todo-${id}`);
        if (!node) return;

        const expandBtn = node.querySelector('[data-action="toggle-expand"]');
        const childrenContainer = node.querySelector('.todo-children');

        if (!this.collapsed.has(id)) {
            this.collapsed.add(id);
            if (expandBtn) expandBtn.classList.add('collapsed');
            if (childrenContainer) childrenContainer.classList.add('hidden');
        } else {
            this.collapsed.delete(id);
            if (expandBtn) expandBtn.classList.remove('collapsed');
            if (childrenContainer) childrenContainer.classList.remove('hidden');
        }
    }

    async addRootTodo() {
        await this.createTodo({
            title: '',
            description: '',
            projectId: this.projectId
        });
    }

    async addChild(parentId) {
        this.collapsed.delete(parentId); // Auto expand parent
        await this.createTodo({
            title: '',
            description: '',
            projectId: this.projectId,
            parentId: parentId
        });
    }

    async createTodo(data) {
        try {
            const res = await fetch(`/projects/${this.projectId}/todos`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
            if (!res.ok) throw new Error('Failed to create todo');
            this.load();
        } catch (err) {
            showToast(err.message, 'error');
        }
    }

    async updateTodo(id, field, value) {
        // Find todo to check if value changed
        const find = (list) => {
            for (let t of list) {
                if (t.id === id) return t;
                if (t.children) {
                    const found = find(t.children);
                    if (found) return found;
                }
            }
            return null;
        };
        const todo = find(this.todos);
        if (!todo) return;
        if (todo[field] === value) return; // No change

        try {
            // Send both title and description to avoid backend wiping the other field
            const update = {
                title: field === 'title' ? value : todo.title,
                description: field === 'description' ? value : todo.description
            };
            const res = await fetch(`/projects/${this.projectId}/todos/${id}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(update)
            });
            if (!res.ok) throw new Error('Failed to update todo');
            
            // Update local state and sync from server without re-rendering
            todo[field] = value;
            await this.load(true); 
        } catch (err) {
            console.error(err);
        }
    }

    async deleteTodo(id) {
        if (!confirm('Are you sure you want to delete this todo and all its children?')) return;
        try {
            const res = await fetch(`/projects/${this.projectId}/todos/${id}`, {
                method: 'DELETE'
            });
            if (!res.ok) throw new Error('Failed to delete todo');
            this.collapsed.delete(id);
            this.load();
        } catch (err) {
            showToast(err.message, 'error');
        }
    }

    async submitTodo(id) {
         const isJira = this.project && this.project.todo_provider === 'jira';
         const isGitHub = this.project && this.project.todo_provider === 'github';
         const isRemote = isJira || isGitHub;
         const msg = isRemote ? `Pick up issue ${id}?` : 'Submit this task to the job queue?';
         if (!confirm(msg)) return;
         try {
            const res = await fetch(`/projects/${this.projectId}/todos/${id}/submit`, {
                method: 'POST'
            });
            if (!res.ok) {
                const text = await res.text();
                throw new Error(text);
            }
            this.load();
        } catch (err) {
            showToast(err.message, 'error');
        }
    }
}
