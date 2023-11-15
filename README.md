# Gardinar

![Gardinar](./gardinar.jpg)

Gardinar is a lightweight tool designed to automate the update process on a server. It is especially useful in Continuous Integration (CI) environments or when you cannot use or want to avoid SSH for any reason. This tool uses a webhook to call from services like GitHub Actions and triggers the update process on the server without the need for SSH access.

## How Gardinar Works

Gardinar is built with simplicity and ease of use in mind. It follows a webhook-based approach to automate the update process seamlessly. Here's how it works:

1. Once set up, Gardinar creates a webhook URL that can be used as the target endpoint in services like GitHub Actions.
2. When a webhook is triggered, Gardinar receives the payload with update instructions from the CI service.
3. Gardinar then processes the payload and performs the necessary update actions on the server.
4. Once completed, Gardinar can provide feedback about the update status to the CI service for further processing or notification purposes.

With Gardinar, you can streamline the deployment process in your CI workflow effortlessly and avoid the need for SSH access to your server.

## Features

Gardinar offers several features designed to make the update process efficient and hassle-free:

- **No SSH Required**: Gardinar eliminates the need for SSH access to your server, making it ideal for situations where SSH is restricted, deemed insecure, or simply not preferred.
- **Webhook Integration**: Gardinar seamlessly integrates with services that support webhooks such as GitHub Actions, allowing you to automate the update process with ease.
- **Flexible Configuration**: Gardinar provides a configuration file that allows you to define the update steps and customize the behavior according to your requirements.
- **Update Logging**: Gardinar logs all activities and provides detailed information about the update process, making it easier to debug and troubleshoot any issues.
- **Error Handling**: Gardinar includes error handling mechanisms to ensure that the update process is reliable and provides feedback to the CI service if any errors occur.

## Getting Started

To get started with Gardinar, follow these simple steps:

1. Download the binary release that corresponds to your operating system and architecture from the [Releases](https://github.com/anvie/gardinar/releases) page. 
2. Extract the downloaded binary to a desired location on your server.
3. Create a `.env` file in the same directory as the binary. You can use the `example.env` file as a template. Modify the necessary variables according to your setup.
4. Optionally, configure the update steps and behavior in the `config.yaml` file located in the same directory.
5. Run Gardinar as a service, for example, using `systemd` or `supervisord`.

Here is an example `systemd` service unit file (`gardinar.service`):

```
[Unit]
Description=Gardinar

[Service]
ExecStart=/path/to/gardinar
WorkingDirectory=/path/to/extracted/binary
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

Gardinar will now be running as a service, ready to receive webhooks and perform updates on your server without the need for SSH access.

## Testing

Once running, you can test using curl:

```bash
curl -vvv -H "X-SECRET-KEY: MySecretKey" -H "Content-type: application/json" -X POST http://localhost:8080/webhook -d '
{
  "version": "1.0",
  "commit_hash": "abc856def",
  "source_dir": "/Users/example/source",
  "git_branch": "main",
  "post_update_params": [
    "restart"
  ]
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
