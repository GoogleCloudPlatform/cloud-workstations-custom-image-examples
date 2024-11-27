# Code OSS for Android Open Source Development

Example installs the requisite tooling to use Code OSS to develop against the [Android Open Source Project](https://source.android.com/).

> Note: In order to interact with the Cuttlefish emulator, you must enable nested virtualization in the [machine settings](https://cloud.google.com/workstations/docs/create-configuration#define_machine_settings) of your Cloud Workstations configuration.

This example can be built with the included cloudbuild.yaml by specifying substitutions for the image name:

```
gcloud builds submit --substitutions _IMAGE_NAME=us-central1-docker.pkg.dev/your-project-id/your-repository/code-oss-aosp
```

Or can be built locally using:

```
docker build -t code-oss-aosp .
```

To run / test the container locally, use the following command:

```
docker run --rm -it --privileged -p 8080:80 code-oss-aosp
```
> Note: we recommend running the command with the `--privileged` which more accurately reflects the running conditions of the container while running in Cloud Workstations. This is not strictly required if you don't need to exercise all functionality provided by Cloud Workstations. For example, omitting the flag will disable docker-in-docker functionality.

Then navigate to localhost:8080 on your local machine to access Code OSS.
