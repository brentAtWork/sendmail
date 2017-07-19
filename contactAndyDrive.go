//contactAndyDrive

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"text/template"
)

type configuration struct {
	SMTPUser     string
	SMTPPass     string
	Reciever     string
	RecieverName string
	SiteName     string
}

type emailInfo struct {
	Sender    string `json:"sender"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
	Recaptcha string `json:"g-recaptcha-response"`
}

type smtpMessage struct {
	From    string
	To      string
	Subject string
	Body    string
}

const smtpMessageTemplate = `From: SendMail <{{.From}}>
To: {{.To}}
Subject: {{.Subject}}

{{.Body}}
`
const emailMessageTemplate = `A New Message From SendMail
 
From: {{.Sender}}
Subject: {{.Subject}}
Message: {{.Message}}
`

var config = loadConfiguration()
var smtpAuth = smtp.PlainAuth(
	"",
	config.SMTPUser,
	config.SMTPPass,
	"smtp.gmail.com",
)

var err error

func main() {

	fmt.Println("Send mail up and running!")
	http.HandleFunc("/sendMail", sendMail)
	http.ListenAndServe(":5678", nil)
}

func sendMail(w http.ResponseWriter, r *http.Request) {
	requestBody := decodeRequestBody(r.Body)
	if err = verifyRequest(requestBody.Recaptcha, r.RemoteAddr); err != nil {
		w.WriteHeader(http.StatusPreconditionFailed)
		return
	}
	message := byteBufferFromTemplateString(requestBody, emailMessageTemplate)
	outgoingMessage := byteBufferFromTemplateString(&smtpMessage{
		config.SMTPUser,
		config.Reciever,
		"A Message from Sendmail at " + config.SiteName,
		message.String(),
	}, smtpMessageTemplate)

	err = smtp.SendMail(
		"smtp.gmail.com:587",
		smtpAuth,
		config.SMTPUser,
		[]string{config.Reciever},
		outgoingMessage.Bytes(),
	)

	if err != nil {
		fmt.Println("Error: ", err)
		w.WriteHeader(504)
		w.Write([]byte("Something went wrong. Try agian."))
	} else {
		fmt.Println("Mail sent!")
		w.WriteHeader(200)
		w.Write([]byte("Mail sent"))
	}
	return
}

func verifyRequest(recaptcha string, remoteIP string) error {
	fmt.Println("Recaptcha: ", recaptcha)
	fmt.Println("Remote IP: ", remoteIP)
	return nil
}

func byteBufferFromTemplateString(info interface{}, temp string) bytes.Buffer {
	var byteBuffer bytes.Buffer
	t := template.New("template")
	t, err = t.Parse(temp)
	if err != nil {
		panic(err)
	}
	err = t.Execute(&byteBuffer, info)
	if err != nil {
		panic(err)
	}
	return byteBuffer
}

func loadConfiguration() configuration {
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	config := configuration{}
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error loading configs: ", err)
	}
	return config
}

func decodeRequestBody(body io.ReadCloser) emailInfo {
	decoder := json.NewDecoder(body)
	var decodedBody emailInfo
	err := decoder.Decode(&decodedBody)
	if err != nil {
		panic(err)
	}
	defer body.Close()
	return decodedBody
}
