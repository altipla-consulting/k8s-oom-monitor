package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

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

	notified := map[string]bool{}
	for {
		if err := watch(ctx, client, notified); err != nil {
			log.WithField("error", err.Error()).Error("Listening failed, retrying in 5 seconds")
		}
		time.Sleep(5 * time.Second)
	}

	return nil
}

func watch(ctx context.Context, client *k8s.Client, notified map[string]bool) error {
	var event corev1.Event
	watcher, err := client.Watch(ctx, "default", &event)
	if err != nil {
		return fmt.Errorf("cannot open the watcher: %v", err)
	}
	defer watcher.Close()

	log.WithField("webhook", os.Getenv("SLACK_WEBHOOK")).Println("Watching events to detect OOM pods")
	for {
		if _, err := watcher.Next(&event); err != nil {
			if err == io.EOF {
				return nil
			}

			return fmt.Errorf("error while watching the next event: %v", err)
		}

		if event.GetReason() != "OOMKilling" {
			continue
		}

		name := event.Metadata.GetName()
		if notified[name] {
			continue
		}
		notified[name] = true

		log.WithFields(log.Fields{
			"name":    name,
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
			return fmt.Errorf("cannot serialize slack request: %v", err)
		}

		resp, err := http.Post(os.Getenv("SLACK_WEBHOOK"), "application/json", &buf)
		if err != nil {
			return fmt.Errorf("cannot notify slack: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("bad slack status: %v", resp.Status)
		}
	}

	return nil
}
