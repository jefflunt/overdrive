import sys

file_path = 'api/templates/home.html'
with open(file_path, 'r') as f:
    content = f.read()

# Update "How it Works" section
old_grid = '''        <div class="grid md:grid-cols-3 gap-12">
            <div class="text-center space-y-4">
                <div class="text-4xl font-mono font-bold text-primary dark:text-primary-dark opacity-30">01</div>
                <h4 class="text-lg font-bold dark:text-white">Connect Repo</h4>
                <p class="text-sm text-slate-500 dark:text-slate-400">Provide your Git SSH URL and optional credentials to link your project.</p>
            </div>
            <div class="text-center space-y-4">
                <div class="text-4xl font-mono font-bold text-primary dark:text-primary-dark opacity-30">02</div>
                <h4 class="text-lg font-bold dark:text-white">Configure Build</h4>
                <p class="text-sm text-slate-500 dark:text-slate-400">Set your primary branch and build triggers in the intuitive settings panel.</p>
            </div>
            <div class="text-center space-y-4">
                <div class="text-4xl font-mono font-bold text-primary dark:text-primary-dark opacity-30">03</div>
                <h4 class="text-lg font-bold dark:text-white">Ship Fast</h4>
                <p class="text-sm text-slate-500 dark:text-slate-400">Watch your jobs run and enjoy automatically updated code.</p>
            </div>
        </div>'''

new_grid = '''        <div class="grid md:grid-cols-2 lg:grid-cols-3 gap-12">
            <div class="text-center space-y-4">
                <div class="text-4xl font-mono font-bold text-primary dark:text-primary-dark opacity-30">01</div>
                <h4 class="text-lg font-bold dark:text-white">Connect Repo</h4>
                <p class="text-sm text-slate-500 dark:text-slate-400">Provide your Git SSH URL and optional credentials to link your project.</p>
            </div>
            <div class="text-center space-y-4">
                <div class="text-4xl font-mono font-bold text-primary dark:text-primary-dark opacity-30">02</div>
                <h4 class="text-lg font-bold dark:text-white">Configure Build</h4>
                <p class="text-sm text-slate-500 dark:text-slate-400">Set your primary branch and build triggers in the intuitive settings panel.</p>
            </div>
            <div class="text-center space-y-4">
                <div class="text-4xl font-mono font-bold text-primary dark:text-primary-dark opacity-30">03</div>
                <h4 class="text-lg font-bold dark:text-white">Full Visibility</h4>
                <p class="text-sm text-slate-500 dark:text-slate-400">Every agentic change is logged in detail. Monitor build logs in real-time to see exactly how your code is being updated.</p>
            </div>
            <div class="text-center space-y-4">
                <div class="text-4xl font-mono font-bold text-primary dark:text-primary-dark opacity-30">04</div>
                <h4 class="text-lg font-bold dark:text-white">Ship Fast</h4>
                <p class="text-sm text-slate-500 dark:text-slate-400">Watch your jobs run and enjoy automatically updated code without manual intervention.</p>
            </div>
            <div class="text-center space-y-4">
                <div class="text-4xl font-mono font-bold text-primary dark:text-primary-dark opacity-30">05</div>
                <h4 class="text-lg font-bold dark:text-white">Rewind Anytime</h4>
                <p class="text-sm text-slate-500 dark:text-slate-400">Not satisfied with an automated change? Use Rewind to instantly hard-reset and force-push back to any previous stable commit.</p>
            </div>
        </div>'''

if old_grid in content:
    content = content.replace(old_grid, new_grid)
else:
    print("Could not find old_grid in content")
    sys.exit(1)

with open(file_path, 'w') as f:
    f.write(content)
