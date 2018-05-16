package main

import (
	"github.com/flashmob/go-guerrilla/backends"
	"github.com/flashmob/go-guerrilla/mail"
)

type OpsGenieConfig struct {
	LogReceivedMails bool `json:"log_received_mails"`
}

func OpsGenieProcessor() backends.Decorator {
	var config *OpsGenieConfig
	initFunc := backends.InitializeWith(func(backendConfig backends.BackendConfig) error {
		configType := backends.BaseConfig(&OpsGenieConfig{})
		bcfg, err := backends.Svc.ExtractConfig(backendConfig, configType)
		if err != nil {
			return err
		}
		config = bcfg.(*OpsGenieConfig)
		return nil
	})
	backends.Svc.AddInitializer(initFunc)
	return func(p backends.Processor) backends.Processor {
		return backends.ProcessWith(func(e *mail.Envelope, t backends.SelectTask) (backends.Result, error) {
			if t == backends.TaskSaveMail {
				if config.LogReceivedMails {
					mainlog.Infof("[OpsGenie] Mail from: %s / to: %v", e.MailFrom.String(), e.RcptTo)
					mainlog.Info("[OpsGenie] Headers are:", e.Header)
					mainlog.Infof("[OpsGenie] Email is: %v", e.String())
					mainlog.Info("[OpsGenie] TODO: POST the above to OpsGenie")
				}
				return p.Process(e, t)
			}
			return p.Process(e, t)
		})
	}
}
