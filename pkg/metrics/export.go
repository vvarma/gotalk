package metrics

import (
	"contrib.go.opencensus.io/exporter/ocagent"
	"github.com/ipfs/go-log/v2"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/zpages"
	"net/http"
	"time"
)

var logger = log.Logger("metrics")
var ExporterInstance *Exporter

type Exporter struct {
	oce *ocagent.Exporter
}

func NewExporter() (*Exporter, error) {
	oce, err := ocagent.NewExporter(
		ocagent.WithInsecure(),
		ocagent.WithReconnectionPeriod(5*time.Second),
		//ocagent.WithAddress("localhost:55678"), // Only included here for demo purposes.
		ocagent.WithServiceName("gotalk"))
	if err != nil {
		return nil, err
	}
	view.RegisterExporter(oce)
	return &Exporter{oce: oce}, nil

}

func (e *Exporter) Start() {

	zPagesMux := http.NewServeMux()
	zpages.Handle(zPagesMux, "/debug")
	go func() {
		if err := http.ListenAndServe(":9999", zPagesMux); err != nil {
			logger.Fatal("Error starting zPages", err)
		}
	}()

}

func RegisterViews(views []*view.View) error {
	return view.Register(views...)
}
