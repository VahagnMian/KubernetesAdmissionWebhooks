package main

import (
	"encoding/json"
	"errors"
	"fmt"
	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/admission/v1"
	"k8s.io/api/admission/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"flag"
	"io/ioutil"
	"strconv"
)

var config *rest.Config
var clientSet *kubernetes.Clientset

type ServerParameters struct {
	port     int
	certFile string
	keyFile  string
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

var parameters ServerParameters

var (
	universalDeserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
)

func main() {

	useKubeConfig := os.Getenv("USE_KUBECONFIG")
	kubeConfigFilePath := os.Getenv("KUBECONFIG")

	flag.IntVar(&parameters.port, "port", 8443, "Webhook server port.")
  
	if useKubeConfig == "true" {
		fmt.Println("Using Kubeconfig")
		fmt.Println("Using local certs")
		flag.StringVar(&parameters.certFile, "tlsCertFile", "../tls/local-dev-certs/tls.crt", "File containing the x509 Certificate for HTTPS.")
		flag.StringVar(&parameters.keyFile, "tlsKeyFile", "../tls/local-dev-certs/tls.key", "File containing the x509 private key to --tlsCertFile.")
		flag.Parse()
	} else {
		flag.StringVar(&parameters.certFile, "tlsCertFile", "/etc/webhook/certs/tls.crt", "File containing the x509 Certificate for HTTPS.")
		flag.StringVar(&parameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/tls.key", "File containing the x509 private key to --tlsCertFile.")
		flag.Parse()
	}

	if len(useKubeConfig) == 0 {
		// default to service account in cluster token
		c, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		config = c
	} else {
		//load from a kube config
		var kubeconfig string

		if kubeConfigFilePath == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = filepath.Join("/Users/vahagn/", ".kube", "config31")
			}
		} else {
			kubeconfig = kubeConfigFilePath
		}

		fmt.Println("kubeconfig: " + kubeconfig)

		c, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		config = c
	}

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	clientSet = cs

	test()
	http.HandleFunc("/", HandleRoot)
	http.HandleFunc("/mutate", HandleMutate)
	http.HandleFunc("/validate", HandleValidate)
	log.Fatal(http.ListenAndServeTLS(":"+strconv.Itoa(parameters.port), parameters.certFile, parameters.keyFile, nil))
}

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("HandleRoot!"))
}

func HandleMutate(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	err = ioutil.WriteFile("/tmp/request", body, 0644)
	if err != nil {
		panic(err.Error())
	}

	var admissionReviewReq v1.AdmissionReview

	if _, _, err := universalDeserializer.Decode(body, nil, &admissionReviewReq); err != nil {
		// check if decoder returnes any error
		w.WriteHeader(http.StatusBadRequest)
		fmt.Errorf("could not deserialize request: %v", err)
	} else if admissionReviewReq.Request == nil {
		// Check if Request from API server is nil or not
		w.WriteHeader(http.StatusBadRequest)
		errors.New("malformed admission review: request is nil")
	}

	// Print logs for webhook
	fmt.Printf("Type: %v \t Event: %v \t Name: %v \n",
		admissionReviewReq.Request.Kind,
		admissionReviewReq.Request.Operation,
		admissionReviewReq.Request.Name,
	)

	var pod apiv1.Pod

	err = json.Unmarshal(admissionReviewReq.Request.Object.Raw, &pod)

	if err != nil {
		fmt.Errorf("could not unmarshal pod on admission request: %v", err)
	}

	var patches []patchOperation

	labels := pod.ObjectMeta.Labels
	if pod.Namespace == getenv("TARGET_NAMESPACE", "prod") {
		labels["deletion-protection"] = "true"
	}

	patches = append(patches, patchOperation{
		Op:    "add",
		Path:  "/metadata/labels",
		Value: labels,
	})

	patchBytes, err := json.Marshal(patches)
	if err != nil {
		fmt.Errorf("could not marshal JSON patch: %v", err)
	}

	admissionReviewResponse := v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			UID:     admissionReviewReq.Request.UID,
			Allowed: true,
		},
	}
	admissionReviewResponse.Response.Patch = patchBytes

	bytes, err := json.Marshal(&admissionReviewResponse)
	if err != nil {
		fmt.Errorf("marshaling response: %v", err)
	}

	w.Write(bytes)

}

func HandleValidate(w http.ResponseWriter, r *http.Request) {

	input := admissionv1.AdmissionReview{}

	body, err := ioutil.ReadAll(r.Body)
	fmt.Println(string(body))
	err = ioutil.WriteFile("/tmp/request", body, 0644)
	if err != nil {
		panic(err.Error())
	}

	if _, _, err := universalDeserializer.Decode(body, nil, &input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Errorf("could not deserialize request: %v", err)
	} else if input.Request == nil {
		w.WriteHeader(http.StatusBadRequest)
		errors.New("malformed admission review: request is nil")
	}

	operation := input.Request.Operation
	var pod apiv1.Pod
	var isAllowed bool
	var Message string

	// unmarshalling request from API server to patch
	err = json.Unmarshal(input.Request.OldObject.Raw, &pod)
	fmt.Println("Request: ", input.Request.Object)

	labels := pod.ObjectMeta.Labels
	fmt.Println(labels)
	if labels["deletion-protection"] == "true" && operation == "DELETE" && pod.Namespace == getenv("TARGET_NAMESPACE", "prod") {
		isAllowed = false
		Message = "Deletion not allowed on this object"
	} else {
		isAllowed = true
		Message = "Deletion allowed on this object"
	}

	fmt.Println("Called", operation, "on", input.Request.Name, "kind of", input.Request.Kind)

	err = json.Unmarshal(input.Request.Object.Raw, &operation)

	if err != nil {
		fmt.Errorf("Could not unmarshal pod on admission request: %v", err)
	}

	output := admissionv1.AdmissionReview{

		Response: &admissionv1.AdmissionResponse{
			UID:     input.Request.UID,
			Allowed: isAllowed,
			Result: &metav1.Status{
				Message: Message,
			},
		},
	}

	output.TypeMeta.Kind = input.TypeMeta.Kind
	output.TypeMeta.APIVersion = input.TypeMeta.APIVersion

	bytes, err := json.Marshal(output)
	if err != nil {
		fmt.Errorf("marshaling response: %v", err)
	}

	fmt.Println(string(bytes))
	w.Write(bytes)

}
