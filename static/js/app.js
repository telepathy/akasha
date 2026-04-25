const API_BASE = '/api/v1';
let currentBranch = '';
let currentBranchStatus = '';

function initTheme() {
    const saved = localStorage.getItem('theme');
    if (saved) {
        document.documentElement.setAttribute('data-theme', saved);
    } else {
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        document.documentElement.setAttribute('data-theme', prefersDark ? 'dark' : 'light');
    }
    updateThemeToggle();
}

function toggleTheme() {
    const current = document.documentElement.getAttribute('data-theme');
    const next = current === 'dark' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', next);
    localStorage.setItem('theme', next);
    updateThemeToggle();
}

function updateThemeToggle() {
    const btn = document.getElementById('themeToggle');
    if (btn) {
        const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
        btn.textContent = isDark ? '☀️' : '🌙';
        btn.title = isDark ? '切换到亮色模式' : '切换到暗色模式';
    }
}

initTheme();

async function loadBranches() {
    try {
        const res = await fetch(API_BASE + '/branches');
        const branches = await res.json();
        // 保存分支数据供各页面使用
        window.allBranches = branches;
        if (document.getElementById('branchList')) renderBranchList(branches);
        if (document.getElementById('currentBranch')) renderBranchSelect(branches);
        if (document.getElementById('branchTable')) renderBranchTable(branches);
        return branches;
    } catch (e) { console.error('loadBranches error:', e); }
}

function renderBranchList(branches) {
    const list = document.getElementById('branchList');
    if (!list) return;
    list.innerHTML = branches.map(b => `<span class="tag ${b.name === currentBranch ? 'active' : ''}" onclick="selectBranch('${b.name}', '${b.status}')">${b.name}</span>`).join('');
}

function renderBranchSelect(branches) {
    const select = document.getElementById('currentBranch');
    if (!select) return;
    select.innerHTML = '<option value="">选择分支</option>' + branches.map(b => `<option value="${b.name}">${b.name}</option>`).join('');
}

function selectBranch(name, status) {
    currentBranch = name;
    currentBranchStatus = status;
    loadDependencies(name);
    // 只更新选中状态，不重新渲染整个列表
    updateBranchListActive(name);
}

// 更新分支列表中的active状态
function updateBranchListActive(activeName) {
    const list = document.getElementById('branchList');
    if (!list) return;
    const tags = list.querySelectorAll('.tag');
    tags.forEach(tag => {
        if (tag.textContent === activeName) {
            tag.classList.add('active');
        } else {
            tag.classList.remove('active');
        }
    });
}

function switchBranch(name) {
    if (name) {
        // Find the status of selected branch
        fetch(API_BASE + '/branches').then(r => r.json()).then(branches => {
            const b = branches.find(x => x.name === name);
            selectBranch(name, b ? b.status : '');
        });
    }
}

async function loadDependencies(branch) {
    const list = document.getElementById('depList');
    if (!list) return;
    if (!branch) { list.innerHTML = '请先选择分支'; return; }
    try {
        const res = await fetch(API_BASE + '/dependencies?branch=' + branch);
        const deps = await res.json();
        renderDepList(deps);
    } catch (e) { list.innerHTML = '加载失败: ' + e.message; }
}

function renderDepList(deps) {
    const list = document.getElementById('depList');
    if (!list) return;
    if (deps.length === 0) { list.innerHTML = '暂无依赖'; return; }
    // 首页只显示列表，不显示操作按钮（操作在分支管理页面进行）
    list.innerHTML = '<table class="table"><thead><tr><th>Group ID</th><th>Artifact ID</th><th>Version</th></tr></thead><tbody>' +
        deps.map(d => `<tr><td>${d.groupId}</td><td>${d.artifact}</td><td>${d.version}</td></tr>`).join('') + '</tbody></table>';
}

function showModal(content) {
    const modal = document.getElementById('modal');
    const body = document.getElementById('modalBody');
    if (modal && body) { body.innerHTML = content; modal.classList.remove('hidden'); }
}

function closeModal() {
    const modal = document.getElementById('modal');
    if (modal) modal.classList.add('hidden');
}

document.addEventListener('click', function(e) {
    const modal = document.getElementById('modal');
    if (modal && !modal.classList.contains('hidden') && e.target === modal) closeModal();
});
document.addEventListener('keydown', function(e) { if (e.key === 'Escape') closeModal(); });

