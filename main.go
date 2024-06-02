package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenPort       string `yaml:"listen_port"`
	SecretKey        string `yaml:"secret_key"`
	GitCmd           string `yaml:"git_cmd"`
	PostUpdateScript string `yaml:"post_update_script"`
}

func NewConfig(configPath string) (*Config, error) {
	config := &Config{}
	yamlFile, err := ioutil.ReadFile(configPath)

	if err != nil {
		panic(err)
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

	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")
	flag.Parse()

	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	return configPath, nil
}

func main() {
	fmt.Println("Gardinar v0.0.6")
	listenPort := "8800"
	gitCmd := ""
	postUpdateScript := ""
	mySecretKey := ""

	cfgPath, err := ParseFlags()
	if err != nil {
		err = godotenv.Load()
		if err != nil {
			log.Fatal("Cannot loading .env file")
		}
		listenPort = os.Getenv("LISTEN_PORT")
		gitCmd = os.Getenv("GARDINAR_GIT_CMD")
		postUpdateScript = os.Getenv("GARDINAR_POST_UPDATE_SCRIPT")
		mySecretKey = os.Getenv("GARDINAR_SECRET_KEY")

		fmt.Println("Config file not found, using .env file")
	} else {
		fmt.Printf("cfgPath: %s\n", cfgPath)
		cfg, err := NewConfig(cfgPath)
		if err != nil {
			log.Fatal(err)
		}

		listenPort = cfg.ListenPort
		gitCmd = cfg.GitCmd
		postUpdateScript = cfg.PostUpdateScript
		mySecretKey = cfg.SecretKey
	}

	fmt.Println("GIT_CMD:", gitCmd)
	fmt.Println("SECRET_KEY:", mySecretKey)
	fmt.Println("POST_UPDATE_SCRIPT:", postUpdateScript)

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {

		// Get X-SECRET-KEY from header
		clientSecretKey := strings.TrimSpace(r.Header.Get("X-SECRET-KEY"))

		// Get SECRET_KEY from .env
		secretKey := strings.TrimSpace(mySecretKey)

		w.Header().Set("Content-Type", "application/json")

		// Check if X-SECRET-KEY is equal to SECRET_KEY
		if clientSecretKey != secretKey {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"result":"unauthorized"}`))
			os.Exit(1)
			return
		}

		var webhook Webhook
		err := json.NewDecoder(r.Body).Decode(&webhook)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			os.Exit(2)
			return
		}

		fmt.Printf("%+v\n", webhook)

		if webhook.SourceDir == "" {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte(`{"error":"source_dir is required"}`)); err != nil {
				log.Println(err)
			}
			os.Exit(3)
			return
		}

		// update git repo
		gitBranch := webhook.GitBranch
		if gitBranch != "" {
			out, err := gitUpdate(gitCmd, gitBranch, webhook.SourceDir)
			if err != nil {
				response := fmt.Sprintf(`{"error":"%s, %s"}`, out, err.Error())
				http.Error(w, response, http.StatusInternalServerError)
				log.Println(err)
				os.Exit(4)
				return
			}
		}

		// run post update script
		if postUpdateScript != "" {
			out, err := runPostUpdateScript(postUpdateScript, webhook.SourceDir, webhook.PostUpdateParams)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				if _, err := w.Write([]byte(`{"error":"` + out + ", " + err.Error() + `"}`)); err != nil {
					log.Println(err)
				}
				os.Exit(5)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"result":"success"}`)); err != nil {
			log.Println(err)
		}
	})

	fmt.Printf("Server is running on port %s\n", listenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", listenPort), nil))
}

func gitUpdate(gitCmd string, gitBranch string, sourceDir string) (string, error) {
	fmt.Printf("%s $ %s pull origin %s\n", sourceDir, gitCmd, gitBranch)

	cmd := exec.Command(gitCmd, "checkout", "-f", gitBranch)
	cmd.Dir = sourceDir

	if _, err := cmd.Output(); err != nil {
		log.Println("Error during git checkout.", err)
		return "", err
	}

	cmd = exec.Command(gitCmd, "pull", "origin", gitBranch)
	cmd.Dir = sourceDir

	out, err := cmd.Output()
	if err != nil {
		log.Println("Error during git pull.", err)
		return "", err
	}
	fmt.Println(string(out))
	return string(out), nil
}

func runPostUpdateScript(postUpdateScript string, sourceDir string, postUpdateParams []string) (string, error) {
	// check is postUpdateScript file exists
	if _, err := os.Stat(postUpdateScript); os.IsNotExist(err) {
		log.Println(err)
		return "", err
	}
	fmt.Printf("%s $ %s\n", sourceDir, postUpdateScript)
	cmd := exec.Command("/bin/bash", append([]string{postUpdateScript}, postUpdateParams...)...)
	cmd.Dir = sourceDir
	out, err := cmd.Output()
	if err != nil {
		log.Println(err)
		return "", err
	}
	fmt.Println(string(out))
	return string(out), nil
}

type Webhook struct {
	Version          string   `json:"version"`
	CommitHash       string   `json:"commit_hash"`
	GitBranch        string   `json:"git_branch"`
	SourceDir        string   `json:"source_dir"`
	PostUpdateParams []string `json:"post_update_params"`
}
