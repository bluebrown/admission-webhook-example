package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"

	"net/http"
	"os"

	"github.com/go-kit/log"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	logger log.Logger
)

func init() {
	w := log.NewSyncWriter(os.Stdout)
	logger = log.NewLogfmtLogger(w)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
}

func main() {
	certpath := flag.String("cert", "certs/tls.crt", "path to the certificate")
	keypath := flag.String("key", "certs/tls.key", "path to the key")
	flag.Parse()

	http.HandleFunc("/inject-sidecar", handleInjectSidecar())

	server := http.Server{Addr: ":8443"}
	logger.Log("msg", "starting server", "tag", "startup", "addr", server.Addr)

	if err := server.ListenAndServeTLS(*certpath, *keypath); err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "ListenAndServe: %s\n", err)
		os.Exit(1)
	}
}

func handleInjectSidecar() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// decode the request body into a AdmissionReview struct
		review := &admissionv1.AdmissionReview{}
		json.NewDecoder(r.Body).Decode(review)

		// log the request
		requestLogger := log.With(logger, "request.uid", review.Request.UID)
		requestLogger.Log("msg", "admission request received", "tag", "request", "operation", review.Request.Operation)

		// prepare the response
		response := &admissionv1.AdmissionResponse{}

		// get the pod object from the admission review
		pod := &corev1.Pod{}
		if err := json.NewDecoder(bytes.NewBuffer(review.Request.Object.Raw)).Decode(pod); err != nil {
			requestLogger.Log("msg", "failed to decode pod object", "tag", "decode_failure", "error", err)
			http.Error(w, "Failed to decode pod object", http.StatusBadRequest)
			return
		}

		// to ensure idempotency, we need to check if the pod already has the sidecar
		// in this example we check the name of the container
		// in a real world scenario you would probably check the image of the container
		// but this is a standard busybox image, so it doesn't make sense to check the image
		for _, container := range pod.Spec.Containers {
			if container.Name == "sidecar" {
				requestLogger.Log("msg", "skipping sidecar injection", "tag", "skipping", "reason", "sidecar already exists", "pod.name", pod.Name)
				response.Allowed = true
				respond(w, r, review, response)
				return
			}
		}

		// create a patch to inject a sidecar into the pod
		patch := []map[string]any{
			{
				"op":   "add",
				"path": "/spec/containers/-",
				"value": corev1.Container{
					Name:    "sidecar",
					Image:   "busybox",
					Command: []string{"sleep", "infinity"},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1m"),
							corev1.ResourceMemory: resource.MustParse("8Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("10m"),
							corev1.ResourceMemory: resource.MustParse("64Mi"),
						},
					},
				},
			},
		}

		// the patch should be a base64 encoded json string
		b, err := json.Marshal(&patch)
		if err != nil {
			requestLogger.Log("msg", "failed to marshal patch", "tag", "encode_failure", "error", err)
			http.Error(w, "Failed to marshal patch", http.StatusBadRequest)
			return
		}
		// []byte is serialized as a base64 string
		// so we only need to set the byte slice returned by json.Marshal
		response.Patch = b
		patchType := admissionv1.PatchTypeJSONPatch
		response.PatchType = &patchType

		// allow the pod to be admitted
		response.Allowed = true
		requestLogger.Log("msg", "sidecar injected", "tag", "injected", "pod.name", pod.Name)

		// send the response
		respond(w, r, review, response)
	}
}

// wrap the response in a new admission review using the same metadata as the original as well as the request uid
func respond(w http.ResponseWriter, r *http.Request, review *admissionv1.AdmissionReview, response *admissionv1.AdmissionResponse) {
	// ensure the same UID is used for the response
	response.UID = review.Request.UID
	// content type is json
	w.Header().Add("Content-Type", "application/json")
	// set the same meta info and the response
	json.NewEncoder(w).Encode(&admissionv1.AdmissionReview{
		TypeMeta: review.TypeMeta,
		Response: response,
	})
}
