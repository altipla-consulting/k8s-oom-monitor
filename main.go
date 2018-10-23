package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	log.SetFormatter(new(log.JSONFormatter))
	k8s.Register("", "v1", "events", true, new(corev1.Event))

	if os.Getenv("SLACK_WEBHOOK") == "" {
		return fmt.Errorf("Define SLACK_WEBHOOK env variable with a webhook URL")
	}

	client, err := k8s.NewInClusterClient()
	if err != nil {
		return err
	}

	var event corev1.Event
	watcher, err := client.Watch(ctx, "default", &event)
	if err != nil {
		return err
	}
	defer watcher.Close()

	log.WithField("webhook", os.Getenv("SLACK_WEBHOOK")).Println("Watching events to detect OOM pods")
	for {
		if _, err := watcher.Next(&event); err != nil {
			return err
		}

		if event.GetReason() != "OOMKilling" {
			continue
		}

		log.WithFields(log.Fields{
			"node":    event.InvolvedObject.GetName(),
			"message": event.GetMessage(),
		}).Info("OOM detected")

		msg := map[string]interface{}{
			"text": "OutOfMemory (OOM) pod killed in the cluster",
			"attachments": []map[string]interface{}{
				{
					"text":      "```" + event.GetMessage() + "```",
					"footer":    "node " + event.InvolvedObject.GetName(),
					"mrkdwn_in": []string{"text"},
				},
			},
		}
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(msg); err != nil {
			return err
		}

		resp, err := http.Post(os.Getenv("SLACK_WEBHOOK"), "application/json", &buf)
		if err != nil {
			log.WithField("error", err.Error()).Error("Cannot notify Slack right now")
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.WithField("status", resp.Status).Error("Cannot notify Slack right now")
			continue
		}
	}

	return nil
}
