package gobaresip

import (
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (b *Baresip) wsCtrl(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close()

	done := make(chan struct{})
	go func(done chan struct{}) {
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				break
			}
			if mt == websocket.TextMessage {
				cmdString := string(message)
				cmdParts := strings.Split(cmdString, " ")
				if len(cmdParts) == 1 {
					b.Cmd(cmdParts[0], "", "command_"+cmdParts[0])
				} else if len(cmdParts) == 2 {
					b.Cmd(cmdParts[0], cmdParts[1], "command_"+cmdParts[0])
				}
			}
		}
		done <- struct{}{}
	}(done)

	for {
		select {
		case <-done:
			return
		case e, ok := <-b.eventWsChan:
			if !ok {
				return
			}
			tn := time.Now()
			tnf := tn.Format("2006-01-02 15:04:05 ")

			err = c.WriteMessage(websocket.TextMessage, []byte(tnf+e.Raw))
			if err != nil {
				log.Println(err)
				return
			}

		case r, ok := <-b.responseWsChan:
			if !ok {
				return
			}
			tn := time.Now()
			tnf := tn.Format("2006-01-02 15:04:05 ")

			err = c.WriteMessage(websocket.TextMessage, []byte(tnf+r.Raw))
			if err != nil {
				log.Println(err)
				return
			}
		}
	}

}

func (b *Baresip) home(w http.ResponseWriter, r *http.Request) {
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
            item.innerHTML = "<b>Connection closed.</b>";
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
<p>accept                  Accept incoming call
<p>acceptdir ..            Accept incoming call with direction.
<p>answermode ..           Set answer mode
<p>aubitrate ..            Set audio bitrate
<p>audio_debug             Audio stream
<p>ausrc ..                Switch audio source
<p>autodial ..             Set auto dial command
<p>autodialcancel          Cancel auto dial
<p>autodialdelay ..        Set delay before auto dial [ms]
<p>autohangup ..           Set auto hangup command
<p>autohangupcancel        Cancel auto hangup
<p>autohangupdelay ..      Set delay before hangup [ms]
<p>autostat                Print autotest status
<p>callfind ..             Find call
<p>callstat                Call status
<p>contact_next            Set next contact
<p>contact_prev            Set previous contact
<p>contacts                List contacts
<p>dial ..                 Dial
<p>dialcontact             Dial current contact
<p>dialdir ..              Dial with audio and videodirection.
<p>dnd ..                  Set Do not Disturb
<p>hangup                  Hangup call
<p>hangupall ..            Hangup all calls with direction
<p>hold                    Call hold
<p>line ..                 Set current call
<p>listcalls               List active calls
<p>medialdir ..            Set local media direction
<p>mute                    Call mute/un-mute
<p>reginfo                 Registration info
<p>reinvite                Send re-INVITE
<p>resume                  Call resume
<p>setadelay ..            Set answer delay for outgoing call
<p>sndcode ..              Send Code
<p>transfer ..             Transfer call
<p>uadel ..                Delete User-Agent
<p>uafind ..               Find User-Agent
<p>uanew ..                Create User-Agent
<p>uareg ..                UA register [index]
</pre>
</div>
</td><td valign="top" width="80%">
<div id="log" style="line-height: 1.7;max-height: 90vh;overflow-y: scroll;"></div>
</td></tr></table>
</body>
</html>
`))
