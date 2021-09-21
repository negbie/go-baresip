package main

import (
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"os"

	gobaresip "github.com/negbie/go-baresip"
)

const v = "v0.1.2"

func main() {
	debug := flag.Bool("debug", false, "Set debug mode")
	ctrlAddr := flag.String("ctrl_address", "127.0.0.1:4444", "Local control listen address")
	guiAddr := flag.String("gui_address", "0.0.0.0:8080", "Local GUI listen address")
	logStd := flag.Bool("log_stderr", true, "Log to stderr")
	lokiURL := flag.String("loki_url", "", "URL to remote Loki server like http://localhost:3100")
	maxCalls := flag.Uint("max_calls", 30, "Maximum number of incoming calls")
	rtpNet := flag.String("rtp_interface", "", "RTP interface like eth0")
	rtpPorts := flag.String("rtp_ports", "10000-11000", "RTP port range")
	rtpTimeout := flag.Uint("rtp_timeout", 10, "Seconds after which a call with no incoming RTP packets will be terminated")
	sipAddr := flag.String("sip_address", "", "SIP listen address like 0.0.0.0:5060")
	smtpHost := flag.String("smtp_host", "", "SMTP Host")
	smtpUser := flag.String("smtp_user", "", "SMTP User")
	smtpPass := flag.String("smtp_pass", "", "SMTP Password")
	smtpFrom := flag.String("smtp_from", "", "SMTP From")
	smtpTo := flag.String("smtp_to", "", "SMTP To")
	smtpSubject := flag.String("smtp_subject", "", "SMTP Subject")
	version := flag.Bool("version", false, "Print version")
	webhookDelay := flag.Uint("webhook_delay", 600, "Webhook resend delay of warnings and errors in seconds")
	webhookURL := flag.String("webhook_url", "", "Send warnings and errors to this Mattermost or Slack webhook URL")
	flag.Parse()

	if *version {
		fmt.Println(v)
		os.Exit(0)
	}

	createConfig(*maxCalls, *rtpNet, *rtpPorts, *rtpTimeout, *ctrlAddr, *sipAddr)

	if *webhookDelay == 0 {
		*webhookDelay = 1
	}

	var loki *LokiClient
	var err error

	if *lokiURL != "" {
		loki, err = NewLokiClient(*lokiURL, 20, 4)
		if err != nil {
			log.Fatal(err)
		}
		defer loki.Close()
	}

	var lokiELabels = map[string]string{
		"job":   "go-baresip",
		"level": "info",
	}
	var lokiRLabels = map[string]string{
		"job":   "go-baresip",
		"level": "info",
	}

	gb, err := gobaresip.New(
		gobaresip.SetAudioPath("sounds"),
		gobaresip.SetConfigPath("."),
		gobaresip.SetDebug(*debug),
		gobaresip.SetUserAgent("telefonist"),
		gobaresip.SetCtrlTCPAddr(*ctrlAddr),
		gobaresip.SetWsAddr(*guiAddr),
	)
	if err != nil {
		log.Fatal(err)
	}

	eChan := gb.GetEventChan()
	rChan := gb.GetResponseChan()

	go func() {
		for {
			select {
			case e, ok := <-eChan:
				if !ok {
					return
				}

				level := eventLevel(&e)

				if *lokiURL != "" {
					lokiELabels["level"] = level
					loki.Send(lokiELabels, string(e.RawJSON))
				}
				if *webhookURL != "" && (level == "warning" || level == "error") {
					go func() {
						if err := page(*webhookURL, level, string(e.RawJSON)); err != nil {
							log.Println(err)
						}
					}()
				}
				if *smtpHost != "" && (level == "warning" || level == "error") {
					go func() {
						body := string(e.RawJSON)
						msg := "From: " + *smtpFrom + "\n" + "To: " + *smtpTo + "\n" + "Subject: " + *smtpSubject + ":" + e.PeerURI + "\n\n" + body
						err := smtp.SendMail(*smtpHost+":587", smtp.PlainAuth("", *smtpUser, *smtpPass, *smtpHost), *smtpFrom, []string{*smtpTo}, []byte(msg))
						if err != nil {
							log.Println(err)
						}
					}()
				}
				if *logStd {
					log.Println(level, ":", string(e.RawJSON))
				}
			case r, ok := <-rChan:
				if !ok {
					return
				}

				if *lokiURL != "" {
					loki.Send(lokiRLabels, string(r.RawJSON))
				}
				if *logStd {
					log.Println("info", ":", string(r.RawJSON))
				}
			}
		}
	}()

	err = gb.Run()
	if err != nil {
		log.Println(err)
	}
	gb.Close()
}
