const menuEl = document.getElementById("menu");
const lobbyEl = document.getElementById("lobby");
const gameEl = document.getElementById("game");

const nameInput = document.getElementById("player-name");
const hostBtn = document.getElementById("host-btn");
const joinBtn = document.getElementById("join-btn");
const menuError = document.getElementById("menu-error");

const playerListEl = document.getElementById("player-list");
const lobbyStatusEl = document.getElementById("lobby-status");
const startBtn = document.getElementById("start-btn");
const leaveBtn = document.getElementById("leave-btn");

const canvas = document.getElementById("viewport");
const ctx = canvas.getContext("2d");
const statusEl = document.getElementById("status");
const roleEl = document.getElementById("role");

let socket = null;
let entities = [];
let isHost = false;
let leavingIntentionally = false;
let screen = "menu";

function resize() {
  canvas.width = window.innerWidth;
  canvas.height = window.innerHeight;
}
window.addEventListener("resize", resize);
resize();

hostBtn.addEventListener("click", () => connect("host"));
joinBtn.addEventListener("click", () => connect("join"));
startBtn.addEventListener("click", () => send("start_game", {}));
leaveBtn.addEventListener("click", leave);

function setMenuBusy(busy) {
  hostBtn.disabled = busy;
  joinBtn.disabled = busy;
}

function showScreen(next) {
  screen = next;
  menuEl.classList.toggle("hidden", next !== "menu");
  lobbyEl.classList.toggle("hidden", next !== "lobby");
  gameEl.classList.toggle("hidden", next !== "game");
}

function connect(role) {
  menuError.textContent = "";
  setMenuBusy(true);
  isHost = role === "host";
  leavingIntentionally = false;

  const name = nameInput.value.trim() || "Player";
  socket = new WebSocket(`ws://${location.host}/ws`);

  socket.addEventListener("open", () => {
    send("join", { name, role });
    enterLobby();
  });

  socket.addEventListener("error", () => {
    menuError.textContent = "Could not reach the game server.";
    setMenuBusy(false);
  });

  socket.addEventListener("close", () => {
    if (leavingIntentionally) return;
    if (screen !== "menu") {
      menuError.textContent = "Disconnected from server.";
    }
    resetToMenu();
  });

  socket.addEventListener("message", handleMessage);
}

function enterLobby() {
  showScreen("lobby");
  startBtn.classList.toggle("hidden", !isHost);
  lobbyStatusEl.textContent = isHost
    ? "You are the host. Start whenever you're ready."
    : "Waiting for the host to start...";
}

function leave() {
  leavingIntentionally = true;
  if (socket) socket.close();
  resetToMenu();
}

function resetToMenu() {
  socket = null;
  entities = [];
  playerListEl.innerHTML = "";
  showScreen("menu");
  setMenuBusy(false);
}

function handleMessage(event) {
  const envelope = JSON.parse(event.data);
  switch (envelope.type) {
    case "lobby_update":
      renderRoster(envelope.payload.players || []);
      break;
    case "game_start":
      showScreen("game");
      statusEl.textContent = "connected";
      roleEl.textContent = isHost ? "Hosting" : "Playing";
      requestAnimationFrame(draw);
      break;
    case "state_update":
      entities = envelope.payload.entities;
      break;
    case "welcome":
      console.log("welcome", envelope.payload);
      break;
    default:
      console.log("unhandled message", envelope);
  }
}

function renderRoster(players) {
  playerListEl.innerHTML = "";
  for (const p of players) {
    const li = document.createElement("li");
    li.textContent = p.name + (p.isHost ? " (host)" : "");
    playerListEl.appendChild(li);
  }
}

function send(type, payload) {
  socket.send(JSON.stringify({ type, payload }));
}

function draw() {
  if (screen !== "game") return;

  ctx.fillStyle = "#12202b";
  ctx.fillRect(0, 0, canvas.width, canvas.height);

  for (const e of entities) {
    ctx.beginPath();
    ctx.fillStyle = e.faction === "federation" ? "#4fa3ff" : "#ff5c5c";
    ctx.arc(e.x, e.y, 5, 0, Math.PI * 2);
    ctx.fill();
  }

  requestAnimationFrame(draw);
}
