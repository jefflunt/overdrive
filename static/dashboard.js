class DashboardManager {
    constructor(runningContainerId, recentContainerId, starredContainerId, allContainerId) {
        this.runningContainer = document.getElementById(runningContainerId);
        this.recentContainer = document.getElementById(recentContainerId);
        this.starredContainer = document.getElementById(starredContainerId);
        this.allContainer = document.getElementById(allContainerId);
        this.data = null; // { projects: { [projectName]: Todo[] }, running_jobs: Job[], recent_jobs: Job[] }
        this.collapsedAll = new Set();
        this.collapsedStarred = new Set();
        
        // All projects are collapsed by default. We'll track project collapses separately.
        this.collapsedProjectsAll = new Set();
        this.collapsedProjectsStarred = new Set();
        
        this.initializedNodesAll = new Set();
        this.initializedNodesStarred = new Set();
        this.recentJobsCollapsed = false;

        this.showCompletedStarred = true;
        this.showCompletedAll = true;
        
        this.init();
    }

    async init() {
        await this.load(true);
        this.setupEventListeners();

        // Refresh every 10 seconds to keep running jobs up to date
        setInterval(() => this.load(), 10000);

        // Update durations every second
        setInterval(() => this.updateDurations(), 1000);
    }

    updateDurations() {
        document.querySelectorAll('.job-duration').forEach(el => {
            const startedAt = el.dataset.startedAt;
            const completedAt = el.dataset.completedAt;
            const status = el.dataset.status;
            
            if (status === 'pending') {
                el.textContent = 'scheduled';
                return;
            }
            
            if (!startedAt) {
                el.textContent = '';
                return;
            }
            
            const isTerminal = ['done', 'crash', 'no-op', 'timeout', 'stopped', 'undone', 'cancelled'].includes(status);
            
            const start = new Date(startedAt);
            let end = new Date();
            
            if (isTerminal && completedAt) {
                end = new Date(completedAt);
            }
            
            const diff = Math.floor((end - start) / 1000);
            
            if (diff < 0) {
                el.textContent = '0s';
            } else if (diff < 60) {
                el.textContent = `${diff}s`;
            } else if (diff < 3600) {
                el.textContent = `${Math.floor(diff / 60)}m${diff % 60}s`;
            } else {
                el.textContent = `${Math.floor(diff / 3600)}h${Math.floor((diff % 3600) / 60)}m${diff % 60}s`;
            }
        });
    }

    async load(isInitial = false) {
        try {
            const res = await fetch('/api/dashboard/todos');
            if (!res.ok) throw new Error('Failed to load dashboard data');
            this.data = await res.json();
            
            // By default, collapse all project groups in the "All" section on initial load
            if (isInitial && this.data && this.data.projects) {
                Object.keys(this.data.projects).forEach(p => this.collapsedProjectsAll.add(p));
            }
            
            this.render();
        } catch (err) {
            console.error('Error loading dashboard:', err);
            this.starredContainer.innerHTML = '<div class="text-red-500">Failed to load</div>';
            this.allContainer.innerHTML = '<div class="text-red-500">Failed to load</div>';
        }
    }

    setupEventListeners() {
        document.addEventListener('click', async (e) => {
            // Handle Recent Jobs Collapse
            const recentJobsHeader = e.target.closest('#recent-jobs-header');
            if (recentJobsHeader) {
                this.recentJobsCollapsed = !this.recentJobsCollapsed;
                this.render();
                return;
            }

            // Handle Collapse/Expand Node
            const expandBtn = e.target.closest('.todo-expand');
            if (expandBtn) {
                const nodeId = expandBtn.dataset.id;
                const isStarredSection = expandBtn.closest('#starred-tree-container') !== null;
                const collapsedSet = isStarredSection ? this.collapsedStarred : this.collapsedAll;
                
                if (collapsedSet.has(nodeId)) {
                    collapsedSet.delete(nodeId);
                } else {
                    collapsedSet.add(nodeId);
                }
                this.render();
                return;
            }

            // Handle Collapse/Expand Project Group
            const projectGroupBtn = e.target.closest('.project-group-expand');
            if (projectGroupBtn) {
                const projectName = projectGroupBtn.dataset.project;
                const isStarredSection = projectGroupBtn.closest('#starred-tree-container') !== null;
                const collapsedSet = isStarredSection ? this.collapsedProjectsStarred : this.collapsedProjectsAll;
                
                if (collapsedSet.has(projectName)) {
                    collapsedSet.delete(projectName);
                } else {
                    collapsedSet.add(projectName);
                }
                this.render();
                return;
            }

            // Handle Star/Unstar
            const starBtn = e.target.closest('.todo-btn.star');
            if (starBtn) {
                const id = starBtn.dataset.id;
                const project = starBtn.dataset.project;
                const isStarred = starBtn.dataset.starred === 'true';
                await this.toggleStar(project, id, !isStarred);
                return;
            }

            // Handle Submit/Enqueue
            const submitBtn = e.target.closest('.todo-btn.submit');
            if (submitBtn) {
                const id = submitBtn.dataset.id;
                const project = submitBtn.dataset.project;
                await this.submitTodo(project, id);
                return;
            }

            // Handle Show Completed Toggle
            const showCompletedStarred = e.target.closest('#show-completed-starred');
            if (showCompletedStarred) {
                this.showCompletedStarred = showCompletedStarred.checked;
                this.render();
                return;
            }

            const showCompletedAll = e.target.closest('#show-completed-all');
            if (showCompletedAll) {
                this.showCompletedAll = showCompletedAll.checked;
                this.render();
                return;
            }

            // Handle Toggle Prompt
            const todoContent = e.target.closest('.todo-item-clickable');
            if (todoContent && !e.target.closest('.todo-btn') && !e.target.closest('.todo-expand') && !e.target.closest('a')) {
                const node = todoContent.closest('.todo-node');
                const prompt = node.querySelector('.todo-prompt');
                if (prompt) {
                    prompt.classList.toggle('hidden');
                }
                return;
            }
        });
    }

    async toggleStar(projectName, todoId, newStarredStatus) {
        try {
            const res = await fetch(`/projects/${projectName}/todos/${todoId}/star`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ starred: newStarredStatus })
            });
            if (!res.ok) throw new Error('Failed to update star');
            await this.load();
        } catch (err) {
            console.error(err);
            showToast('Failed to update star status', 'error');
        }
    }

    async submitTodo(projectName, todoId) {
        if (!confirm('Submit this task to the job queue?')) {
            return;
        }
        try {
            const res = await fetch(`/projects/${projectName}/todos/${todoId}/submit`, {
                method: 'POST'
            });
            if (!res.ok) throw new Error(await res.text());
            await this.load();
        } catch (err) {
            console.error(err);
            showToast(err.message, 'error');
        }
    }

    filterTodos(todos, showCompleted) {
        if (showCompleted) return todos;
        return todos
            .filter(todo => todo.status !== 'completed')
            .map(todo => ({
                ...todo,
                children: todo.children ? this.filterTodos(todo.children, showCompleted) : []
            }));
    }

    extractStarredRoots(todos, showCompleted) {
        let roots = [];
        if (!todos) return roots;
        for (const todo of todos) {
            if (todo.starred) {
                if (showCompleted || todo.status !== 'completed') {
                    // When we add a starred root, we should also filter its children if we're not showing completed
                    if (!showCompleted) {
                        const filteredTodo = {
                            ...todo,
                            children: todo.children ? this.filterTodos(todo.children, false) : []
                        };
                        roots.push(filteredTodo);
                    } else {
                        roots.push(todo);
                    }
                }
            } else if (todo.children) {
                roots = roots.concat(this.extractStarredRoots(todo.children, showCompleted));
            }
        }
        return roots;
    }

    flattenTodos(todos) {
        let flat = [];
        if (!todos) return flat;
        for (const todo of todos) {
            flat.push(todo);
            if (todo.children) {
                flat = flat.concat(this.flattenTodos(todo.children));
            }
        }
        return flat;
    }

    render() {
        if (!this.data) return;

        // Update Toggle Checkboxes
        const starredToggle = document.getElementById('show-completed-starred');
        if (starredToggle) starredToggle.checked = this.showCompletedStarred;
        const allToggle = document.getElementById('show-completed-all');
        if (allToggle) allToggle.checked = this.showCompletedAll;

        // Render Running Jobs Section
        if (this.data.running_jobs && this.data.running_jobs.length > 0) {
            this.runningContainer.innerHTML = '';
            this.data.running_jobs.forEach(job => {
                this.runningContainer.appendChild(this.createJobNode(job));
            });
            document.getElementById('running-jobs-container-wrapper').classList.remove('hidden');
        } else {
            document.getElementById('running-jobs-container-wrapper').classList.add('hidden');
        }

        // Render Recent Jobs Section
        if (this.data.recent_jobs && this.data.recent_jobs.length > 0) {
            const toggleIcon = document.getElementById('recent-jobs-toggle-icon');
            const recentJobsHeader = document.getElementById('recent-jobs-header');
            
            if (recentJobsHeader) {
                recentJobsHeader.setAttribute('aria-expanded', !this.recentJobsCollapsed);
            }

            if (toggleIcon) {
                if (this.recentJobsCollapsed) {
                    toggleIcon.classList.add('-rotate-90');
                } else {
                    toggleIcon.classList.remove('-rotate-90');
                }
            }

            if (this.recentJobsCollapsed) {
                this.recentContainer.classList.add('hidden');
            } else {
                this.recentContainer.classList.remove('hidden');
                this.recentContainer.innerHTML = '';
                this.data.recent_jobs.forEach(job => {
                    this.recentContainer.appendChild(this.createJobNode(job));
                });
            }
            document.getElementById('recent-jobs-container-wrapper').classList.remove('hidden');
        } else {
            document.getElementById('recent-jobs-container-wrapper').classList.add('hidden');
        }

        if (!this.data.projects) return;

        // Render Starred Section
        this.starredContainer.innerHTML = '';
        let hasStarred = false;
        
        for (const [projectName, todos] of Object.entries(this.data.projects)) {
            const starredRoots = this.extractStarredRoots(todos, this.showCompletedStarred);
            if (starredRoots.length > 0) {
                hasStarred = true;
                this.starredContainer.appendChild(this.createProjectGroup(projectName, starredRoots, true, false, true));
            }
        }

        if (!hasStarred) {
            this.starredContainer.innerHTML = `
                <div class="text-center p-6 border border-dashed border-slate-200 dark:border-border-dark rounded-lg">
                    <span class="material-symbols-outlined text-4xl text-slate-300 dark:text-slate-600 mb-2">star_border</span>
                    <p class="text-slate-500">No starred items yet.</p>
                </div>
            `;
        }

        // Render All Section
        this.allContainer.innerHTML = '';
        let hasProjects = false;

        for (const [projectName, todos] of Object.entries(this.data.projects)) {
            const filteredTodos = this.filterTodos(todos, this.showCompletedAll);
            if (filteredTodos && filteredTodos.length > 0) {
                hasProjects = true;
                // DO NOT flatten todos here anymore, we want the tree structure
                this.allContainer.appendChild(this.createProjectGroup(projectName, filteredTodos, false, false, true));
            }
        }

        if (!hasProjects) {
            this.allContainer.innerHTML = `
                <div class="text-center p-6 border border-dashed border-slate-200 dark:border-border-dark rounded-lg">
                    <p class="text-slate-500">No projects or todos found.</p>
                </div>
            `;
        }
    }

    createJobNode(job) {
        const node = document.createElement('div');
        node.className = 'flex items-center justify-between p-3 bg-white dark:bg-panel-dark border border-slate-200 dark:border-border-dark rounded-lg mb-2 hover:border-primary/50 transition-all cursor-pointer';
        node.onclick = () => {
            window.location.href = `/projects/${job.project}/jobs#job-${job.id}`;
        };
        
        let statusColor = 'slate-500';
        let animateClass = '';
        let statusDisplay = job.status;

        if (job.status === 'working') {
            statusColor = 'blue-500';
            animateClass = 'animate-pulse';
            statusDisplay = 'build';
        } else if (job.status === 'pending') {
            statusColor = 'slate-500';
            statusDisplay = 'todo';
        } else if (job.status === 'done') {
            statusColor = 'green-500';
            statusDisplay = 'done';
        } else if (job.status === 'crash') {
            statusColor = 'red-500';
            statusDisplay = 'crash';
        }
        
        let promptSummary = job.request.prompt;
        if (promptSummary.startsWith('/bdoc-engineer ')) {
            promptSummary = promptSummary.substring('/bdoc-engineer '.length);
        } else if (promptSummary.startsWith('/bdoc-quick ')) {
            promptSummary = 'Quick fix: ' + promptSummary.substring('/bdoc-quick '.length);
        } else if (promptSummary.startsWith('/bdoc-update')) {
            promptSummary = 'Updating documentation';
        }

        node.innerHTML = `
            <div class="flex items-center gap-4 flex-1 min-w-0">
                <div class="flex items-center gap-2 flex-shrink-0">
                    <div class="w-2 h-2 rounded-full bg-${statusColor} ${animateClass}"></div>
                    <span class="text-xs font-mono font-bold text-${statusColor} uppercase tracking-widest">${statusDisplay}</span>
                </div>
                <div class="flex flex-col min-w-0">
                    <div class="flex items-center gap-2">
                        <span class="text-[10px] font-mono font-bold text-primary uppercase tracking-wide">${job.project}</span>
                    </div>
                    <div class="text-sm text-slate-900 dark:text-slate-100 truncate font-medium">${promptSummary}</div>
                </div>
            </div>
            <div class="flex items-center gap-4 flex-shrink-0 ml-4">
                <div class="job-duration text-xs font-mono text-slate-500" data-started-at="${job.started_at || ''}" data-completed-at="${job.completed_at || ''}" data-status="${job.status}">${this.formatDuration(job)}</div>
            </div>
        `;
        return node;
    }

    formatDuration(job) {
        if (job.status === 'pending') return 'scheduled';
        if (!job.started_at) return '';
        
        const start = new Date(job.started_at);
        let end = new Date();
        
        const isTerminal = ['done', 'crash', 'no-op', 'timeout', 'stopped', 'undone', 'cancelled'].includes(job.status);
        if (isTerminal && job.completed_at) {
            end = new Date(job.completed_at);
        }
        
        const diff = Math.floor((end - start) / 1000);
        
        if (diff < 0) return '0s';
        if (diff < 60) return `${diff}s`;
        if (diff < 3600) return `${Math.floor(diff / 60)}m${diff % 60}s`;
        return `${Math.floor(diff / 3600)}h${Math.floor((diff % 3600) / 60)}m${diff % 60}s`;
    }

    createProjectGroup(projectName, todos, isStarredSection, isFlat = false, isCompact = false) {
        const group = document.createElement('div');
        group.className = isCompact ? 'mb-4 border border-slate-200 dark:border-border-dark rounded-lg bg-white dark:bg-panel-dark overflow-hidden' : 'mb-6 border border-slate-200 dark:border-border-dark rounded-lg bg-white dark:bg-panel-dark overflow-hidden';
        
        const header = document.createElement('div');
        header.className = isCompact ? 'bg-slate-50 dark:bg-slate-900/50 px-3 py-2 border-b border-slate-200 dark:border-border-dark flex items-center justify-between cursor-pointer project-group-expand hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors' : 'bg-slate-50 dark:bg-slate-900/50 px-4 py-3 border-b border-slate-200 dark:border-border-dark flex items-center justify-between cursor-pointer project-group-expand hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors';
        header.dataset.project = projectName;
        
        // In Starred section, we generally default to expanded. In All section, collapsed by default.
        const isCollapsed = isStarredSection ? this.collapsedProjectsStarred.has(projectName) : this.collapsedProjectsAll.has(projectName);

        header.innerHTML = `
            <div class="flex items-center gap-2">
                <span class="material-symbols-outlined text-sm text-slate-400 transition-transform ${isCollapsed ? '-rotate-90' : ''}">expand_more</span>
                <span class="material-symbols-outlined text-slate-400 ${isCompact ? 'text-lg' : ''}">folder</span>
                <span class="font-mono font-bold text-sm uppercase tracking-wide text-primary">${projectName}</span>
            </div>
            <div class="text-xs text-slate-500 font-mono">${todos.length} items</div>
        `;
        group.appendChild(header);

        if (!isCollapsed) {
            const list = document.createElement('div');
            list.className = isCompact ? 'todo-tree p-2' : 'todo-tree p-4';
            todos.forEach(todo => {
                // Determine if we should forcefully collapse all nodes recursively to respect "fully collapsed by default"
                // Actually, the user said "fully collapsed by default, for this dashboard view".
                // I will make sure all nodes are added to 'collapsedSet' by default if we haven't seen them before.
                list.appendChild(this.createTodoNode(todo, projectName, isStarredSection, isFlat, isCompact));
            });
            group.appendChild(list);
        }

        return group;
    }

    createTodoNode(todo, projectName, isStarredSection, isFlat = false, isCompact = false) {
        // Automatically collapse by default if it's our first time rendering this node and it has children
        if (!isFlat && todo.children && todo.children.length > 0) {
            const collapsedSet = isStarredSection ? this.collapsedStarred : this.collapsedAll;
            const initializedSet = isStarredSection ? this.initializedNodesStarred : this.initializedNodesAll;
            
            if (!initializedSet.has(todo.id)) {
                collapsedSet.add(todo.id);
                initializedSet.add(todo.id);
            }
        }

        const node = document.createElement('div');
        node.className = 'todo-node';
        
        const isCollapsed = isStarredSection ? this.collapsedStarred.has(todo.id) : this.collapsedAll.has(todo.id);
        const hasChildren = !isFlat && todo.children && todo.children.length > 0;
        const isLeaf = !todo.children || todo.children.length === 0;

        const content = document.createElement('div');
        content.className = isCompact 
            ? 'todo-content todo-item-clickable cursor-pointer flex items-center gap-2 py-1.5 px-3 bg-white dark:bg-panel-dark border border-slate-200 dark:border-border-dark rounded-md mb-1 relative transition-all hover:border-primary/50'
            : 'todo-content todo-item-clickable cursor-pointer flex items-start gap-3 p-3 bg-white dark:bg-panel-dark border border-slate-200 dark:border-border-dark rounded-lg mb-2 relative transition-all hover:border-primary/50';
        
        let statusClass = 'status-draft';
        if (todo.status === 'submitted') statusClass = 'status-submitted';
        if (todo.status === 'completed') statusClass = 'status-completed';
        if (todo.status === 'crashed') statusClass = 'status-crashed';

        const starIcon = todo.starred ? 'star' : 'star_border';
        const starColor = todo.starred ? 'text-yellow-500' : 'text-slate-300 dark:text-slate-600 hover:text-yellow-500';

        if (isCompact) {
            content.innerHTML = `
                ${hasChildren ? `
                    <div class="todo-expand ${isCollapsed ? 'collapsed' : ''}" data-id="${todo.id}">
                        <span class="material-symbols-outlined text-sm">expand_more</span>
                    </div>
                ` : `
                    <div class="w-5 flex-shrink-0"></div>
                `}
                
                <div class="flex-1 min-w-0 flex items-center gap-2 overflow-hidden">
                    <span class="font-bold text-sm text-slate-900 dark:text-slate-100 truncate">${todo.title}</span>
                    <span class="todo-status ${statusClass} text-[9px] px-1.5 py-0 whitespace-nowrap flex-shrink-0" style="line-height: 1.2;">${todo.status}</span>
                    ${todo.jobId ? `<a href="/projects/${projectName}/jobs/logs/${todo.jobId}" class="text-[10px] font-mono text-primary hover:underline flex-shrink-0" title="View Job Logs">#${todo.jobId.substring(0,8)}</a>` : ''}
                    ${todo.description ? `<span class="text-[11px] text-slate-400 truncate flex-1 hidden md:inline">— ${todo.description}</span>` : ''}
                </div>

                <div class="flex items-center gap-1 transition-opacity">
                    <button class="todo-btn star p-1 rounded hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors" data-id="${todo.id}" data-project="${projectName}" data-starred="${todo.starred ? 'true' : 'false'}" title="${todo.starred ? 'Unstar' : 'Star'}">
                        <span class="material-symbols-outlined text-[18px] ${starColor}">${starIcon}</span>
                    </button>
                    ${(isLeaf && todo.status === 'draft') ? `
                    <button class="todo-btn submit p-1 rounded hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors text-primary" data-id="${todo.id}" data-project="${projectName}" title="Enqueue Job">
                        <span class="material-symbols-outlined text-[18px]">play_circle</span>
                    </button>
                    ` : ''}
                    <a href="/projects/${projectName}/todos" class="todo-btn p-1 rounded hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors text-slate-400" title="Go to Project Todos">
                        <span class="material-symbols-outlined text-[18px]">open_in_new</span>
                    </a>
                </div>
            `;
        } else {
            content.innerHTML = `
                ${hasChildren ? `
                    <div class="todo-expand ${isCollapsed ? 'collapsed' : ''} mt-1" data-id="${todo.id}">
                        <span class="material-symbols-outlined text-sm">expand_more</span>
                    </div>
                ` : `
                    <div class="w-5 mt-1 flex-shrink-0"></div>
                `}
                
                <div class="flex-1 min-w-0 flex flex-col gap-1">
                    <div class="flex items-center gap-2 flex-wrap">
                        <span class="font-bold text-sm text-slate-900 dark:text-slate-100">${todo.title}</span>
                        <span class="todo-status ${statusClass}">${todo.status}</span>
                        ${todo.jobId ? `<a href="/projects/${projectName}/jobs/logs/${todo.jobId}" class="text-[10px] font-mono text-primary hover:underline ml-2" title="View Job Logs">#${todo.jobId.substring(0,8)}</a>` : ''}
                    </div>
                    ${todo.description ? `<p class="text-xs text-slate-500 line-clamp-2" title="${todo.description.replace(/"/g, '&quot;')}">${todo.description}</p>` : ''}
                </div>

                <div class="flex items-center gap-1 transition-opacity">
                    <button class="todo-btn star p-1 rounded hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors" data-id="${todo.id}" data-project="${projectName}" data-starred="${todo.starred ? 'true' : 'false'}" title="${todo.starred ? 'Unstar' : 'Star'}">
                        <span class="material-symbols-outlined text-[20px] ${starColor}">${starIcon}</span>
                    </button>
                    ${(isLeaf && todo.status === 'draft') ? `
                    <button class="todo-btn submit p-1 rounded hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors text-primary" data-id="${todo.id}" data-project="${projectName}" title="Enqueue Job">
                        <span class="material-symbols-outlined text-[20px]">play_circle</span>
                    </button>
                    ` : ''}
                    <a href="/projects/${projectName}/todos" class="todo-btn p-1 rounded hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors text-slate-400" title="Go to Project Todos">
                        <span class="material-symbols-outlined text-[20px]">open_in_new</span>
                    </a>
                </div>
            `;
        }
 
        node.appendChild(content);

        const promptDiv = document.createElement('div');
        promptDiv.className = 'todo-prompt hidden mb-2 ml-6 p-3 bg-slate-50 dark:bg-slate-900/50 rounded border border-slate-200 dark:border-border-dark font-mono text-[11px] whitespace-pre-wrap text-slate-600 dark:text-slate-400';
        promptDiv.textContent = `/bdoc-engineer # title\n\n${todo.title}\n\n## details\n\n${todo.description || ''}`;
        node.appendChild(promptDiv);

        if (hasChildren) {
            const childrenContainer = document.createElement('div');
            childrenContainer.className = isCompact
                ? `todo-children border-l border-slate-100 dark:border-slate-800 ml-2 pl-3 ${isCollapsed ? 'hidden' : ''}`
                : `todo-children border-l-2 border-slate-100 dark:border-slate-800 ml-4 pl-4 ${isCollapsed ? 'hidden' : ''}`;
            todo.children.forEach(child => {
                childrenContainer.appendChild(this.createTodoNode(child, projectName, isStarredSection, isFlat, isCompact));
            });
            node.appendChild(childrenContainer);
        }

        return node;
    }
}
