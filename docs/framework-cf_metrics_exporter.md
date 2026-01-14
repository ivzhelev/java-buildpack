# cf-metrics-exporter (Agent Mode)

This framework integrates the [cf-metrics-exporter](https://github.com/rabobank/cf-metrics-exporter) as a Java agent in the Java buildpack.

## Enabling the Exporter

Set the following environment variable to enable the agent:

```
CF_METRICS_EXPORTER_ENABLED=true
```

## Configuration

- **CF_METRICS_EXPORTER_ENABLED**: Set to `true` to enable the agent (default: disabled).
- **CF_METRICS_EXPORTER_PROPS**: (Optional) Properties string to pass to the agent, e.g. `port=9090,foo=bar`.

## How it Works

- The agent JAR is downloaded during the buildpack supply phase.
- The agent is injected into the JVM at runtime using the `-javaagent` option.
- If `CF_METRICS_EXPORTER_PROPS` is set, its value is appended to the `-javaagent` option.

## Example

```
CF_METRICS_EXPORTER_ENABLED=true
CF_METRICS_EXPORTER_PROPS="port=9090,foo=bar"
```

## Version

- Default version: 0.7.1
- Default download URI: https://github.com/rabobank/cf-metrics-exporter/releases/download/0.7.1/cf-metrics-exporter-0.7.1.jar

## Notes

- The agent is injected with priority 43 in JAVA_OPTS (after other APM agents).
- The agent JAR is placed in `.java-buildpack/cf_metrics_exporter/` within the dependency directory.