function renderBranchTable(branches) {
    const table = document.getElementById('branchTable');
    if (!table) return;
    window.branches = branches;
    if (branches.length === 0) { table.innerHTML = '<tr><td colspan="6">暂无分支</td></tr>'; return; }
    table.innerHTML = branches.map(b => `<tr>
        <td>${b.name}</td>
        <td>${b.baseBranch || '-'}</td>
        <td><span class="status-tag status-${b.status}">${b.status}</span></td>
        <td>${new Date(b.createdAt).toLocaleString()}</td>
        <td class="btn-group">
            <button class="btn-xs" onclick="viewBranchDeps('${b.name}')">查看</button>
            ${b.status === 'active' ? `<button class="btn-xs" onclick="showAddGav('${b.name}')">+GAV</button>` : ''}
            <button class="btn-xs" onclick="viewBranchHistory('${b.name}')">变更</button>
            <button class="btn-xs" onclick="showFlashback('${b.name}')">闪回</button>
            ${b.status === 'active' ? `<button class="btn-xs" onclick="archiveBranch('${b.name}')">锁定</button>` : ''}
            ${b.status === 'archived' ? `<button class="btn-xs" onclick="unlockBranch('${b.name}')">解锁</button>` : ''}
            ${b.status === 'active' ? `<button class="btn-xs" onclick="deleteBranch('${b.name}')">删除</button>` : ''}
        </td>
    </tr>`).join('');
}

function viewBranch(name) {
    showModal(`<h2>分支: ${name}</h2><div id="branchDeps">加载中...</div>`);
    fetch(API_BASE + '/dependencies?branch=' + name).then(r => r.json()).then(deps => {
        const div = document.getElementById('branchDeps');
        if (deps.length === 0) { div.innerHTML = '暂无依赖'; return; }
        div.innerHTML = '<table class="table"><thead><tr><th>Group</th><th>Artifact</th><th>Version</th></tr></thead><tbody>' + deps.map(d => `<tr><td>${d.groupId}</td><td>${d.artifact}</td><td>${d.version}</td></tr>`).join('') + '</tbody></table>';
    });
}

function viewBranchDeps(name) {
    window.open('/api/v1/branches/' + name + '/deps-text', '_blank');
}

function viewBranchHistory(name) {
    showModal(`<h2>分支 ${name} 变更历史</h2>
        <form onsubmit="queryHistory(event, '${name}')">
            <div class="form-group" style="display:flex;gap:10px">
                <div><label>开始时间</label><input type="datetime-local" name="startAt"></div>
                <div><label>结束时间</label><input type="datetime-local" name="endAt"></div>
            </div>
            <div class="form-actions"><button type="submit">查询</button><button type="button" onclick="closeModal()">关闭</button></div>
        </form><div id="historyResult"></div>`);
}

function queryHistory(e, name) {
    e.preventDefault();
    const form = e.target;
    const startAt = form.startAt.value;
    const endAt = form.endAt.value;
    if (!startAt || !endAt) { alert('请选择开始和结束时间'); return; }
    // 先获取所有依赖名称，然后对每个依赖按时间范围查询
    fetch(API_BASE + '/dependencies?branch=' + name).then(r => r.json()).then(deps => {
        const div = document.getElementById('historyResult');
        if (deps.length === 0) { div.innerHTML = '暂无依赖'; return; }
        // 构造并行查询
        const promises = deps.map(dep => {
            const startISO = new Date(startAt).toISOString();
            const endISO = new Date(endAt).toISOString();
            return fetch(API_BASE + '/dependencies/' + dep.name + '/history-between?branch=' + name + '&startAt=' + encodeURIComponent(startISO) + '&endAt=' + encodeURIComponent(endISO))
                .then(r => r.json())
                .then(history => ({name: dep.name, history}));
        });
        Promise.all(promises).then(results => {
            // 过滤有变更的
            const changed = results.filter(r => r.history && r.history.length > 0);
            if (changed.length === 0) {
                div.innerHTML = '<p>该时间段内无变更</p>';
                return;
            }
            let html = `<p>共 ${changed.length} 个依赖在此时期内有变更</p><table class="table"><thead><tr><th>依赖名</th><th>变更次数</th><th>最新版本</th><th>最后变更时间</th></tr></thead><tbody>`;
            changed.forEach(item => {
                const lastTime = item.history[0] ? new Date(item.history[0].createdAt).toLocaleString() : '-';
                html += `<tr><td>${item.name}</td><td>${item.history.length}</td><td>${item.history[0].version}</td><td>${lastTime}</td></tr>`;
            });
            html += '</tbody></table>';
            div.innerHTML = html;
        });
    });
}

