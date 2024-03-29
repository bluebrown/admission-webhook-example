package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
)

func TestAdmission(t *testing.T) {
	req := httptest.NewRequest("POST", "/inject-sidecar", bytes.NewReader([]byte(reviewJson)))
	rr := httptest.NewRecorder()
	handler := handleInjectSidecar()
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	review := &admissionv1.AdmissionReview{}
	if err := json.NewDecoder(rr.Body).Decode(review); err != nil {
		t.Errorf("Failed to decode response body: %s", err)
	}

	if !review.Response.Allowed {
		t.Errorf("Response not allowed")
	}

	if review.Response.UID != "e82c9f94-691e-4dbc-a215-f24a4f09a2ac" {
		t.Errorf("Response UID incorrect")
	}

	var patch []map[string]any
	if err := json.Unmarshal(review.Response.Patch, &patch); err != nil {
		t.Errorf("Failed to unmarshal patch: %s", err)
	}

	if len(patch) != 1 {
		t.Errorf("Patch should have one element")
	}

	if patch[0]["op"] != "add" {
		t.Errorf("Patch op should be add")
	}

	if patch[0]["path"] != "/spec/containers/-" {
		t.Errorf("Patch path should be /spec/containers/-")
	}

	container, ok := patch[0]["value"].(map[string]any)
	if !ok {
		t.Errorf("Patch value should be a map")
	}

	if container["name"] != "sidecar" {
		t.Errorf("Patch value name should be sidecar")
	}

}

var reviewJson = `{
    "apiVersion": "admission.k8s.io/v1",
    "kind": "AdmissionReview",
    "request": {
        "dryRun": false,
        "kind": {
            "group": "",
            "kind": "Pod",
            "version": "v1"
        },
        "name": "busybox",
        "namespace": "example",
        "object": {
            "apiVersion": "v1",
            "kind": "Pod",
            "metadata": {
                "labels": {
                    "inject-sidecar-example": "enabled",
                    "name": "busybox"
                },
                "name": "busybox",
                "namespace": "example"
            },
            "spec": {
                "containers": [
                    {
                        "command": [
                            "sleep",
                            "infinity"
                        ],
                        "image": "busybox",
                        "imagePullPolicy": "Always",
                        "name": "busybox",
                        "resources": {
                            "limits": {
                                "cpu": "10m",
                                "memory": "64Mi"
                            },
                            "requests": {
                                "cpu": "1m",
                                "memory": "8Mi"
                            }
                        }
                    }
                ]
            }
        },
        "oldObject": null,
        "operation": "CREATE",
        "options": {
            "apiVersion": "meta.k8s.io/v1",
            "fieldManager": "kubectl-client-side-apply",
            "kind": "CreateOptions"
        },
        "requestKind": {
            "group": "",
            "kind": "Pod",
            "version": "v1"
        },
        "requestResource": {
            "group": "",
            "resource": "pods",
            "version": "v1"
        },
        "resource": {
            "group": "",
            "resource": "pods",
            "version": "v1"
        },
        "uid": "e82c9f94-691e-4dbc-a215-f24a4f09a2ac",
        "userInfo": {
            "extra": {
                "oid": [
                    "bb088c9e-3dbc-470f-8175-6dd6b25ed13a"
                ]
            },
            "groups": [
                "system:authenticated"
            ],
            "username": "bluebrown"
        }
    }
}`
