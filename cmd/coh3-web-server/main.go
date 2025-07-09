package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/scharissis/coh3-replay-analyser/vault"
)

const defaultDataDir = "./data/coh3-data"

type WebServer struct {
	dataDir string
	port    int
}

type TimelineEvent struct {
	PlayerID     int    `json:"player_id"`
	PlayerName   string `json:"player_name"`
	Faction      string `json:"faction"`
	Timestamp    uint32 `json:"timestamp"`
	TimestampStr string `json:"timestamp_str"`
	CommandType  string `json:"command_type"`
	Description  string `json:"description"`
	Color        string `json:"color"`
}

type ReplayResponse struct {
	Success    bool            `json:"success"`
	Error      string          `json:"error,omitempty"`
	MapName    string          `json:"map_name,omitempty"`
	Duration   string          `json:"duration,omitempty"`
	Players    []PlayerSummary `json:"players,omitempty"`
	Timeline   []TimelineEvent `json:"timeline,omitempty"`
}

type PlayerSummary struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Faction  string `json:"faction"`
	Color    string `json:"color"`
	Commands int    `json:"commands"`
}

func main() {
	server := &WebServer{
		dataDir: defaultDataDir,
		port:    8080,
	}

	// Parse command line arguments
	if len(os.Args) > 1 {
		if port, err := strconv.Atoi(os.Args[1]); err == nil {
			server.port = port
		}
	}

	server.setupRoutes()
	fmt.Printf("🚀 CoH3 Replay Analyzer Web Server starting on http://localhost:%d\n", server.port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", server.port), nil))
}

func (s *WebServer) setupRoutes() {
	// Serve static files
	http.HandleFunc("/", s.handleHome)
	http.HandleFunc("/upload", s.handleUpload)
	http.HandleFunc("/api/parse", s.handleParseReplay)
	
	// Static assets
	http.HandleFunc("/static/", s.handleStatic)
}

func (s *WebServer) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CoH3 Replay Analyzer</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            color: #333;
        }
        .container { 
            max-width: 95vw; 
            margin: 0 auto; 
            padding: 20px; 
        }
        .header {
            text-align: center;
            color: white;
            margin-bottom: 30px;
        }
        .header h1 {
            font-size: 2.5rem;
            margin-bottom: 10px;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
        }
        .header p {
            font-size: 1.1rem;
            opacity: 0.9;
        }
        .upload-area {
            background: white;
            border-radius: 15px;
            padding: 40px;
            text-align: center;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            margin-bottom: 30px;
            border: 3px dashed #ddd;
            transition: all 0.3s ease;
        }
        .upload-area:hover {
            border-color: #667eea;
            transform: translateY(-2px);
        }
        .upload-area.dragover {
            border-color: #667eea;
            background: #f8f9ff;
        }
        #file-input {
            display: none;
        }
        .upload-btn {
            background: #667eea;
            color: white;
            padding: 12px 30px;
            border: none;
            border-radius: 25px;
            font-size: 1rem;
            cursor: pointer;
            transition: all 0.3s ease;
        }
        .upload-btn:hover {
            background: #5a67d8;
            transform: translateY(-1px);
        }
        .results {
            background: white;
            border-radius: 15px;
            padding: 30px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            display: none;
        }
        .filters {
            background: white;
            border-radius: 15px;
            padding: 20px;
            margin-bottom: 20px;
            box-shadow: 0 5px 15px rgba(0,0,0,0.1);
        }
        .filter-row {
            display: flex;
            gap: 15px;
            align-items: center;
            flex-wrap: wrap;
            margin-bottom: 15px;
        }
        .filter-group {
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .filter-label {
            font-weight: 500;
            color: #333;
            white-space: nowrap;
        }
        select, input[type="range"] {
            padding: 8px 12px;
            border: 2px solid #ddd;
            border-radius: 6px;
            font-size: 0.9rem;
        }
        select:focus, input[type="range"]:focus {
            outline: none;
            border-color: #667eea;
        }
        .timeline-container {
            background: white;
            border-radius: 15px;
            padding: 20px;
            box-shadow: 0 5px 15px rgba(0,0,0,0.1);
            overflow-x: auto;
        }
        .timeline-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
        }
        .timeline-grid {
            display: flex;
            gap: 20px;
            min-width: 800px;
        }
        .timeline-columns {
            display: flex;
            gap: 15px;
            flex: 1;
        }
        .player-column {
            background: #f8f9ff;
            border-radius: 10px;
            padding: 15px;
            border-top: 4px solid;
            flex: 1;
            min-height: 800px;
            position: relative;
        }
        .player-column-header {
            text-align: center;
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 2px solid #eee;
        }
        .player-column-name {
            font-weight: bold;
            font-size: 1.1rem;
            margin-bottom: 5px;
        }
        .player-column-faction {
            color: #666;
            font-size: 0.9rem;
        }
        .timeline-content {
            position: relative;
            height: calc(100% - 60px);
        }
        .timeline-item {
            position: absolute;
            left: 0;
            right: 0;
            background: white;
            padding: 10px 15px;
            border-radius: 8px;
            border-left: 4px solid;
            box-shadow: 0 2px 6px rgba(0,0,0,0.1);
            transition: all 0.2s ease;
            min-height: 50px;
            display: flex;
            flex-direction: column;
            justify-content: center;
            margin-bottom: 8px;
        }
        .timeline-item:hover {
            transform: translateX(3px);
            box-shadow: 0 4px 8px rgba(0,0,0,0.15);
            z-index: 10;
        }
        .timeline-time {
            font-weight: bold;
            font-size: 0.9rem;
            color: #666;
            margin-bottom: 6px;
        }
        .timeline-command {
            font-size: 0.95rem;
            line-height: 1.4;
            word-wrap: break-word;
        }
        .time-markers {
            background: linear-gradient(135deg, #f8f9ff 0%, #e8eaff 100%);
            border-right: 3px solid #ddd;
            padding: 15px;
            min-width: 80px;
            border-radius: 10px 0 0 10px;
            position: relative;
        }
        .time-markers-header {
            text-align: center;
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 2px solid #ddd;
            font-weight: bold;
            color: #333;
        }
        .time-markers-content {
            position: relative;
            height: calc(100% - 60px);
        }
        .time-marker {
            position: absolute;
            left: 0;
            right: 0;
            height: 40px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 0.9rem;
            font-weight: bold;
            color: #666;
            border-bottom: 1px solid #ccc;
            background: rgba(255,255,255,0.7);
        }
        .time-marker.major {
            background: rgba(102, 126, 234, 0.1);
            border-bottom: 2px solid #667eea;
            color: #667eea;
            font-size: 1rem;
        }
        .loading {
            text-align: center;
            padding: 20px;
            display: none;
        }
        .spinner {
            border: 4px solid #f3f3f3;
            border-top: 4px solid #667eea;
            border-radius: 50%;
            width: 40px;
            height: 40px;
            animation: spin 1s linear infinite;
            margin: 0 auto 20px;
        }
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        .error {
            background: #fee;
            color: #c53030;
            padding: 15px;
            border-radius: 8px;
            margin: 20px 0;
            border-left: 4px solid #c53030;
        }
        .replay-info {
            background: linear-gradient(135deg, #f8f9ff 0%, #e8eaff 100%);
            border-radius: 15px;
            padding: 25px;
            margin-bottom: 30px;
            box-shadow: 0 5px 15px rgba(0,0,0,0.1);
        }
        .info-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            flex-wrap: wrap;
            gap: 20px;
        }
        .info-title {
            font-size: 1.5rem;
            font-weight: bold;
            color: #333;
        }
        .info-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 15px;
            margin-bottom: 20px;
        }
        .info-card {
            background: white;
            padding: 15px;
            border-radius: 10px;
            text-align: center;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .info-value {
            font-size: 1.3rem;
            font-weight: bold;
            color: #667eea;
            margin-bottom: 5px;
        }
        .info-label {
            color: #666;
            font-size: 0.85rem;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        .players-summary {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
        }
        .player-summary {
            background: white;
            padding: 15px;
            border-radius: 10px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            border-left: 4px solid;
        }
        .player-name {
            font-weight: bold;
            font-size: 1.1rem;
            margin-bottom: 5px;
        }
        .player-faction {
            color: #666;
            font-size: 0.9rem;
            margin-bottom: 10px;
        }
        .player-stats {
            font-size: 0.85rem;
            color: #888;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎮 CoH3 Replay Analyzer</h1>
            <p>Upload a Company of Heroes 3 replay file (.rec) to analyze build orders and timelines</p>
        </div>

        <div class="upload-area" id="upload-area">
            <h3>📁 Upload Replay File</h3>
            <p style="margin: 15px 0; color: #666;">Drag and drop a .rec file here, or click to browse</p>
            <input type="file" id="file-input" accept=".rec" />
            <button class="upload-btn" onclick="document.getElementById('file-input').click()">
                Choose File
            </button>
        </div>

        <div class="loading" id="loading">
            <div class="spinner"></div>
            <p>Parsing replay file...</p>
        </div>

        <div class="results" id="results">
            <div class="replay-info" id="replay-info"></div>
            
            <div class="filters">
                <div class="filter-row">
                    <div class="filter-group">
                        <label class="filter-label">Command Type:</label>
                        <select id="command-filter">
                            <option value="all">All Commands</option>
                            <option value="build_squad">Units Only</option>
                            <option value="construct_entity">Buildings Only</option>
                            <option value="build_global_upgrade">Upgrades Only</option>
                            <option value="select_battlegroup">Battlegroups Only</option>
                        </select>
                    </div>
                    <div class="filter-group">
                        <label class="filter-label">Player:</label>
                        <select id="player-filter">
                            <option value="all">All Players</option>
                        </select>
                    </div>
                    <div class="filter-group">
                        <label class="filter-label">Time Range:</label>
                        <input type="range" id="time-filter" min="0" max="100" value="100">
                        <span id="time-display">All</span>
                    </div>
                </div>
            </div>

            <div class="timeline-container">
                <div class="timeline-header">
                    <h3>📊 Build Order Timeline</h3>
                    <div id="visible-commands">Showing all commands</div>
                </div>
                <div class="timeline-grid">
                    <div class="timeline-columns" id="timeline-columns"></div>
                </div>
            </div>
        </div>
    </div>

    <script>
        const uploadArea = document.getElementById('upload-area');
        const fileInput = document.getElementById('file-input');
        const loading = document.getElementById('loading');
        const results = document.getElementById('results');

        // Drag and drop functionality
        uploadArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            uploadArea.classList.add('dragover');
        });

        uploadArea.addEventListener('dragleave', () => {
            uploadArea.classList.remove('dragover');
        });

        uploadArea.addEventListener('drop', (e) => {
            e.preventDefault();
            uploadArea.classList.remove('dragover');
            const files = e.dataTransfer.files;
            if (files.length > 0) {
                handleFile(files[0]);
            }
        });

        fileInput.addEventListener('change', (e) => {
            if (e.target.files.length > 0) {
                handleFile(e.target.files[0]);
            }
        });

        function handleFile(file) {
            if (!file.name.endsWith('.rec')) {
                alert('Please select a .rec replay file');
                return;
            }

            const formData = new FormData();
            formData.append('replay', file);

            loading.style.display = 'block';
            results.style.display = 'none';

            fetch('/api/parse', {
                method: 'POST',
                body: formData
            })
            .then(response => response.json())
            .then(data => {
                loading.style.display = 'none';
                if (data.success) {
                    displayResults(data);
                } else {
                    displayError(data.error);
                }
            })
            .catch(error => {
                loading.style.display = 'none';
                displayError('Error parsing replay: ' + error.message);
            });
        }

        let currentData = null;
        let maxTimestamp = 0;

        function displayResults(data) {
            currentData = data;
            maxTimestamp = Math.max(...data.timeline.map(e => e.timestamp));
            
            displayReplayInfo(data);
            setupFilters(data);
            displayTimeline(data.timeline);
            
            results.style.display = 'block';
        }

        function displayReplayInfo(data) {
            const replayInfo = document.getElementById('replay-info');
            
            // Basic info
            const infoGrid = 
                '<div class="info-header">' +
                    '<div class="info-title">📋 ' + data.map_name + '</div>' +
                '</div>' +
                '<div class="info-grid">' +
                    '<div class="info-card">' +
                        '<div class="info-value">' + data.duration + '</div>' +
                        '<div class="info-label">Duration</div>' +
                    '</div>' +
                    '<div class="info-card">' +
                        '<div class="info-value">' + data.players.length + '</div>' +
                        '<div class="info-label">Players</div>' +
                    '</div>' +
                    '<div class="info-card">' +
                        '<div class="info-value">' + data.timeline.length + '</div>' +
                        '<div class="info-label">Commands</div>' +
                    '</div>' +
                '</div>';

            // Players summary
            const playersSummary = 
                '<div class="players-summary">' +
                data.players.map(player => 
                    '<div class="player-summary" style="border-left-color: ' + player.color + '">' +
                        '<div class="player-name" style="color: ' + player.color + '">' + player.name + '</div>' +
                        '<div class="player-faction">' + player.faction + '</div>' +
                        '<div class="player-stats">' + player.commands + ' commands</div>' +
                    '</div>'
                ).join('') +
                '</div>';

            replayInfo.innerHTML = infoGrid + playersSummary;
        }

        function setupFilters(data) {
            // Populate player filter
            const playerFilter = document.getElementById('player-filter');
            const players = data.players;
            playerFilter.innerHTML = '<option value="all">All Players</option>' +
                players.map(player => '<option value="' + player.id + '">' + player.name + '</option>').join('');

            // Set up time filter
            const timeFilter = document.getElementById('time-filter');
            const timeDisplay = document.getElementById('time-display');
            
            timeFilter.addEventListener('input', function() {
                const percentage = this.value;
                const maxTime = maxTimestamp * (percentage / 100);
                const minutes = Math.floor(maxTime / 60000);
                const seconds = Math.floor((maxTime % 60000) / 1000);
                timeDisplay.textContent = percentage == 100 ? 'All' : 
                    String(minutes).padStart(2, '0') + ':' + String(seconds).padStart(2, '0');
                applyFilters();
            });

            // Set up other filters
            document.getElementById('command-filter').addEventListener('change', applyFilters);
            document.getElementById('player-filter').addEventListener('change', applyFilters);
        }

        function applyFilters() {
            const commandFilter = document.getElementById('command-filter').value;
            const playerFilter = document.getElementById('player-filter').value;
            const timePercentage = document.getElementById('time-filter').value;
            const maxTime = maxTimestamp * (timePercentage / 100);

            let filteredTimeline = currentData.timeline.filter(event => {
                if (commandFilter !== 'all' && event.command_type !== commandFilter) return false;
                if (playerFilter !== 'all' && event.player_id.toString() !== playerFilter) return false;
                if (event.timestamp > maxTime) return false;
                return true;
            });

            displayTimeline(filteredTimeline);
            
            const visibleCommands = document.getElementById('visible-commands');
            visibleCommands.textContent = 'Showing ' + filteredTimeline.length + ' of ' + currentData.timeline.length + ' commands';
        }

        function displayTimeline(timeline) {
            const timelineGrid = document.querySelector('.timeline-grid');
            
            // Calculate timeline parameters
            const timelineHeight = 800; // Increased height for better spacing
            const headerHeight = 60;
            const contentHeight = timelineHeight - headerHeight;
            
            // Group events by player
            const playerEvents = {};
            currentData.players.forEach(player => {
                playerEvents[player.id] = timeline.filter(event => event.player_id === player.id);
            });
            
            // Create a unified timeline that handles both time markers and events
            const unifiedTimeline = createUnifiedTimeline(playerEvents, contentHeight);
            
            // Create time markers based on unified timeline
            const timeMarkersHtml = createUnifiedTimeMarkers(unifiedTimeline, contentHeight);
            
            // Create player columns HTML from unified timeline
            const playerColumnsHtml = currentData.players.map(player => {
                return createUnifiedPlayerColumn(player, unifiedTimeline, contentHeight);
            }).join('');
            
            timelineGrid.innerHTML = timeMarkersHtml + '<div class="timeline-columns">' + playerColumnsHtml + '</div>';
        }

        function createUnifiedTimeline(playerEvents, contentHeight) {
            // Collect all events and time markers into a unified timeline
            const allItems = [];
            
            // Add time markers every 30 seconds
            const maxMinutes = Math.ceil(maxTimestamp / 60000);
            for (let seconds = 0; seconds <= maxMinutes * 60; seconds += 30) {
                const minutes = Math.floor(seconds / 60);
                const remainingSeconds = seconds % 60;
                const timeStr = String(minutes).padStart(2, '0') + ':' + String(remainingSeconds).padStart(2, '0');
                const timestamp = seconds * 1000;
                const isMajor = remainingSeconds === 0;
                
                allItems.push({
                    type: 'time_marker',
                    timestamp: timestamp,
                    timeStr: timeStr,
                    major: isMajor
                });
            }
            
            // Add all player events
            Object.keys(playerEvents).forEach(playerId => {
                playerEvents[playerId].forEach(event => {
                    allItems.push({
                        type: 'event',
                        timestamp: event.timestamp,
                        playerId: event.player_id,
                        event: event
                    });
                });
            });
            
            // Sort all items by timestamp
            allItems.sort((a, b) => a.timestamp - b.timestamp);
            
            // Assign positions with overlap prevention
            const itemHeight = 58;
            const positionedItems = [];
            let currentPosition = 0;
            
            allItems.forEach(item => {
                const idealPosition = (item.timestamp / maxTimestamp) * contentHeight;
                let actualPosition = Math.max(idealPosition, currentPosition);
                
                // For events, ensure minimum spacing
                if (item.type === 'event') {
                    actualPosition = Math.max(actualPosition, currentPosition);
                    currentPosition = actualPosition + itemHeight;
                }
                
                positionedItems.push({
                    ...item,
                    position: actualPosition
                });
            });
            
            return positionedItems;
        }

        function createUnifiedTimeMarkers(unifiedTimeline, contentHeight) {
            const timeMarkers = unifiedTimeline.filter(item => item.type === 'time_marker');
            
            const markersContent = timeMarkers.map(marker => 
                '<div class="time-marker' + (marker.major ? ' major' : '') + '" style="top: ' + marker.position + 'px">' +
                    marker.timeStr +
                '</div>'
            ).join('');
            
            return '<div class="time-markers">' +
                '<div class="time-markers-header">⏱️ Time</div>' +
                '<div class="time-markers-content">' + markersContent + '</div>' +
            '</div>';
        }

        function createUnifiedPlayerColumn(player, unifiedTimeline, contentHeight) {
            const playerEvents = unifiedTimeline.filter(item => 
                item.type === 'event' && item.playerId === player.id
            );
            
            const eventsHtml = playerEvents.map(item => {
                return '<div class="timeline-item" style="top: ' + item.position + 'px; border-left-color: ' + item.event.color + '">' +
                    '<div class="timeline-time">' + item.event.timestamp_str + '</div>' +
                    '<div class="timeline-command">' + item.event.description + '</div>' +
                '</div>';
            }).join('');
            
            return '<div class="player-column" style="border-top-color: ' + player.color + '; height: 800px">' +
                '<div class="player-column-header">' +
                    '<div class="player-column-name" style="color: ' + player.color + '">' + player.name + '</div>' +
                    '<div class="player-column-faction">' + player.faction + ' (' + playerEvents.length + ' commands)</div>' +
                '</div>' +
                '<div class="timeline-content">' + eventsHtml + '</div>' +
            '</div>';
        }

        function displayError(message) {
            results.innerHTML = '<div class="error">❌ ' + message + '</div>';
            results.style.display = 'block';
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(tmpl))
}

