# Gardinar

![Gardinar](./gardinar.jpg)

Gardinar is a lightweight tool designed to automate the update process on a server. It is especially useful in Continuous Integration (CI) environments or when you cannot use or want to avoid SSH for any reason. This tool uses a webhook to call from services like GitHub Actions and triggers the update process on the server without the need for SSH access.

## How Gardinar Works

Gardinar is built with simplicity and ease of use in mind. It follows a webhook-based approach to automate tasks seamlessly. Here's how it works:

1.  Once set up, Gardinar creates a webhook URL that can be used as the target endpoint in services like GitHub Actions.
2.  When a webhook is triggered, Gardinar receives a JSON payload specifying a task to run and any parameters.
3.  Gardinar looks up the task in its configuration, finds the corresponding script or command, and executes it.
4.  Once completed, Gardinar can provide feedback about the task status to the CI service for further processing or notification purposes.

With Gardinar, you can streamline your deployment and operational tasks in your CI workflow effortlessly and avoid the need for SSH access to your server.

## Features

Gardinar offers several features designed to make automation efficient and hassle-free:

-   **No SSH Required**: Gardinar eliminates the need for SSH access to your server, making it ideal for situations where SSH is restricted, deemed insecure, or simply not preferred.
-   **Webhook Integration**: Gardinar seamlessly integrates with services that support webhooks such as GitHub Actions, allowing you to automate any scripted task with ease.
-   **Flexible Configuration**: Gardinar provides a configuration file that allows you to define a set of named tasks and the scripts or commands they execute.
-   **Task Logging**: Gardinar logs all activities and provides detailed information about the task execution, making it easier to debug and troubleshoot any issues.
-   **Error Handling**: Gardinar includes error handling mechanisms to ensure that the task execution is reliable and provides feedback to the CI service if any errors occur.

## Getting Started

To get started with Gardinar, follow these simple steps:

1.  Download the binary release that corresponds to your operating system and architecture from the [Releases](https://github.com/anvie/gardinar/releases) page.
2.  Extract the downloaded binary to a desired location on your server.
3.  Create a `config.yaml` file in the same directory as the binary. You can use the `example.yaml` file as a template.
4.  Run Gardinar as a service, for example, using `systemd` or `supervisord`.

Here is an example `config.yaml`:

```yaml
listen_port: "8800"
secret_key: "YourSecretKey"
tasks:
  git-update: "/path/to/your/git-update-script.sh"
  deploy-frontend: "/path/to/your/deploy-script.sh"
  restart-service: "systemctl restart my-app"
```

Here is an example `systemd` service unit file (`gardinar.service`):

```
[Unit]
Description=Gardinar

[Service]
ExecStart=/path/to/gardinar --config /path/to/your/config.yaml
WorkingDirectory=/path/to/your/gardinar/directory
User=yourusername
Group=yourgroup

[Install]
WantedBy=multi-user.target
```

Make sure to update the `ExecStart` and `WorkingDirectory` paths to match your setup.

After creating the service unit file, you can enable and start the service using the following commands:

```
sudo systemctl enable gardinar.service
sudo systemctl start gardinar.service
```

Gardinar will now be running as a service, ready to receive webhooks and execute your defined tasks.

## Testing

Once running, you can test a task using curl:

```bash
curl -vvv -H "X-SECRET-KEY: YourSecretKey" -H "Content-type: application/json" -X POST http://localhost:8800/webhook -d '
{
  "task": "git-update",
  "source_dir": "/Users/example/source",
  "params": [
    "main",
    "restart"
  ]
}
'
```

## Github Action Example

```yaml
name: deploy
on:
  push:
    branches:
      - main
jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Notify Gardinar
        run: |
          curl -vvv -H "X-SECRET-KEY: ${{ secrets.GARDINAR_SECRET }}" -H "Content-type: application/json" -X POST https://gardinar-endpoint.example.com/webhook -d '
          {
            "task": "deploy-frontend",
            "source_dir": "/frontend",
            "params": ["v1.2.3"]
          }
          '
```

## Contributing

Contributions to Gardinar are always welcome! Whether it's bug fixes, new features, or documentation improvements, feel free to submit a pull request. Please make sure to follow the project's code style guidelines and include tests for any new functionality.

## License

Gardinar is released under the [MIT License](https://github.com/anvie/gardinar/LICENSE). Feel free to use, modify, and distribute this tool as per the terms of the license.

## Support

If you encounter any issues, have questions, or need assistance with Gardinar, please open an issue on the GitHub repository. Our team will be more than happy to help.

[] Robin Syihab
