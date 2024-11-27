# Gnome Example

This example shows how to create a Cloud Workstations image that runs a desktop manager which can be accessed by remote desktop software. This sample uses [GNOME](https://www.gnome.org/) to provide the desktop environment and preinstalls [TigerVNC](https://tigervnc.org/) along with [noVNC](https://novnc.com/info.html) to provide a one-click, browser-based, interactive session.

This example can be built with the included cloudbuild.yaml by specifying substitutions for the base image and the image name for the newly built systemd image:

```
gcloud builds submit --substitutions _IMAGE_NAME=us-central1-docker.pkg.dev/your-project-id/your-repository/your-image-name
```

Or can be built locally using:

```
docker build -t gnome-vnc .
```

To run / test the container locally, use the following command:

```
docker run --rm -it --privileged -p 8080:80 gnome-vnc
```

> Note: this conainer must be started using the `--privileged` switch.

Then navigate to localhost:8080 on your local machine to access the NoVNC html client.
