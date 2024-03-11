# domain_exporter

The `domain_exporter` is a lightweight service designed to perform WHOIS lookups for a list of domains specified in the configuration file and expose the results on a "/metrics" endpoint for consumption via Prometheus.

This project was heavily inspired by:
- [shift/domain_exporter](https://github.com/shift/domain_exporter)
- [caarlos0/domain_exporter](https://github.com/caarlos0/domain_exporter)


## Configuration

The service is configured via a YAML file (domains.yml by default) which specifies the domains to be monitored. An example configuration looks like this:

```yaml
domains:
  - google.com
  - google.co.uk
```

## Flags:

The following flags can be used to configure the behavior of the domain_exporter:

```bash
usage: domain_exporter [<flags>]

Flags:
  -bind string
    	The address to listen on for HTTP requests. (default ":9203")
  -config string
    	domain_exporter configuration file. (default "domains.yml")
  -debug
    	sets log level to debu
  -version
    	prints the version
```

## Example Output
```text
# HELP domain_expiration_seconds UNIX timestamp when the WHOIS record states this domain will expire
# TYPE domain_expiration_seconds gauge
domain_expiration_seconds{domain="google.co.uk"} 1.7394912e+09
domain_expiration_seconds{domain="google.com"} 1.8525168e+09
# HELP domain_last_change_seconds UNIX timestamp when the WHOIS record states this domain will expire
# TYPE domain_last_change_seconds gauge
domain_last_change_seconds{domain="google.co.uk"} 1.6781472e+09
domain_last_change_seconds{domain="google.com"} 1.7019072e+09
# HELP domain_parsed That the domain was parsed
# TYPE domain_parsed gauge
domain_parsed{domain="google.co.uk"} 1
domain_parsed{domain="google.com"} 1
```

## Docker image
We provide a Docker image for easy deployment, available on the GitHub Container Registry. You can pull the image using:

```bash
docker pull ghcr.io/numero33/domain_exporter/domain_exporter:main
```

## Example Prometheus Alert

Below is an example of Prometheus alert rules that can be configured to monitor domain expiration and availability:

```yaml
groups:
  - name: ./domain.rules
    rules:
      - alert: DomainExpiring
        expr: ((domain_expiration_seconds-time())/86400) < 30
        for: 24h
        labels:
          severity: warning
        annotations:
          description: '{{ $labels.domain }} expires in {{ $value }} days'
      - alert: DomainUnfindable
        expr: domain_expiration_parsed == 1
        for: 24h
        labels:
          severity: critical
        annotations:
          description: 'Unable to find or parse expiry for {{ $labels.domain }}'
      - alert: DomainMetricsAbsent
        expr: absent(domain_expiration_seconds) > 0
        for: 1h
        labels:
          severity: warning
        annotations:
          description: Metrics for domain-exporter are absent.

```
These rules will trigger alerts based on domain expiration dates and the availability of WHOIS records. Adjust the thresholds and severity levels as needed for your monitoring setup.

## Contribute and Report Issues

If you encounter any bugs or issues while using `domain_exporter`, please don't hesitate to create an issue on GitHub. Additionally, we welcome contributions and improvements from the community. If you have a feature request or would like to contribute code, feel free to open a pull request. Your feedback and contributions are invaluable in helping us maintain and improve this tool for everyone's benefit 🎉