function showFlashback(name) {
    showModal(`<h2>分支 ${name} 闪回查询</h2><p>选择时间点，查看当时该分支的依赖快照</p>
        <form onsubmit="queryFlashback(event, '${name}')">
            <div class="form-group"><label>查询时间</label><input type="datetime-local" name="at" required></div>
            <div class="form-actions"><button type="submit">查询</button><button type="button" onclick="closeModal()">取消</button></div>
        </form><div id="flashbackResult"></div>`);
}

function queryFlashback(e, name) {
    e.preventDefault();
    const form = e.target;
    const at = form.at.value;
    if (!at) { alert('请输入查询时间'); return; }
    // 转换为ISO格式传递给后端
    const atISO = new Date(at).toISOString();
    // 调用批量闪回API
    fetch(API_BASE + '/branches/' + name + '/deps-at?at=' + encodeURIComponent(atISO)).then(r => r.json()).then(deps => {
        const div = document.getElementById('flashbackResult');
        if (deps.length === 0) { div.innerHTML = '该时间点无依赖数据'; return; }
        div.innerHTML = `<p>${at} 时该分支共有 ${deps.length} 个依赖</p><table class="table"><thead><tr><th>Group</th><th>Artifact</th><th>Version</th></tr></thead><tbody>` + deps.map(d => `<tr><td>${d.groupId}</td><td>${d.artifact}</td><td>${d.version}</td></tr>`).join('') + '</tbody></table>';
    });
}

function showAddGav(name) {
    showModal(`<h2>添加 GAV - ${name}</h2>
        <form onsubmit="addGav(event, '${name}')">
            <div class="form-group"><label>名称</label><input type="text" name="name" required placeholder="如: spring-core"></div>
            <div class="form-group"><label>Group ID</label><input type="text" name="groupId" required placeholder="如: org.springframework"></div>
            <div class="form-group"><label>Artifact ID</label><input type="text" name="artifact" required placeholder="如: spring-core"></div>
            <div class="form-group"><label>Version</label><input type="text" name="version" required placeholder="如: 6.2.7"></div>
            <div class="form-group"><label>备注</label><textarea name="remark"></textarea></div>
            <div class="form-actions"><button type="submit">保存</button><button type="button" onclick="closeModal()">取消</button></div>
        </form>`);
}

function addGav(e, branch) {
    e.preventDefault();
    const form = e.target;
    const data = {name: form.name.value, groupId: form.groupId.value, artifact: form.artifact.value, version: form.version.value, branch: branch, remark: form.remark.value};
    fetch(API_BASE + '/dependencies', {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(data)})
        .then(res => res.json())
        .then(d => { closeModal(); loadBranches(); if (currentBranch === branch) loadDependencies(branch); })
        .catch(err => alert('保存失败: ' + err.message));
}

function archiveBranch(name) {
    if (!confirm('确定锁定分支 ' + name + '？锁定后不可修改，但可继续获取依赖')) return;
    fetch(API_BASE + '/branches/' + name + '/archive', {method: 'POST'}).then(() => loadBranches());
}

function unlockBranch(name) {
    if (!confirm('确定解锁分支 ' + name + '？解锁后可继续修改')) return;
    fetch(API_BASE + '/branches/' + name + '/unlock', {method: 'POST'}).then(() => loadBranches());
}

function deleteBranch(name) {
    if (!confirm('确定删除分支 ' + name + '？删除后不可使用')) return;
    fetch(API_BASE + '/branches/' + name, {method: 'DELETE'}).then(() => loadBranches());
}

function showCreateBranch() {
    showModal(`<h2>新建分支</h2>
        <form onsubmit="createBranch(event)">
            <div class="form-group"><label>分支名称</label><input type="text" name="name" required placeholder="如: 202603"></div>
            <div class="form-group"><label>基分支（可选）</label><select name="baseBranch"><option value="">无</option></select></div>
            <div class="form-actions"><button type="submit">创建</button><button type="button" onclick="closeModal()">取消</button></div>
        </form>`);
    loadBranches().then(branches => {
        const select = document.querySelector('select[name="baseBranch"]');
        if (select) select.innerHTML = '<option value="">无</option>' + branches.map(b => `<option value="${b.name}">${b.name}</option>`).join('');
    });
}

