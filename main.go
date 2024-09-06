package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type GitLabWebhookPayload struct {
	ObjectKind   string       `json:"object_kind"`
	UserName     string       `json:"user_name"`
	UserAvatar   string       `json:"user_avatar"` // Added to correct the missing field
	Project      Project      `json:"project"`
	Commits      []Commit     `json:"commits,omitempty"`
	After        string       `json:"after,omitempty"`
	Ref          string       `json:"ref,omitempty"`
	Repository   Repository   `json:"repository"` // Added for push events, tag push events
	MergeRequest MergeRequest `json:"merge_request,omitempty"`
}

type Project struct {
	Name   string `json:"name"`
	WebURL string `json:"web_url"`
}

type Commit struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	Author  Author `json:"author"`
	URL     string `json:"url"` // Corrected the missing URL field
}

type Author struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type Repository struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type MergeRequest struct {
	Title  string `json:"title"`
	State  string `json:"state"`
	Author Author `json:"author"`
	URL    string `json:"url"`
}

type DiscordMessage struct {
	Content string `json:"content"`
}

var discordWebhookURL = os.Getenv("DISCORD_WEBHOOK_URL")

// Convert GitLab payload to a beautiful Discord Markdown format
func formatGitLabWebhookToDiscord(payload GitLabWebhookPayload) string {
	var message strings.Builder

	if payload.ObjectKind == "push" {
		if strings.HasPrefix(payload.Ref, "refs/tags/") {
			// Handle Tag Push Event
			tagName := strings.TrimPrefix(payload.Ref, "refs/tags/")
			message.WriteString(fmt.Sprintf("**%s** pushed a new tag **%s** in **%s**:\n\n", payload.UserName, tagName, payload.Project.Name))
			message.WriteString(fmt.Sprintf("🔗 **View tag**: <%s/tags/%s>\n", payload.Project.WebURL, tagName))
		} else {
			// Handle Regular Push Event
			message.WriteString(fmt.Sprintf("**%s** pushed to branch **%s** in **%s**:\n\n", payload.UserName, payload.Ref, payload.Project.Name))
			for _, commit := range payload.Commits {
				message.WriteString(fmt.Sprintf("• **Commit ID**: %s\n", commit.ID))
				message.WriteString(fmt.Sprintf("  _by %s_ ![avatar](%s)\n", commit.Author.Name, payload.UserAvatar))
				message.WriteString(fmt.Sprintf("  [Commit URL](%s)\n", commit.URL)) // Corrected the reference to Commit URL
				message.WriteString(fmt.Sprintf("  ```\n%s\n```", commit.Message))
			}
			message.WriteString(fmt.Sprintf("🔗 **View changes**: <%s/commits/%s>\n", payload.Project.WebURL, payload.After))
		}
	} else if payload.ObjectKind == "merge_request" {
		// Handle Merge Request Event
		mr := payload.MergeRequest
		message.WriteString(fmt.Sprintf("**Merge Request**: **%s** in **%s**\n", mr.Title, payload.Project.Name))
		message.WriteString(fmt.Sprintf("State: **%s**\n", mr.State))
		message.WriteString(fmt.Sprintf("Author: _%s_ ![avatar](%s)\n", mr.Author.Name, payload.UserAvatar))
		message.WriteString(fmt.Sprintf("🔗 **View Merge Request**: <%s>\n", mr.URL))
	} else if payload.ObjectKind == "repository_update" {
		// Handle Repository Update Event
		message.WriteString(fmt.Sprintf("**Repository** **%s** was updated:\n", payload.Project.Name))
		message.WriteString(fmt.Sprintf("Updated by _%s_ ![avatar](%s)\n", payload.UserName, payload.UserAvatar))
		// Removed non-existent Changes field
		message.WriteString(fmt.Sprintf("🔗 **View Repository**: <%s>\n", payload.Project.WebURL))
	} else {
		message.WriteString("Unhandled event type")
	}

	return message.String()
}

// Send message to Discord
func sendToDiscord(message string) error {
	discordMessage := DiscordMessage{Content: message}
	body, err := json.Marshal(discordMessage)
	if err != nil {
		return err
	}

	resp, err := http.Post(discordWebhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to send message to Discord, status code: %d", resp.StatusCode)
	}
	return nil
}

// Handle GitLab Webhook
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	var payload GitLabWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Convert the payload to Discord-friendly format
	message := formatGitLabWebhookToDiscord(payload)

	// Send to Discord
	if err := sendToDiscord(message); err != nil {
		log.Printf("Error sending to Discord: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Webhook processed successfully")
}

func main() {
	http.HandleFunc("/gitlab-webhook", webhookHandler)
	port := ":4455"
	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
