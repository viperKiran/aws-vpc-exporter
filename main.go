package main

import (
	"context"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"

	"github.com/aws/smithy-go"
	"github.com/pkg/errors"

	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	namespace string = "aws_vpc"
)

// AvailableIPAddressCount tracks available ip address in the subnet.
var AvailableIPAddressCount = prometheus.NewDesc(
	"aws_vpc_subnet_available_ip_addresses",
	"The number of unused private IPv4 addresses in the subnet.",
	[]string{"subnet_id"}, nil,
)

type Exporter struct {
	client                     *ec2.Client
	vpcID                      string
	ctx                        context.Context
	logger                     log.Logger
	lastScrapeError            prometheus.Gauge
	totalScrapes, scrapeErrors prometheus.Counter
}

func (e *Exporter) ec2Client() (*ec2.Client, error) {
	if e.client != nil {
		return e.client, nil
	}
	// Create AWS session
	s, err := config.LoadDefaultConfig(e.ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to load aws default config")
	}

	return ec2.NewFromConfig(s), nil
}

func NewExporter(ctx context.Context, vpcID string, logger log.Logger) (*Exporter, error) {
	return &Exporter{
		client: nil,
		vpcID:  vpcID,
		logger: logger,
		ctx:    ctx,
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrape_total",
			Help:      "Current total VPC subnet scrapes.",
		}),
		lastScrapeError: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "exporter_last_scrape_error",
			Help:      "Whether the last scrape of metrics from VPC subnet resulted in an error (1 for error, 0 for success).",
		}),
		scrapeErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrape_errors_total",
			Help:      "Total number of times an error occurred scraping a VPC subnet.",
		}),
	}, nil
}

func (e *Exporter) getSubnetsAvailableIPs() (map[string]int32, error) {
	ec2Client, err := e.ec2Client()
	if err != nil {
		level.Error(e.logger).Log("msg", "Authenticate error", "err", err)
		e.scrapeErrors.Inc()
		e.lastScrapeError.Set(1)
		return nil, err
	}
	input := &ec2.DescribeSubnetsInput{
		Filters: []types.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []string{
					e.vpcID,
				},
			},
		},
	}
	p := ec2.NewDescribeSubnetsPaginator(ec2Client, input)
	subnets := []types.Subnet{}
	for {
		o, err := p.NextPage(e.ctx)
		if err != nil {
			var apiErr smithy.APIError
			if ok := errors.As(err, &apiErr); ok && (apiErr.ErrorCode() == "AuthFailure" || apiErr.ErrorCode() == "UnauthorizedOperation") {
				level.Error(e.logger).Log("msg", "Authentication error", "err", err)
				e.client = nil
			}
			e.scrapeErrors.Inc()
			e.lastScrapeError.Set(1)
			return nil, errors.Wrap(err, "could not describe subnets")
		}
		level.Debug(e.logger).Log("msg", "Describe subnets result in paginator", "result", o)
		subnets = append(subnets, o.Subnets...)
		if !p.HasMorePages() {
			break
		}
	}

	level.Debug(e.logger).Log("msg", "Described subnets", "output", subnets)

	subnetInfo := map[string]int32{}
	for _, subnet := range subnets {
		subnetInfo[aws.ToString(subnet.SubnetId)] = *subnet.AvailableIpAddressCount
	}
	return subnetInfo, nil
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- AvailableIPAddressCount

	ch <- e.totalScrapes.Desc()
	ch <- e.lastScrapeError.Desc()
	ch <- e.scrapeErrors.Desc()
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.totalScrapes.Inc()
	subnetInfo, _ := e.getSubnetsAvailableIPs()
	level.Debug(e.logger).Log("msg", "Debug: subnet info", "subnetInfo", subnetInfo)
	for subnetName, availableIP := range subnetInfo {
		ch <- prometheus.MustNewConstMetric(AvailableIPAddressCount, prometheus.GaugeValue, float64(availableIP), subnetName)
	}
	ch <- e.totalScrapes
	ch <- e.lastScrapeError
	ch <- e.scrapeErrors
}

func main() {
	var (
		webConfig  = webflag.AddFlags(kingpin.CommandLine)
		listenAddr = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9223").String()
		vpcID      = kingpin.Flag("aws.vpc-id", "AWS VPC ID").Required().String()
	)
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)

	level.Info(logger).Log("msg", "Starting aws-vpc-exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "context", version.BuildContext())

	exporter, err := NewExporter(context.Background(), *vpcID, logger)
	if err != nil {
		level.Error(logger).Log("msg", "Error initializing exporter", "err", err)
		os.Exit(1)
	}

	level.Debug(logger).Log("msg", "Registered exporter")
	prometheus.MustRegister(exporter)
	prometheus.MustRegister(version.NewCollector("aws_vpc_exporter"))

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>AWS VPC Exporter</title></head>
             <body>
             <h1>AWS VPC Exporter</h1>
             <p><a href='/metrics'>Metrics</a></p>
             </body>
             </html>`))
	})
	level.Info(logger).Log("msg", "Listening on address", "address", *listenAddr)

	srv := &http.Server{Addr: *listenAddr}
	if err := web.ListenAndServe(srv, *webConfig, logger); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
