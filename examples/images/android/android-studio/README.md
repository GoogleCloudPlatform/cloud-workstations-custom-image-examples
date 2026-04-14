<!--
Copyright 2024-2026 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

# Android Studio

This example installs the requisite tooling to develop applications using Android Studio in a Cloud Workstations environment.

This example can be built with the included cloudbuild.yaml by specifying substitutions for the image name and optionally the version of Android Studio you would like to install:

```
gcloud builds submit --substitutions=_IMAGE_NAME=us-central1-docker.pkg.dev/your-project-id/your-repository/android-studio,_ANDROID_STUDIO_VERSION=2024.2.2.13
```

Or can be built locally using:

```
docker build -t android-studio .
```

To run / test the container locally, use the following command:

```
docker run --rm -it --privileged -p 8080:80 android-studio
```

> Note: we recommend running the command with the `--privileged` which more accurately reflects the running conditions of the container while running in Cloud Workstations. This is not strictly required if you don't need to exercise all functionality provided by Cloud Workstations. For example, omitting the flag will disable docker-in-docker functionality.

Then navigate to localhost:8080 on your local machine to interact with the container using a web browser.
