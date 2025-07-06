package internal

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	"github.com/VictoriaMetrics/metrics"
	saiav1 "github.com/clementnuss/saia-grpc-service/gen/go/saia/v1"
	"github.com/clementnuss/saia-grpc-service/gen/go/saia/v1/saiav1connect"
)

type BiogasExporter struct {
	Metrics      []Metric
	client       saiav1connect.SaiaPcdServiceClient
	updateTicker *time.Ticker
}

type RegisterType int

const (
	Register RegisterType = iota
	RegisterFloat
	Flag
	Input
	Output
	Counter
	Timer
)

func (rt RegisterType) String() string {
	switch rt {
	case Register:
		return "R"
	case RegisterFloat:
		return "R Float"
	case Flag:
		return "Flag"
	case Input:
		return "Input"
	case Output:
		return "Output"
	case Counter:
		return "Counter"
	case Timer:
		return "Timer"
	default:
		return "unknown type"
	}
}

type Metric struct {
	Name         string       `csv:"Nom Métrique Prometheus"`
	RegisterType RegisterType `csv:"Type Registre"`
	Address      uint32       `csv:"Adresse"`
	Description  string       `csv:"Description"`
}

func NewBiogasExporter(csvPath string, grpcClient saiav1connect.SaiaPcdServiceClient) (*BiogasExporter, error) {
	metrics, err := ParseCSV(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load metrics from CSV: %w", err)
	}

	return &BiogasExporter{
		Metrics: metrics,
		client:  grpcClient,
	}, nil
}

// Start begins the metric collection process
func (e *BiogasExporter) Start(updateInterval time.Duration) {
	// Initial update
	e.updateAllMetrics()

	// Start periodic updates
	e.updateTicker = time.NewTicker(updateInterval)
	go func() {
		for range e.updateTicker.C {
			e.updateAllMetrics()
		}
	}()
}

// Stop stops the metric collection
func (e *BiogasExporter) Stop() {
	if e.updateTicker != nil {
		e.updateTicker.Stop()
	}
}

// updateAllMetrics updates all configured metrics
func (e *BiogasExporter) updateAllMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s := time.Now()

	for _, metric := range e.Metrics {
		value, err := e.readMetricValue(ctx, metric)
		if err != nil {
			slog.Info("Failed to read metric", "error", err, "metric", metric)
			continue
		}

		metrics.GetOrCreateGauge(metric.Name, nil).Set(value)
	}

	metrics.GetOrCreateHistogram("biogaz_exporter_parse_duration_seconds").UpdateDuration(s)
	slog.Info("parse duration", "duration", time.Since(s))
}

// readMetricValue reads a single metric value via gRPC
func (e *BiogasExporter) readMetricValue(ctx context.Context, metric Metric) (float64, error) {
	switch metric.RegisterType {
	case Register:
		resp, err := e.client.ReadRegister(
			ctx, connect.NewRequest(&saiav1.ReadRegisterRequest{
				Address:  metric.Address,
				DataType: &saiav1.ReadRegisterRequest_AsInt{},
			}))
		if err != nil {
			return 0, err
		}

		return float64(resp.Msg.GetIntValue()), nil
	case RegisterFloat:
		resp, err := e.client.ReadRegister(
			ctx, connect.NewRequest(&saiav1.ReadRegisterRequest{
				Address:  metric.Address,
				DataType: &saiav1.ReadRegisterRequest_AsFloat{},
			}))
		if err != nil {
			return 0, err
		}

		return float64(resp.Msg.GetFloatValue()), nil
	case Flag:
		resp, err := e.client.ReadFlag(
			ctx, connect.NewRequest(&saiav1.ReadFlagRequest{Address: metric.Address}))
		if err != nil {
			return 0, err
		}

		return boolToFloat64(resp.Msg.GetValue()), nil
	case Input:
		resp, err := e.client.ReadInput(
			ctx, connect.NewRequest(&saiav1.ReadInputRequest{Address: metric.Address}))
		if err != nil {
			return 0, err
		}

		return boolToFloat64(resp.Msg.GetValue()), nil
	case Output:
		resp, err := e.client.ReadOutput(
			ctx, connect.NewRequest(&saiav1.ReadOutputRequest{Address: metric.Address}))
		if err != nil {
			return 0, err
		}

		return boolToFloat64(resp.Msg.GetValue()), nil
	default:
		return 0, fmt.Errorf("unknown register type: %v", metric.RegisterType)
	}
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1.0
	} else {
		return 0.0
	}
}
