package main

import (
	"context"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func main() {
	// setup logging using klog
	log.SetLogger(klog.NewKlogr())

	// create the webhook server
	hookServer := &webhook.Server{
		Port:    8443,
		CertDir: "./certs",
	}

	// register one or more webhooks
	hookServer.Register("/annotate", asHook(annotate, "annotate"))

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

// convert a handler to a webhook adding some basic logging
func asHook(handler func(ctx context.Context, req webhook.AdmissionRequest) webhook.AdmissionResponse, name string) *admission.Webhook {
	l := log.Log.WithName("webhooks/" + name)
	wh := &admission.Webhook{
		Handler:      admission.HandlerFunc(loggingMiddleware(l)(handler)),
		RecoverPanic: false,
	}
	wh.InjectLogger(l)
	return wh
}

// types to create middleware
type handleFunc func(ctx context.Context, r admission.Request) admission.Response
type middleware func(handleFunc) handleFunc

// logging middleware that logs information about request and response,
// after the request was handled
func loggingMiddleware(logger logr.Logger) middleware {
	return func(handler handleFunc) handleFunc {
		return func(ctx context.Context, r admission.Request) (res admission.Response) {
			defer func(ts time.Time) {
				logger.Info("request_handled",
					"uid", r.UID,
					"allowed", res.Allowed,
					"operation", r.Operation,
					"group", r.Kind.Group,
					"version", r.Kind.Version,
					"kind", r.Kind.Kind,
					"name", r.Name,
					"namespace", r.Namespace,
					"elapsed", time.Since(ts),
					"dryrun", *r.DryRun,
				)
			}(time.Now())
			res = handler(ctx, r)
			return res
		}
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
