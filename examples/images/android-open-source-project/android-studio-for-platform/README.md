# Android Studio for Platform

This example installs the requisite tooling to use Android Studio for Platform to develop against the [Android Open Source Project](https://source.android.com/).

> Note: In order to interact with the Cuttlefish emulator, you must enable nested virtualization in the [machine settings](https://cloud.google.com/workstations/docs/create-configuration#define_machine_settings) of your Cloud Workstations configuration.

This example can be built with the included cloudbuild.yaml by specifying substitutions for the image name and Android Studio for Platform release channel:

```
# Stable build
gcloud builds submit --substitutions=_IMAGE_NAME=us-central1-docker.pkg.dev/your-project-id/your-repository/android-studio-for-platform-stable,_ASFP_VERSION=stable

# EAP build
gcloud builds submit --substitutions=_IMAGE_NAME=us-central1-docker.pkg.dev/your-project-id/your-repository/android-studio-for-platform-canary,_ASFP_VERSION=canary
```

Or can be built locally using:

```
docker build -t android-studio-for-platform .
```

To run / test the container locally, use the following command:

```
docker run --rm -it --privileged -p 8080:80 android-studio-for-platform
```
> Note: we recommend running the command with the `--privileged` which more accurately reflects the running conditions of the container while running in Cloud Workstations. This is not strictly required if you don't need to exercise all functionality provided by Cloud Workstations. For example, omitting the flag will disable docker-in-docker functionality.

Then navigate to localhost:8080 on your local machine to access Code OSS.
