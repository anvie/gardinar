package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	gitCmd := os.Getenv("GARDINAR_GIT_CMD")
	postUpdateScript := os.Getenv("GARDINAR_POST_UPDATE_SCRIPT")
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {

		// Get X-SECRET-KEY from header
		clientSecretKey := strings.TrimSpace(r.Header.Get("X-SECRET-KEY"))

		// Get SECRET_KEY from .env
		secretKey := strings.TrimSpace(os.Getenv("GARDINAR_SECRET_KEY"))

		w.Header().Set("Content-Type", "application/json")

		// Check if X-SECRET-KEY is equal to SECRET_KEY
		if clientSecretKey != secretKey {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"result":"unauthorized"}`))
			return
		}

		var webhook Webhook
		err := json.NewDecoder(r.Body).Decode(&webhook)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fmt.Printf("%+v\n", webhook)

		if webhook.SourceDir == "" {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte(`{"error":"source_dir is required"}`)); err != nil {
				log.Println(err)
			}
			return
		}

		// update git repo
		out, err := gitUpdate(gitCmd, webhook.SourceDir)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte(`{"error":"` + out + ", " + err.Error() + `"}`)); err != nil {
				log.Println(err)
			}
			return
		}

		// run post update script
		if postUpdateScript != "" {
			out, err := runPostUpdateScript(postUpdateScript, webhook.SourceDir, webhook.PostUpdateParams)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				if _, err := w.Write([]byte(`{"error":"` + out + ", " + err.Error() + `"}`)); err != nil {
					log.Println(err)
				}
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"result":"success"}`)); err != nil {
			log.Println(err)
		}
	})

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func gitUpdate(gitCmd string, sourceDir string) (string, error) {
	fmt.Printf("%s $ pull origin main\n", gitCmd)
	cmd := exec.Command(gitCmd, "pull", "origin", "main")
	cmd.Dir = sourceDir
	out, err := cmd.Output()
	if err != nil {
		log.Println(err)
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
	fmt.Printf("%s $ %s\n", postUpdateScript, sourceDir)
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
	SourceDir        string   `json:"source_dir"`
	PostUpdateParams []string `json:"post_update_params"`
}
