package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenPort string            `yaml:"listen_port"`
	SecretKey  string            `yaml:"secret_key"`
	Tasks      map[string]string `yaml:"tasks"`
}

func NewConfig(configPath string) (*Config, error) {
	config := &Config{}
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}

func ParseFlags() (string, error) {
	var configPath string
	flag.StringVar(&configPath, "config", "./config.yaml", "path to config file")
	flag.Parse()
	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}
	return configPath, nil
}

func main() {
	fmt.Println("Gardinar v0.2.0")
	cfgPath, err := ParseFlags()
	if err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	cfg, err := NewConfig(cfgPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	log.Printf("Loaded %d tasks from config file", len(cfg.Tasks))

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		clientSecretKey := strings.TrimSpace(r.Header.Get("X-SECRET-KEY"))
		secretKey := strings.TrimSpace(cfg.SecretKey)

		w.Header().Set("Content-Type", "application/json")

		if clientSecretKey != secretKey {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"result":"unauthorized"}`))
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, `{"error":"Error reading body"}`, http.StatusBadRequest)
			return
		}

		var webhook Webhook
		if err := json.Unmarshal(body, &webhook); err != nil {
			http.Error(w, `{"error":"Error decoding webhook"}`, http.StatusBadRequest)
			return
		}

		log.Printf("Received webhook for task '%s' with params: %v", webhook.Task, webhook.Params)

		taskScript, ok := cfg.Tasks[webhook.Task]
		if !ok {
			response := fmt.Sprintf(`{"error":"task '%s' not found in config"}`, webhook.Task)
			http.Error(w, response, http.StatusBadRequest)
			return
		}

		if _, err := os.Stat(taskScript); os.IsNotExist(err) {
			log.Printf("Task script '%s' not found, attempting to run as shell command.", taskScript)
		}

		out, err := runTask(taskScript, webhook.SourceDir, webhook.Params)
		if err != nil {
			response := fmt.Sprintf(`{"error":"%s", "details":"%s"}`, out, err.Error())
			http.Error(w, response, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"success"}`))
	})

	log.Printf("Server is running on port %s", cfg.ListenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.ListenPort), nil))
}

func runTask(taskScript string, sourceDir string, params []string) (string, error) {
	log.Printf("Executing task '%s' in directory '%s' with params %v", taskScript, sourceDir, params)

	cmd := exec.Command("/bin/bash", "-c", taskScript+" "+strings.Join(params, " "))
	if sourceDir != "" {
		cmd.Dir = sourceDir
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error executing task: %v\nOutput: %s", err, string(out))
		return string(out), err
	}

	log.Printf("Task executed successfully. Output:\n%s", string(out))
	return string(out), nil
}

type Webhook struct {
	Task      string   `json:"task"`
	SourceDir string   `json:"source_dir"`
	Params    []string `json:"params"`
}

