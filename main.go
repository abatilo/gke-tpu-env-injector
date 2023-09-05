package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func extractStatefulSet(admissionReview *admissionv1.AdmissionReview) (*appsv1.StatefulSet, error) {
	if admissionReview.Request.Kind.Kind != "StatefulSet" {
		return nil, fmt.Errorf("Expected StatefulSet but got %s", admissionReview.Request.Kind.Kind)
	}

	statefulset := appsv1.StatefulSet{}
	if err := json.Unmarshal(admissionReview.Request.Object.Raw, &statefulset); err != nil {
		return nil, err
	}

	return &statefulset, nil
}

// mutateStatefulSet is a mutating webhook that adds environment variables to a
// StatefulSet's containers that contain the hostnames of all other replicas in
// the StatefulSet. This is useful for distributed training jobs that need to
// know the hostnames of all other replicas in the job.
func mutateStatefulSet(
	admissionReview *admissionv1.AdmissionReview,
) (*admissionv1.AdmissionResponse, error) {
	statefulset, _ := extractStatefulSet(admissionReview)

	statefulsetName := statefulset.Name
	serviceName := statefulset.Spec.ServiceName
	replicas := statefulset.Spec.Replicas

	hostNames := make([]string, *replicas)
	for i := 0; i < int(*replicas); i++ {
		hostNames[i] = fmt.Sprintf("%s-%d.%s", statefulsetName, i, serviceName)
	}
	joinedHostNames := strings.Join(hostNames, ",")

	patches := []map[string]interface{}{}
	for i := 0; i < len(statefulset.Spec.Template.Spec.Containers); i++ {
		patch := map[string]interface{}{
			"op": "add",
		}
		container := statefulset.Spec.Template.Spec.Containers[i]
		path := fmt.Sprintf("/spec/template/spec/containers/%d/env", i)
		value := corev1.EnvVar{
			Name:  "TPU_WORKER_HOSTNAMES",
			Value: joinedHostNames,
		}

		if len(container.Env) == 0 {
			// If there aren't any environment variables, set env to an array
			patch["path"] = path
			patch["value"] = []corev1.EnvVar{value}
		} else {
			// When there are already environment variables, append to the array with
			// a single item
			patch["path"] = fmt.Sprintf("%s/-", path)
			patch["value"] = value
		}
		patches = append(patches, patch)
	}
	patchBytes, _ := json.Marshal(patches)

	// Create AdmissionResponse
	admissionResponse := &admissionv1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}(),
	}
	return admissionResponse, nil
}

func extractPod(admissionReview *admissionv1.AdmissionReview) (*corev1.Pod, error) {
	if admissionReview.Request.Kind.Kind != "Pod" {
		return nil, fmt.Errorf("Expected Pod but got %s", admissionReview.Request.Kind.Kind)
	}

	pod := corev1.Pod{}
	if err := json.Unmarshal(admissionReview.Request.Object.Raw, &pod); err != nil {
		return nil, err
	}

	return &pod, nil
}

// mutatePod is a mutating webhook that adds environment variables to a Pod's
// containers that gets the ordinal index of the pod in the StatefulSet. This
// is useful for distributed training jobs that need to know the ordinal index
// of the pod in the job.
func mutatePod(
	admissionReview *admissionv1.AdmissionReview,
) (*admissionv1.AdmissionResponse, error) {
	pod, _ := extractPod(admissionReview)

	// Verify that the pod is part of a StatefulSet
	if pod.OwnerReferences[0].Kind != "StatefulSet" {
		return nil, fmt.Errorf(
			"Expected Pod to be part of a StatefulSet but got %s",
			pod.OwnerReferences[0].Kind,
		)
	}

	podName := pod.Labels["statefulset.kubernetes.io/pod-name"]
	ordinalIndex := podName[strings.LastIndex(podName, "-")+1:]

	patches := []map[string]interface{}{}

	for i := 0; i < len(pod.Spec.Containers); i++ {
		// Add the TPU_WORKER_ID environment variable
		patch := map[string]interface{}{
			"op":   "add",
			"path": fmt.Sprintf("/spec/containers/%d/env/-", i),
			"value": corev1.EnvVar{
				Name:  "TPU_WORKER_ID",
				Value: ordinalIndex,
			},
		}
		patches = append(patches, patch)
	}
	patchBytes, _ := json.Marshal(patches)

	admissionResponse := &admissionv1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}(),
	}
	return admissionResponse, nil
}

func main() {
	ctx := context.Background()

	viper.SetEnvPrefix("GTEI")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	pflag.String(
		FlagTLSCertFile,
		"/etc/tls/tls.crt",
		"File containing the default x509 Certificate for HTTPS.",
	)
	pflag.String(
		FlagTLSKeyFile,
		"/etc/tls/tls.key",
		"File containing the default x509 private key matching --tls-cert-file.",
	)
	pflag.Bool(
		FlagVerbose,
		false,
		"Enable verbose logging.",
	)

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	tlsCertFilePath := viper.GetString(FlagTLSCertFile)
	tlsKeyFilePath := viper.GetString(FlagTLSKeyFile)
	verbose := viper.GetBool(FlagVerbose)

	if verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Verbose logging enabled")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "gke-tpu-env-injector")
	})
	mux.HandleFunc("/mutate", func(w http.ResponseWriter, r *http.Request) {
		admissionReview := &admissionv1.AdmissionReview{}
		if err := json.NewDecoder(r.Body).Decode(admissionReview); err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}

		if admissionReview.Request.Kind.Kind == "StatefulSet" {
			log.Debug().Msg("Received review for StatefulSet")
			admissionReview.Response, _ = mutateStatefulSet(admissionReview)
			responseBytes, _ := json.Marshal(admissionReview)
			fmt.Fprint(w, string(responseBytes))
			return
		}

		if admissionReview.Request.Kind.Kind == "Pod" {
			log.Debug().Msg("Received review for Pod")
			admissionReview.Response, _ = mutatePod(admissionReview)
			responseBytes, _ := json.Marshal(admissionReview)
			fmt.Fprint(w, string(responseBytes))
			return
		}
	})

	srv := &http.Server{
		Addr:    ":443",
		Handler: mux,
	}

	// Register signal handlers for graceful shutdown
	done := make(chan struct{})
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Info().Msg("Shutting down gracefully")
		_ = srv.Shutdown(ctx)
		close(done)
	}()

	log.Info().Msg("Starting server on port 443")
	if err := srv.ListenAndServeTLS(tlsCertFilePath, tlsKeyFilePath); err != nil {
		if err == http.ErrServerClosed {
			log.Info().Msg("Server closed")
			return
		}
		log.Fatal().Err(err).Msg("Failed to start server")
	}
	<-done
}
