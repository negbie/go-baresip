package gobaresip

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *WsHub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.hub.command <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *WsHub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

// WsHub maintains the set of active clients and broadcasts events to the
// clients.
type WsHub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound command from the clients.
	command chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	bs *Baresip
}

func newWsHub(bs *Baresip) *WsHub {
	return &WsHub{
		clients:    make(map[*Client]bool),
		command:    make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		bs:         bs,
	}
}

func (h *WsHub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case msg := <-h.command:
			cm := string(msg)
			cp := strings.Split(cm, " ")
			if len(cp) == 1 {
				h.bs.Cmd(cp[0], "", "command_"+cp[0])
			} else if len(cp) == 2 {
				h.bs.Cmd(cp[0], cp[1], "command_"+cp[0])
			} else if len(cp) > 2 {
				h.bs.Cmd(cm, "", "command_too_long")
			}
		case e, ok := <-h.bs.eventWsChan:
			if !ok {
				continue
			}
			for client := range h.clients {
				select {
				case client.send <- prefixTime(time.Now().Format("2006-01-02 15:04:05 "), e):
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		case r, ok := <-h.bs.responseWsChan:
			if !ok {
				continue
			}
			for client := range h.clients {
				select {
				case client.send <- prefixTime(time.Now().Format("2006-01-02 15:04:05 "), r):
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func prefixTime(prefix string, suffix []byte) []byte {
	b := make([]byte, len(prefix)+len(suffix))
	n := copy(b, prefix)
	copy(b[n:], suffix)
	return b
}

func serveRoot(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/ws")
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script type="text/javascript">
window.onload = function () {
    var conn;
    var msg = document.getElementById("msg");
    var log = document.getElementById("log");

    function appendLog(item) {
        var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
        log.appendChild(item);
        if (doScroll) {
            log.scrollTop = log.scrollHeight - log.clientHeight;
        }
    }

    document.getElementById("form").onsubmit = function () {
        if (!conn) {
            return false;
        }
        if (!msg.value) {
            return false;
        }
        conn.send(msg.value);
        msg.value = "";
        return false;
    };

    if (window["WebSocket"]) {
        conn = new WebSocket("{{.}}");
        conn.onclose = function (evt) {
            var item = document.createElement("div");
            item.innerHTML = "<b>Connection closed. Please reload.</b>";
            appendLog(item);
        };
        conn.onmessage = function (evt) {
            var messages = evt.data.split('\n');
            for (var i = 0; i < messages.length; i++) {
                var item = document.createElement("div");
                item.innerText = messages[i];
                appendLog(item);
            }
        };
    } else {
        var item = document.createElement("div");
        item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
        appendLog(item);
    }
};
</script>

</head>
<body>
<table border="1">
<tr><td valign="top" width="20%">
<form id="form">
    <input type="text" id="msg" size="35" autofocus />
    <input type="submit" value="Send" />
</form>

<div id="cmd" style="line-height: 0.7;max-height: 90vh;overflow-y: scroll;">
<pre>
<p>accept                Accept incoming call
<p>acceptdir ..          Accept incoming call with direction.
<p>answermode ..         Set answer mode
<p>aubitrate ..          Set audio bitrate
<p>audio_debug           Audio stream
<p>ausrc ..              Switch audio source
<p>autodial ..           Set auto dial command
<p>autodialcancel        Cancel auto dial
<p>autodialdelay ..      Set delay before auto dial [ms]
<p>autohangup ..         Set auto hangup command
<p>autohangupcancel      Cancel auto hangup
<p>autohangupdelay ..    Set delay before hangup [ms]
<p>autostat              Print autotest status
<p>callfind ..           Find call
<p>callstat              Call status
<p>contact_next          Set next contact
<p>contact_prev          Set previous contact
<p>contacts              List contacts
<p>dial ..               Dial
<p>dialcontact           Dial current contact
<p>dialdir ..            Dial with audio and videodirection.
<p>dnd ..                Set Do not Disturb
<p>hangup                Hangup call
<p>hangupall ..          Hangup all calls with direction
<p>hold                  Call hold
<p>line ..               Set current call
<p>listcalls             List active calls
<p>medialdir ..          Set local media direction
<p>mute                  Call mute/un-mute
<p>reginfo               Registration info
<p>reinvite              Send re-INVITE
<p>resume                Call resume
<p>setadelay ..          Set answer delay for outgoing call
<p>sndcode ..            Send Code
<p>transfer ..           Transfer call
<p>uadel ..              Delete User-Agent
<p>uafind ..             Find User-Agent
<p>uanew ..              Create User-Agent
<p>uareg ..              UA register [index]
</pre>
</div>
</td><td valign="top" width="80%">
<div id="log" style="line-height: 1.7;max-height: 90vh;overflow-y: scroll;"></div>
</td></tr></table>
</body>
</html>
`))
