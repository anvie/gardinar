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

type TaskRunner func(taskScript, sourceDir string, params []string) (string, error)

type webhookResponse struct {
	Result     string `json:"result,omitempty"`
	State      string `json:"state,omitempty"`
	FailedStep string `json:"failed_step,omitempty"`
	Error      string `json:"error,omitempty"`
}

func NewWebhookHandler(cfg *Config, run TaskRunner) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.TrimSpace(r.Header.Get("X-SECRET-KEY")) != strings.TrimSpace(cfg.SecretKey) {
			writeJSON(w, http.StatusUnauthorized, webhookResponse{Error: "unauthorized"})
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, webhookResponse{Error: "invalid request body"})
			return
		}

		var webhook Webhook
		if err := json.Unmarshal(body, &webhook); err != nil {
			writeJSON(w, http.StatusBadRequest, webhookResponse{Error: "invalid webhook payload"})
			return
		}

		log.Printf("Received webhook for task '%s' with params: %v", webhook.Task, webhook.Params)
		taskScript, ok := cfg.Tasks[webhook.Task]
		if !ok {
			writeJSON(w, http.StatusBadRequest, webhookResponse{Error: "task not found"})
			return
		}

		if _, err := os.Stat(taskScript); os.IsNotExist(err) {
			log.Printf("Task script '%s' not found, attempting to run as shell command.", taskScript)
		}
		if _, err := run(taskScript, webhook.SourceDir, webhook.Params); err != nil {
			log.Printf("Task '%s' failed: %v", webhook.Task, err)
			writeJSON(w, http.StatusInternalServerError, webhookResponse{State: "failed", FailedStep: "task", Error: "task execution failed"})
			return
		}

		writeJSON(w, http.StatusOK, webhookResponse{Result: "success", State: "succeeded"})
	})
}

func writeJSON(w http.ResponseWriter, status int, response webhookResponse) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error writing response: %v", err)
	}
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

	http.Handle("/webhook", NewWebhookHandler(cfg, runTask))

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
