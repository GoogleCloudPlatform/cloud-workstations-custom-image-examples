# Chrome - Code OSS Xpra

This example extends the [Cloud Workstations base editor image](https://cloud.google.com/workstations/docs/preconfigured-base-images) by installing [Google Chrome](https://www.google.com/chrome/) and [Xpra](https://xpra.org/index.html). Users are able to interact with google-chrome (or **any arbitrary application**) running in an xpra session that can be accessed via the Xpra html client.

To use, start a Cloud Workstation and launch Code OSS, then open a terminal in Code OSS and execute `xpra-chrome`. This will execute a [helper script](assets/usr/bin/xpra-chrome) to start google chrome under an xpra session and serve the xpra client on port 8080 of the workstation. Access port 8080 either by using the `Ports View` in Code OSS, or browse to the port at https://8080-${WORKSTATION_HOST} to interact with google-chrome.

This example can be be built with the included cloudbuild.yaml by specifying an image name:

```
gcloud builds submit --substitutions _IMAGE_NAME=us-central1-docker.pkg.dev/your-project-id/your-repository/chrome-code-oss-browse-xpra
```

Or can be built locally using:

```
docker build -t chrome-code-oss-xpra .
```

To run / test the container locally, use the following command:

```
docker run --rm -it --privileged -p 8080:80 -p 8081:8080 chrome-code-oss-xpra
```

> Note: this conainer must be started using the `--privileged` switch.

Then navigate to localhost:8080 on your local machine to access Code OSS. After connecting to Code OSS, open a terminal and execute `xpra-chrome` to start chrome on the workstation, then navigate to localhost:8081 in a new browser tab.