func (s *WebServer) handleUpload(w http.ResponseWriter, r *http.Request) {
	// This is just a fallback, main upload handling is in /api/parse
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *WebServer) handleParseReplay(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		s.sendJSONError(w, "Failed to parse upload: "+err.Error())
		return
	}

	file, handler, err := r.FormFile("replay")
	if err != nil {
		s.sendJSONError(w, "No replay file uploaded")
		return
	}
	defer file.Close()

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(handler.Filename), ".rec") {
		s.sendJSONError(w, "Please upload a .rec replay file")
		return
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "replay_*.rec")
	if err != nil {
		s.sendJSONError(w, "Failed to create temporary file")
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy uploaded file to temp file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		s.sendJSONError(w, "Failed to save uploaded file")
		return
	}

	// Parse the replay
	replayData, err := vault.ParseReplayWithLookup(tempFile.Name(), s.dataDir)
	if err != nil {
		s.sendJSONError(w, "Failed to parse replay: "+err.Error())
		return
	}

	// Convert to web response format
	response := s.convertToWebResponse(replayData)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *WebServer) handleStatic(w http.ResponseWriter, r *http.Request) {
	// For future static assets (CSS, JS files)
	http.NotFound(w, r)
}

func (s *WebServer) sendJSONError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(ReplayResponse{
		Success: false,
		Error:   message,
	})
}

