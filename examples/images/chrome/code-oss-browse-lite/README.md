# Code OSS with Browse Lite preinstalled

This example extends the [Cloud Workstations base editor image](https://cloud.google.com/workstations/docs/preconfigured-base-images) by installing [Google Chrome](https://www.google.com/chrome/) and adding the [Browse Lite extension](https://open-vsx.org/extension/antfu/browse-lite) to Code OSS as a `@builtin` extension. Users are able to interact with a rendered view of chrome (running on the remote workstation) within the Code OSS UI.

To use, start a Cloud Workstation and launch Code OSS, then open the command palette (Ctrl + Shift + P) and enter / select "Browse Lite: Open...".

This example can be be built with the included cloudbuild.yaml by specifying an image name:

```
gcloud builds submit --substitutions _IMAGE_NAME=us-central1-docker.pkg.dev/your-project-id/your-repository/chrome-code-oss-browse-lite
```

Or can be built locally using:

```
docker build -t chrome-code-oss-browse-lite .
```

To run / test the container locally, use the following command:

```
docker run --rm -it --privileged -p 8080:80 chrome-code-oss-browse-lite
```

> Note: we recommend running the command with the `--privileged` which more accurately reflects the running conditions of the container while running in Cloud Workstations. This is not strictly required if you don't need to exercise all functionality provided by Cloud Workstations. For example, omitting the flag will disable docker-in-docker functionality.

Then navigate to localhost:8080 on your local machine to access Code OSS.
