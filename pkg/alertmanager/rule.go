package alertmanager

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	alertmanagerv1 "github.com/linclaus/alertmanager-operator/api/v1"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func AddAlertmanagerStrategy(rule alertmanagerv1.AlertmanagerRule) error {
	cfg, err := loadFromFile(viper.GetString("configName"))
	if err != nil {
		fmt.Printf("Error load alertmanager config for reason:%s\n", err)
		return err
	}
	contactValues := []string{}
	for _, ec := range rule.Spec.Receiver.EmailConfigs {
		contactValues = append(contactValues, ec.To)
	}
	cfg.Receivers = updateReceivers(cfg.Receivers, rule.Spec.Receiver.Name, contactValues)
	cfg.Route.Routes = updateRoutes(cfg.Route.Routes, rule.Spec.Route)
	err = writeConfigToFile(*cfg)
	return err
}

func DeleteAlertmanagerStrategy(rule alertmanagerv1.AlertmanagerRule) error {
	cfg, err := loadFromFile(viper.GetString("configName"))
	if err != nil {
		fmt.Printf("Error load alertmanager config for reason:%s\n", err)
		return err
	}
	contactValues := []string{}
	for _, ec := range rule.Spec.Receiver.EmailConfigs {
		contactValues = append(contactValues, ec.To)
	}
	rvs, deleteRoute := deleteReceivers(cfg.Receivers, rule.Spec.Receiver.Name, contactValues, false)
	cfg.Receivers = rvs
	if deleteRoute {
		cfg.Route.Routes = deleteRoutes(cfg.Route.Routes, rule.Spec.Route.Receiver)
	}
	fmt.Println(cfg)
	err = writeConfigToFile(*cfg)
	return err
}

func writeConfigToFile(cfg Config) error {
	file, err := os.OpenFile(viper.GetString("configPath")+viper.GetString("configName"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer file.Close()
	if err != nil {
		fmt.Println("文件打开失败", err)
		return err
	}
	write := bufio.NewWriter(file)
	write.WriteString(cfg.String())
	write.Flush()
	reloadAlertmanager()
	return nil
}

func reloadAlertmanager() {
	http.Post(viper.GetString("alertmanager.host")+"/-/reload", "application/json", nil)
}

func loadFromFile(fileName string) (*Config, error) {
	content, err := ioutil.ReadFile(viper.GetString("configPath") + fileName)
	if err != nil {
		return nil, err
	}
	cfg, err := load(string(content))
	return cfg, err
}

func load(s string) (*Config, error) {
	cfg := &Config{}
	err := yaml.UnmarshalStrict([]byte(s), cfg)
	if err != nil {
		return nil, err
	}
	// Check if we have a root route. We cannot check for it in the
	// UnmarshalYAML method because it won't be called if the input is empty
	// (e.g. the config file is empty or only contains whitespace).
	if cfg.Route == nil {
		return nil, errors.New("no route provided in config")
	}

	// Check if continue in root route.
	if cfg.Route.Continue {
		return nil, errors.New("cannot have continue in root route")
	}

	cfg.original = s
	return cfg, nil
}
