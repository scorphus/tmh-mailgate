package main

import (
	"bytes"
	"html/template"

	"github.com/flashmob/go-guerrilla/backends"
	"github.com/flashmob/go-guerrilla/mail"
)

type (
	OpsGenieConfig struct {
		APIURL string `json:"opsgenie_api_url"`
		APIKey string `json:"opsgenie_api_key"`
	}
	OpsGenieAlert struct {
		Message      string
		Alias        string
		Descripttion string
		Entity       []mail.Address
		Source       string
	}
)

var OpsGenieAlertTpl = template.Must(template.New("OpsGenieAlert").Parse(
	`{
		"message": "{{.Subject}}",
		"alias": "{{.Subject}}",
		"description":"{{.String}}",
		"tags": ["MailGate"],
		"entity": "{{.RcptTo}}",
		"source": "{{.MailFrom.String}}"
	}`,
))

func OpsGenieProcessor() backends.Decorator {
	var config *OpsGenieConfig
	initFunc := backends.InitializeWith(func(backendConfig backends.BackendConfig) error {
		configType := backends.BaseConfig(&OpsGenieConfig{})
		bcfg, err := backends.Svc.ExtractConfig(backendConfig, configType)
		if err != nil {
			return err
		}
		config = bcfg.(*OpsGenieConfig)
		mainlog.Infof("[OpsGenieProcessor] Config: %v", config)
		return nil
	})
	backends.Svc.AddInitializer(initFunc)
	return func(p backends.Processor) backends.Processor {
		return backends.ProcessWith(func(e *mail.Envelope, t backends.SelectTask) (backends.Result, error) {
			if t == backends.TaskSaveMail {
				mainlog.Info("[OpsGenieProcessor] Sending alert to OpsGenie")
				_, err := execOpsGenieAlertTpl(e)
				if err != nil {
					mainlog.Error("[OpsGenieProcessor] Failed to execute alert template")
					return p.Process(e, t)
				}
				mainlog.Infof("[OpsGenieProcessor] POST 201 OK: %s", e.Subject)
			}
			return p.Process(e, t)
		})
	}
}

func execOpsGenieAlertTpl(e *mail.Envelope) (string, error) {
	var b bytes.Buffer
	err := OpsGenieAlertTpl.Execute(&b, e)
	return b.String(), err
}
