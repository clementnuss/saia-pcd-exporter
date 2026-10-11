# SAIA PCD Exporter

A Prometheus exporter for SAIA PCD controllers, reading their registers, flags, inputs and outputs via gRPC and exposing them as metrics. Built to monitor a biogas plant's two-decade-old PLC, which offers no modern interface beyond the SAIA S-BUS protocol.

It requires a running [saia-grpc-service](https://github.com/clementnuss/saia-grpc-service), which exposes the PLC over gRPC.

## How it works

Metrics are defined declaratively in a CSV file, mapping Prometheus metric names to PLC register addresses:

```csv
Nom Métrique Prometheus,Type Registre,Adresse,Description
biogaz_temperature_fermenteur_celsius,R,9,Température fermenteur principal
biogaz_temperature_post_fermenteur_celsius,R,12,Température post-fermenteur/nachgärer
biogaz_etat_brasseur_fermenteur1_superieur,Output,161,État brasseur supérieur fermenteur 1
```

Supported register types:

- `R` - integer register
- `R Float` - float register
- `Flag` - flag
- `Input` - digital input
- `Output` - digital output

The exporter periodically reads all configured metrics over gRPC (default: every 30 seconds) and serves them as gauges.

## Usage

```bash
go run . --addr=http://<saia-grpc-service>:50051 --csv_file=misc/prometheus_metrics.csv
```

## Endpoints

- `http://localhost:8080/metrics` - Prometheus metrics
- `http://localhost:8080/health` - health check
- `http://localhost:8080/info` - JSON list of the configured metrics

## Configuration

- `--addr`: Address of the saia-grpc-service (default: `http://192.168.85.41:50051`)
- `--csv_file`: Path to the metric list CSV (default: `misc/prometheus_metrics.csv`)