func (s *WebServer) convertToWebResponse(replayData *vault.ReplayData) ReplayResponse {
	response := ReplayResponse{
		Success:  true,
		MapName:  replayData.MapName,
		Duration: formatDuration(replayData.DurationSeconds),
	}

	// Player colors for visualization
	colors := []string{
		"#e74c3c", "#3498db", "#2ecc71", "#f39c12", 
		"#9b59b6", "#1abc9c", "#e67e22", "#95a5a6",
	}

	// Create player summaries
	playerSummaries := make([]PlayerSummary, 0, len(replayData.Players))
	for i, player := range replayData.Players {
		faction := "Unknown"
		if player.Faction != nil {
			faction = *player.Faction
		}
		
		playerSummaries = append(playerSummaries, PlayerSummary{
			ID:       int(player.PlayerID),
			Name:     player.PlayerName,
			Faction:  faction,
			Color:    colors[i%len(colors)],
			Commands: len(player.BuildCommands),
		})
	}
	response.Players = playerSummaries

	// Create timeline events
	timeline := make([]TimelineEvent, 0)
	
	for i, player := range replayData.Players {
		color := colors[i%len(colors)]
		faction := "Unknown"
		if player.Faction != nil {
			faction = *player.Faction
		}

		for _, cmd := range player.BuildCommands {
			description := s.formatCommandDescription(cmd)
			
			timeline = append(timeline, TimelineEvent{
				PlayerID:     int(player.PlayerID),
				PlayerName:   player.PlayerName,
				Faction:      faction,
				Timestamp:    cmd.Timestamp,
				TimestampStr: formatTimestamp(cmd.Timestamp),
				CommandType:  cmd.CommandType,
				Description:  description,
				Color:        color,
			})
		}
	}

	// Sort timeline by timestamp
	for i := 0; i < len(timeline)-1; i++ {
		for j := i + 1; j < len(timeline); j++ {
			if timeline[i].Timestamp > timeline[j].Timestamp {
				timeline[i], timeline[j] = timeline[j], timeline[i]
			}
		}
	}

	response.Timeline = timeline
	return response
}

func (s *WebServer) formatCommandDescription(cmd vault.Command) string {
	switch cmd.CommandType {
	case "build_squad":
		if cmd.UnitName != nil {
			return fmt.Sprintf("🪖 Built: %s", *cmd.UnitName)
		}
		return "🪖 Built unit"
		
	case "construct_entity":
		if cmd.BuildingName != nil {
			return fmt.Sprintf("🏗️ Constructed: %s", *cmd.BuildingName)
		}
		return "🏗️ Constructed building"
		
	case "build_global_upgrade":
		if cmd.UnitName != nil {
			return fmt.Sprintf("🔬 Researched: %s", *cmd.UnitName)
		}
		return "🔬 Researched upgrade"
		
	case "select_battlegroup":
		if cmd.UnitName != nil {
			return fmt.Sprintf("⚔️ Selected: %s", *cmd.UnitName)
		}
		return "⚔️ Selected battlegroup"
		
	default:
		return fmt.Sprintf("📋 %s", cmd.CommandType)
	}
}

func formatTimestamp(ms uint32) string {
	seconds := ms / 1000
	minutes := seconds / 60
	seconds = seconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func formatDuration(seconds uint32) string {
	minutes := seconds / 60
	seconds = seconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}