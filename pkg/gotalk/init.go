package gotalk

import (
	"github.com/libp2p/go-libp2p-core/metrics"
	dhtMetric "github.com/libp2p/go-libp2p-kad-dht/metrics"
	quicMetrics "github.com/lucas-clemente/quic-go/metrics"
	"go.opencensus.io/stats/view"
)

func init() {
	metrics.NewBandwidthCounter()
	er := view.Register(quicMetrics.DefaultViews...)
	if er != nil {
		logger.Error("Error registering views ", er)
	}
	er = view.Register(dhtMetric.DefaultViews...)
	if er != nil {
		logger.Error("Error registering views ", er)
	}
}
