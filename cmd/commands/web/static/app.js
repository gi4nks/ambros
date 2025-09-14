// Extracted front-end script for Ambros Dashboard
// This file contains the client-side functions previously embedded in server.go

async function loadSection(section) {
    const content = document.getElementById('main-content');
    content.innerHTML = '<div class="loading">Loading...</div>';

    try {
        switch(section) {
            case 'dashboard':
                await loadDashboard();
                break;
            case 'commands':
                await loadCommands();
                break;
            case 'environments':
                await loadEnvironments();
                break;
            case 'templates':
                await loadTemplates();
                break;
            case 'scheduler':
                await loadScheduler();
                break;
            case 'analytics':
                await loadAnalytics();
                break;
        }
    } catch (error) {
        content.innerHTML = '<div style="color: red;">Error loading section: ' + error.message + '</div>';
    }
}

async function loadDashboard() {
    const response = await fetch('/api/dashboard');
    const data = await response.json();

    const content = document.getElementById('main-content');
    content.innerHTML = `
        <h2>📊 Dashboard Overview</h2>
        <div class="stats">
            <div class="stat-card">
                <div class="stat-value">${data.summary.total_commands}</div>
                <div class="stat-label">Total Commands</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">${data.summary.success_rate.toFixed(1)}%</div>
                <div class="stat-label">Success Rate</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">${data.summary.recent_commands}</div>
                <div class="stat-label">Recent (24h)</div>
            </div>
            <div class="stat-card">
                <div class="stat-value">${data.summary.template_count}</div>
                <div class="stat-label">Templates</div>
            </div>
        </div>
        <h3>Recent Activity</h3>
        <div style="max-height: 400px; overflow-y: auto;">
            ${data.recent_activity.map(cmd => `
                <div style="padding: 0.5rem; border-bottom: 1px solid #eee;">
                    <strong>${cmd.command || cmd.name}</strong>
                    <span style="color: ${cmd.status ? 'green' : 'red'}; margin-left: 1rem;">
                        ${cmd.status ? '✅' : '❌'}
                    </span>
                    <small style="float: right; color: #666;">
                        ${new Date(cmd.created_at).toLocaleString()}
                    </small>
                </div>
            `).join('')}
        </div>
    `;
}

async function loadCommands() {
    const response = await fetch('/api/commands');
    const commands = await response.json();

    const content = document.getElementById('main-content');
    content.innerHTML = `
        <h2>💻 Command History</h2>
        <p>Total: ${commands.length} commands</p>
        <div style="max-height: 600px; overflow-y: auto;">
            ${commands.map(cmd => `
                <div style="padding: 1rem; border-bottom: 1px solid #eee; margin-bottom: 0.5rem;">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <strong>${cmd.command || cmd.name}</strong>
                        <span style="color: ${cmd.status ? 'green' : 'red'};">
                            ${cmd.status ? '✅ Success' : '❌ Failed'}
                        </span>
                    </div>
                    <small style="color: #666;">
                        ID: ${cmd.id} | ${new Date(cmd.created_at).toLocaleString()}
                    </small>
                    ${cmd.tags && cmd.tags.length > 0 ? `
                        <div style="margin-top: 0.5rem;">
                            ${cmd.tags.map(tag => `<span style="background: #667eea; color: white; padding: 0.2rem 0.5rem; border-radius: 3px; font-size: 0.8rem; margin-right: 0.5rem;">${tag}</span>`).join('')}
                        </div>
                    ` : ''}
                </div>
            `).join('')}
        </div>
    `;
}

async function loadEnvironments() {
    const response = await fetch('/api/environments');
    const environments = await response.json();

    const content = document.getElementById('main-content');
    content.innerHTML = `
        <h2>🌍 Environments</h2>
        ${Object.keys(environments).length === 0 ? 
            '<p>No environments found. Create one using: <code>ambros env create myenv</code></p>' :
            Object.entries(environments).map(([name, env]) => `
                <div style="background: #f8f9fa; padding: 1rem; border-radius: 8px; margin-bottom: 1rem;">
                    <h3>${name}</h3>
                    <div style="display: grid; gap: 0.5rem;">
                        ${env.variables.map(v => `
                            <div style="background: white; padding: 0.5rem; border-radius: 4px;">
                                <strong>${v.key}</strong> = <code>${v.value}</code>
                            </div>
                        `).join('')}
                    </div>
                </div>
            `).join('')
        }
    `;
}

async function loadTemplates() {
    const response = await fetch('/api/templates');
    const templates = await response.json();

    const content = document.getElementById('main-content');
    content.innerHTML = `
        <h2>🎯 Command Templates</h2>
        ${templates.length === 0 ? 
            '<p>No templates found. Create one using: <code>ambros template save mytemplate "echo hello"</code></p>' :
            templates.map(template => `
                <div style="background: #f8f9fa; padding: 1rem; border-radius: 8px; margin-bottom: 1rem;">
                    <h3>${template.name || 'Unnamed Template'}</h3>
                    <code style="background: white; padding: 0.5rem; border-radius: 4px; display: block;">
                        ${template.command}
                    </code>
                    <small style="color: #666;">
                        Created: ${new Date(template.created_at).toLocaleString()}
                    </small>
                </div>
            `).join('')
        }
    `;
}

