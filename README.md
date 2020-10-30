# drone-datadog

[![Build Status](https://cloud.drone.io/api/badges/upgreydd/drone-datadog/status.svg)](https://cloud.drone.io/upgreydd/drone-datadog)
[![Docker image metadata](https://images.microbadger.com/badges/image/upgreydd/drone-datadog.svg)](https://microbadger.com/images/upgreydd/drone-datadog "Get your own image badge on microbadger.com")

This plugin lets you send events and metrics to Datadog from a drone pipeline.

## Usage

To send a metric every time a pipeline runs, add this step:

```yml
- name: count-pipeline
  image: upgreydd/drone-datadog
  settings:
    api_key:
      from_secret: datadog_api_key
    metrics:
      - type: "count"
        name: "upgreydd.pipelines.count"
        value: 1.0
        tags: ["project:${DRONE_REPO_NAME}", "branch:${DRONE_BRANCH}"]
```

Sending an event is similar, both `metrics` and `events` support the `host` field:

```yml
- name: notify-pipeline
  image: upgreydd/drone-datadog
  settings:
    api_key:
      from_secret: datadog_api_key
    events:
      - title: "Building drone-datadog success"
        text: "Version ${DRONE_TAG} is available on Docker Hub"
        alert_type: "info"
        host: ${DRONE_SYSTEM_HOSTNAME}
```

You can use events to notify something bad happened:

```yml
- name: notify-pipeline
  image: upgreydd/drone-datadog
  settings:
    api_key:
      from_secret: datadog_api_key
    events:
      - title: "Build failure"
        text: "Build ${DRONE_BUILD_NUMBER} has failed"
        alert_type: "error"
  when:
    status:
    - failure
```

You can look at [this repo .drone.yml](.drone.yml) file for a real world example.
