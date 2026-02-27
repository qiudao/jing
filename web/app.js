(function() {
    'use strict';

    let ws = null;
    let gameActive = false;
    let currentMode = 'hva';
    let chartData = [];

    // DOM elements
    const board = document.getElementById('board');
    const cells = document.querySelectorAll('.cell');
    const statusEl = document.getElementById('game-status');
    const connStatus = document.getElementById('connection-status');
    const speedSlider = document.getElementById('speed-slider');
    const speedValue = document.getElementById('speed-value');
    const speedSection = document.getElementById('speed-section');
    const difficultySection = document.getElementById('difficulty-section');
    const trainProgressContainer = document.getElementById('train-progress-container');
    const trainProgressFill = document.getElementById('train-progress-fill');
    const trainProgressText = document.getElementById('train-progress-text');
    const chart = document.getElementById('chart');
    const chartCtx = chart.getContext('2d');

    function connect() {
        const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
        ws = new WebSocket(protocol + '//' + location.host + '/ws');

        ws.onopen = function() {
            connStatus.className = 'status-dot connected';
            ws.send(JSON.stringify({type: 'get_stats', data: {}}));
        };

        ws.onclose = function() {
            connStatus.className = 'status-dot disconnected';
            gameActive = false;
            setTimeout(connect, 2000);
        };

        ws.onerror = function() {
            connStatus.className = 'status-dot disconnected';
        };

        ws.onmessage = function(e) {
            var msg = JSON.parse(e.data);
            handleMessage(msg);
        };
    }

    function handleMessage(msg) {
        switch (msg.type) {
            case 'game_state':
                updateBoard(msg.data);
                break;
            case 'train_progress':
                updateTrainProgress(msg.data);
                break;
            case 'train_done':
                onTrainDone(msg.data);
                break;
            case 'watch_game':
                statusEl.textContent = 'Training: Episode ' + msg.data.episode + '/' + msg.data.total;
                break;
            case 'stats':
                updateStats(msg.data);
                break;
            case 'error':
                statusEl.textContent = 'Error: ' + msg.data.message;
                break;
        }
    }

    function updateBoard(state) {
        var symbols = ['', 'X', 'O'];
        var classes = ['', 'x', 'o'];
        cells.forEach(function(cell, i) {
            cell.textContent = symbols[state.board[i]];
            cell.className = 'cell' + (state.board[i] ? ' ' + classes[state.board[i]] : '');
        });

        if (state.status === 'playing') {
            var turnName = state.turn === 1 ? 'X' : 'O';
            statusEl.textContent = turnName + "'s turn";
            gameActive = true;
        } else if (state.status === 'win') {
            var winner = state.winner === 1 ? 'X' : 'O';
            statusEl.textContent = winner + ' wins!';
            gameActive = false;
            requestStats();
        } else if (state.status === 'draw') {
            statusEl.textContent = 'Draw!';
            gameActive = false;
            requestStats();
        }
    }

    function requestStats() {
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({type: 'get_stats', data: {}}));
        }
    }

    function updateStats(stats) {
        // Try mode-specific key first, then fall back
        var s = stats[currentMode] || {xWins: 0, oWins: 0, draws: 0};
        document.getElementById('stat-xwins').textContent = s.xWins || 0;
        document.getElementById('stat-owins').textContent = s.oWins || 0;
        document.getElementById('stat-draws').textContent = s.draws || 0;
    }

    function updateTrainProgress(data) {
        trainProgressContainer.classList.remove('hidden');
        var pct = Math.round((data.episode / data.total) * 100);
        trainProgressFill.style.width = pct + '%';
        trainProgressText.textContent = data.episode + '/' + data.total + ' (' + pct + '%)';

        chartData.push({episode: data.episode, winRate: data.winRateX});
        drawChart();
    }

    function onTrainDone(data) {
        statusEl.textContent = 'Training complete: ' + data.difficulty;
        var btn = document.getElementById('btn-train');
        btn.textContent = 'Start Training';
        btn.disabled = false;
    }

    function drawChart() {
        var w = chart.width;
        var h = chart.height;
        chartCtx.clearRect(0, 0, w, h);

        // Background
        chartCtx.fillStyle = '#0a0a1a';
        chartCtx.fillRect(0, 0, w, h);

        if (chartData.length < 2) return;

        var pad = 40;

        // Grid lines
        chartCtx.strokeStyle = '#222';
        chartCtx.lineWidth = 1;
        for (var i = 0; i <= 4; i++) {
            var val = i * 0.25;
            var y = h - pad - (val * (h - 2 * pad));
            chartCtx.beginPath();
            chartCtx.moveTo(pad, y);
            chartCtx.lineTo(w - pad, y);
            chartCtx.stroke();
        }

        // Axes
        chartCtx.strokeStyle = '#444';
        chartCtx.lineWidth = 1;
        chartCtx.beginPath();
        chartCtx.moveTo(pad, pad);
        chartCtx.lineTo(pad, h - pad);
        chartCtx.lineTo(w - pad, h - pad);
        chartCtx.stroke();

        // Y axis labels
        chartCtx.fillStyle = '#888';
        chartCtx.font = '11px monospace';
        chartCtx.textAlign = 'right';
        for (var j = 0; j <= 4; j++) {
            var v = j * 0.25;
            var ly = h - pad - (v * (h - 2 * pad));
            chartCtx.fillText(v.toFixed(2), pad - 5, ly + 4);
        }

        // Axis titles
        chartCtx.textAlign = 'left';
        chartCtx.fillText('Win Rate (X)', pad + 5, pad - 8);
        chartCtx.textAlign = 'right';
        chartCtx.fillText('Episodes', w - pad, h - 5);

        // Data line
        var maxEp = chartData[chartData.length - 1].episode;
        chartCtx.strokeStyle = '#e94560';
        chartCtx.lineWidth = 2;
        chartCtx.beginPath();
        chartData.forEach(function(d, idx) {
            var x = pad + (d.episode / maxEp) * (w - 2 * pad);
            var cy = h - pad - (d.winRate * (h - 2 * pad));
            if (idx === 0) chartCtx.moveTo(x, cy);
            else chartCtx.lineTo(x, cy);
        });
        chartCtx.stroke();

        // 50% reference line
        chartCtx.strokeStyle = '#4ecca3';
        chartCtx.lineWidth = 1;
        chartCtx.setLineDash([4, 4]);
        var refY = h - pad - (0.5 * (h - 2 * pad));
        chartCtx.beginPath();
        chartCtx.moveTo(pad, refY);
        chartCtx.lineTo(w - pad, refY);
        chartCtx.stroke();
        chartCtx.setLineDash([]);
    }

    function getSelectedMode() {
        return document.querySelector('input[name="mode"]:checked').value;
    }

    function getSelectedDifficulty() {
        return document.querySelector('input[name="difficulty"]:checked').value;
    }

    function clearBoard() {
        cells.forEach(function(cell) {
            cell.textContent = '';
            cell.className = 'cell';
        });
    }

    // === Event Listeners ===

    // Cell clicks
    cells.forEach(function(cell) {
        cell.addEventListener('click', function() {
            if (!gameActive) return;
            if (currentMode === 'ava') return;
            var pos = parseInt(cell.dataset.pos);
            if (cell.textContent !== '') return;
            ws.send(JSON.stringify({type: 'move', data: {position: pos}}));
        });
    });

    // Mode selection
    document.querySelectorAll('input[name="mode"]').forEach(function(radio) {
        radio.addEventListener('change', function() {
            var mode = getSelectedMode();
            speedSection.classList.toggle('hidden', mode !== 'ava');
            difficultySection.classList.toggle('hidden', mode === 'hvh');
        });
    });

    // Speed slider
    speedSlider.addEventListener('input', function() {
        speedValue.textContent = speedSlider.value + 'ms';
    });

    // Start button
    document.getElementById('btn-start').addEventListener('click', function() {
        currentMode = getSelectedMode();
        var data = {
            mode: currentMode,
            difficulty: getSelectedDifficulty(),
            speed: parseInt(speedSlider.value)
        };
        ws.send(JSON.stringify({type: 'start_game', data: data}));
        gameActive = true;
    });

    // Reset button
    document.getElementById('btn-reset').addEventListener('click', function() {
        clearBoard();
        gameActive = false;
        statusEl.textContent = 'Select mode and click Start';
        currentMode = getSelectedMode();
        var data = {
            mode: currentMode,
            difficulty: getSelectedDifficulty(),
            speed: parseInt(speedSlider.value)
        };
        ws.send(JSON.stringify({type: 'start_game', data: data}));
    });

    // Train button
    document.getElementById('btn-train').addEventListener('click', function() {
        var btn = document.getElementById('btn-train');
        btn.textContent = 'Training...';
        btn.disabled = true;
        chartData = [];
        drawChart();
        var data = {
            difficulty: document.getElementById('train-difficulty').value,
            watch: document.getElementById('train-watch').checked,
            speed: parseInt(speedSlider.value)
        };
        ws.send(JSON.stringify({type: 'train', data: data}));
    });

    // Handle high-DPI canvas
    function setupCanvas() {
        var dpr = window.devicePixelRatio || 1;
        var rect = chart.getBoundingClientRect();
        chart.width = rect.width * dpr;
        chart.height = rect.height * dpr;
        chartCtx.scale(dpr, dpr);
        // Reset canvas CSS size
        chart.style.width = rect.width + 'px';
        chart.style.height = rect.height + 'px';
    }

    // Initialize
    setupCanvas();
    connect();

    window.addEventListener('resize', function() {
        setupCanvas();
        if (chartData.length > 0) drawChart();
    });
})();
