
# k8s-oom-monitor

Monitor that alerts to Slack when a Kubernetes pod is killed with OutOfMemory (OOM).

This Go application has less than 100 lines of code but provides a critical monitoring feature for our infrastructure.


### Install

1. Go to [Your Slack Apps](https://api.slack.com/apps) and create a new one. You can use any name you want and select the workspace you will use to receive the messages.
2. In the sidebar visit the _Incoming Webhooks_ section and enable it with the switch On/Off in the top-right of the screen.
3. In the same sidebar move to _Install App_ and install the application in your workspace. It will ask for permissions and you can select the destination channel.
4. Copy the _Webhook URL for Your Workspace_ when you return to the install page after succesfully authorizing the new application.
5. Save the example deployment configuration to `k8s-oom-monitor.yaml` replacing the URL with the webhook URL you copied in step 4.

```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: k8s-oom-monitor
spec:
  replicas: 1
  revisionHistoryLimit: 10
  strategy:
    rollingUpdate:
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: k8s-oom-monitor
    spec:
      containers:
      - name: k8s-oom-monitor
        image: altipla/k8s-oom-monitor:v1.0.0
        env:
        - name: SLACK_WEBHOOK
          value: https://REPLACE_URL/WITH_THE_REAL_ONE
        resources:
          requests:
            cpu: 10m
            memory: 50Mi
          limits:
            memory: 50Mi
```

6. Deploy the file to the Kubernetes cluster you want to monitorize: `kubectl apply -f k8s-oom-monitor.yaml`


### Contributing

You can make pull requests or create issues in GitHub. Any code you send should be formatted using `gofmt`


### License

[MIT License](LICENSE)