async function loadScheduler() {
    const response = await fetch('/api/scheduler');
    const scheduled = await response.json();

    const content = document.getElementById('main-content');
    content.innerHTML = `
        <h2>📅 Scheduled Commands</h2>
        ${scheduled.length === 0 ? 
            '<p>No scheduled commands found.</p>' :
            scheduled.map(cmd => `
                <div style="background: #f8f9fa; padding: 1rem; border-radius: 8px; margin-bottom: 1rem;">
                    <div style="display: flex; justify-content: space-between;">
                        <strong>${cmd.command || cmd.name}</strong>
                        <span style="color: ${cmd.schedule.enabled ? 'green' : 'red'};">
                            ${cmd.schedule.enabled ? '🟢 Enabled' : '🔴 Disabled'}
                        </span>
                    </div>
                    <div style="margin-top: 0.5rem;">
                        <strong>Cron:</strong> <code>${cmd.schedule.cron_expr}</code><br>
                        <strong>Next Run:</strong> ${new Date(cmd.schedule.next_run).toLocaleString()}
                    </div>
                </div>
            `).join('')
        }
    `;
}

async function loadAnalytics() {
    const response = await fetch('/api/analytics/advanced');
    const analytics = await response.json();

    const cp = analytics.command_patterns || {};
    const et = analytics.execution_trends || {};
    const fa = analytics.failure_analysis || {};
    const pm = analytics.performance_metrics || {};
    const up = analytics.usage_predictions || {};
    const recs = analytics.recommendations || [];

    const content = document.getElementById('main-content');
    var out = '';
    out += '<h2>📈 Advanced Analytics</h2>';
    out += '<div style="display:grid;grid-template-columns:repeat(auto-fit,minmax(300px,1fr));gap:1rem;">';

    // Command Patterns
    out += '<div style="background:#f8f9fa;padding:1rem;border-radius:8px;">';
    out += '<h3>🔍 Command Patterns</h3>';
    out += '<p><strong>Most common:</strong></p>';
    out += '<ul>';
    (cp.most_common || []).forEach(function(c){ out += '<li>' + c + '</li>'; });
    out += '</ul>';
    out += '<p><strong>Detected patterns:</strong> ' + ((cp.patterns || []).join(', ')) + '</p>';
    out += '</div>';

    // Execution Trends
    out += '<div style="background:#f8f9fa;padding:1rem;border-radius:8px;">';
    out += '<h3>📊 Execution Trends</h3>';
    out += '<p><strong>Trend:</strong> ' + (et.trend || 'stable') + '</p>';
    out += '<p><strong>Predicted peak hour:</strong> ' + (up.predicted_peak || 'N/A') + '</p>';
    out += '<p><strong>Peak hours:</strong> ' + ((et.peak_hours || []).join(', ')) + '</p>';
    out += '<div style="max-height:180px;overflow:auto;margin-top:0.5rem;background:white;padding:0.5rem;border-radius:4px;">';
    out += '<strong>Activity by day</strong><ul>';
    Object.keys(et.by_day || {}).sort().forEach(function(d){ out += '<li>' + d + ': ' + (et.by_day[d] || 0) + '</li>'; });
    out += '</ul></div></div>';

    // Failure Analysis
    out += '<div style="background:#f8f9fa;padding:1rem;border-radius:8px;">';
    out += '<h3>🔧 Failure Analysis</h3>';
    out += '<p><strong>Total failures:</strong> ' + (fa.total_failures || 0) + '</p>';
    var failureRate = fa.failure_rate;
    if (typeof failureRate === 'number') {
        out += '<p><strong>Failure rate:</strong> ' + (failureRate.toFixed(1) + '%') + '</p>';
    } else {
        out += '<p><strong>Failure rate:</strong> 0.0%</p>';
    }
    out += '<p><strong>Common causes:</strong> ' + ((fa.common_causes || []).join(', ')) + '</p>';
    out += '</div>';

    // Performance Metrics
    out += '<div style="background:#f8f9fa;padding:1rem;border-radius:8px;">';
    out += '<h3>⏱️ Performance Metrics</h3>';
    out += '<p><strong>Average duration:</strong> ' + (pm.avg_duration || 'N/A') + '</p>';
    out += '<p><strong>Slowest commands:</strong></p><ol>';
    (pm.slowest_commands || []).forEach(function(s){ out += '<li><code>' + s.command + '</code> — ' + s.duration + '</li>'; });
    out += '</ol></div>';

    // Usage Predictions & Recommendations
    out += '<div style="background:#f8f9fa;padding:1rem;border-radius:8px;grid-column:span 2;">';
    out += '<h3>📈 Usage Predictions & Trends</h3>';
    out += '<p><strong>Predicted peak:</strong> ' + (up.predicted_peak || 'N/A') + '</p>';
    out += '<p><strong>Trending commands:</strong> ' + ((up.trending_commands || []).join(', ')) + '</p>';
    out += '<div style="margin-top:0.5rem;"><strong>Recommendations:</strong><ul>';
    (recs || []).forEach(function(r){ out += '<li>' + r + '</li>'; });
    out += '</ul></div></div>';

    out += '</div>';
    content.innerHTML = out;
}

// Load dashboard on page load
window.onload = () => loadSection('dashboard');

// Expose functions for onclick attributes
window.loadSection = loadSection;
