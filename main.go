package main

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// based on https://github.com/kubernetes-sigs/controller-runtime/blob/v0.13.0/pkg/webhook/example_test.go

func main() {
	// create the webhook server
	hookServer := &webhook.Server{
		Port:    8443,
		CertDir: "./certs",
	}

	// register one or more webhooks
	hookServer.Register("/annotate", asHook(annotate))

	// optionally start the metrics server
	go metricsServer(":8080")

	// start the webhooks server
	err := hookServer.StartStandalone(signals.SetupSignalHandler(), scheme.Scheme)
	if err != nil {
		panic(err)
	}
}

// example mutating webhook
func annotate(ctx context.Context, req webhook.AdmissionRequest) webhook.AdmissionResponse {
	// create a basic patch response
	res := webhook.Patched("annotating object",
		webhook.JSONPatchOp{Operation: "add", Path: "/metadata/annotations/access", Value: "granted"},
		webhook.JSONPatchOp{Operation: "add", Path: "/metadata/annotations/reason", Value: "not so secret"},
	)

	// do some more with it. i.e.  add warnings
	res.Warnings = append(res.Warnings, "be careful, now!")

	// then return the final product
	return res
}

// helper function to register a new webhook. This is only to reduce verbosity
func asHook(handler func(ctx context.Context, req webhook.AdmissionRequest) webhook.AdmissionResponse) *admission.Webhook {
	return &admission.Webhook{
		Handler: admission.HandlerFunc(handler),
	}
}

// initialize and start the metrics server.
// panics if the listener cannot be created or
// if the server could not start
func metricsServer(addr string) {
	metricsListener, err := metrics.NewListener(addr)
	if err != nil {
		panic(err)
	}
	s := http.Server{
		Handler: promhttp.InstrumentMetricHandler(
			metrics.Registry,
			promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}),
		),
	}
	if err := s.Serve(metricsListener); err != nil {
		panic(err)
	}
}
