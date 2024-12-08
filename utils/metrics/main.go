package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/sjhitchner/toolbox/pkg/metrics"
	"github.com/sjhitchner/toolbox/pkg/metrics/datadog"
)

func main() {
	done := make(chan struct{})

	backend, err := datadog.New("10.0.0.53:8125", "metricTest")
	if err != nil {
		log.Fatal(err)
	}
	metrics.Initialize(done, backend)

	for {
		testGauge()
		testTimer()
		<-time.After(time.Second)
	}
}

func testGauge() {
	m := metrics.GaugeAt("gauge", rand.Float64())
	defer m.Emit()

	fmt.Println("gauge")
}

func testTimer() {
	m := metrics.Timer("timer")
	defer m.Emit()

	fmt.Println("timer")
	delay := time.Duration(rand.Intn(10) + 1)
	<-time.After(delay * time.Millisecond)
}