function showImportBranch() {
    showModal(`<h2>导入分支</h2>
        <form onsubmit="importBranch(event)">
            <div class="form-group"><label>分支名称</label><input type="text" name="name" required placeholder="如: 202603"></div>
            <div class="form-group"><label>选择文件</label><input type="file" name="file" accept=".gradle,.txt" required></div>
            <div class="form-actions"><button type="submit">导入</button><button type="button" onclick="closeModal()">取消</button></div>
        </form>`);
}

async function importBranch(e) {
    e.preventDefault();
    const form = e.target;
    const name = form.name.value;
    const file = form.file.files[0];
    if (!name || !file) { alert('请填写分支名称并选择文件'); return; }

    const reader = new FileReader();
    reader.onload = async function(event) {
        const content = event.target.result;
        const deps = parseGradleFile(content);
        if (deps.length === 0) { alert('未解析到依赖条目'); return; }

        try {
            await fetch(API_BASE + '/branches', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({name: name, baseBranch: ''})
            });
        } catch (err) { /* 分支可能已存在 */ }

        let success = 0, fail = 0;
        for (const dep of deps) {
            const res = await fetch(API_BASE + '/dependencies', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({...dep, branch: name})
            });
            if (res.ok) success++; else fail++;
        }

        closeModal();
        loadBranches();
        alert('导入完成：成功 ' + success + ' 条' + (fail > 0 ? '，失败 ' + fail + ' 条' : ''));
    };
    reader.readAsText(file);
}

function parseGradleFile(content) {
    const deps = [];
    const re = /"([^"]+)"\s*:\s*"([^:]+):([^:]+):([^"]+)"/g;
    let match;
    while ((match = re.exec(content)) !== null) {
        deps.push({
            name: match[1],
            groupId: match[2],
            artifact: match[3],
            version: match[4]
        });
    }
    return deps;
}

async function createBranch(e) {
    e.preventDefault();
    const form = e.target;
    const data = {name: form.name.value, baseBranch: form.baseBranch.value};
    try {
        const res = await fetch(API_BASE + '/branches', {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(data)});
        if (res.ok) { closeModal(); loadBranches(); }
        else { const err = await res.json(); alert(err.error || '创建失败'); }
    } catch (e) { alert('创建失败: ' + e.message); }
}

function viewHistory(name) {
    if (!currentBranch) { alert('请先选择分支'); return; }
    showModal(`<h2>${name} 版本历史</h2><div id="historyList">加载中...</div>`);
    fetch(API_BASE + '/dependencies/' + name + '/history?branch=' + currentBranch).then(r => r.json()).then(history => {
        const div = document.getElementById('historyList');
        div.innerHTML = '<table class="table"><thead><tr><th>Version</th><th>时间</th></tr></thead><tbody>' + history.map(h => `<tr><td>${h.version}</td><td>${new Date(h.createdAt).toLocaleString()}</td></tr>`).join('') + '</tbody></table>';
    });
}

function editDep(name) {
    if (currentBranchStatus !== 'active') { alert('只有 active 状态的分支才能编辑依赖'); return; }
    showModal(`<h2>编辑依赖 - ${name}</h2>
        <form onsubmit="updateDep(event, '${name}')">
            <div class="form-group"><label>新版本</label><input type="text" name="version" required placeholder="如: 6.2.8"></div>
            <div class="form-group"><label>备注</label><textarea name="remark"></textarea></div>
            <div class="form-actions"><button type="submit">保存</button><button type="button" onclick="closeModal()">取消</button></div>
        </form>`);
}

function updateDep(e, name) {
    e.preventDefault();
    const form = e.target;
    const data = {name: name, groupId: '', artifact: '', version: form.version.value, branch: currentBranch, remark: form.remark.value};
    fetch(API_BASE + '/dependencies', {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(data)})
        .then(res => res.json())
        .then(d => { closeModal(); loadDependencies(currentBranch); })
        .catch(err => alert('保存失败: ' + err.message));
}
