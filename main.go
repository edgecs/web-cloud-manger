/*
 * ECS API
 *
 * ECS API
 *
 * API version: 0.0.1
 * Contact: eiffel.zhou@gmail.com
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	imp "./go"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	portPtr := flag.Int("port", 8080, "server port")
	port := strconv.Itoa(*portPtr)
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	imp.InitClientConfig(config)

	router := imp.NewRouter()

	log.Printf("Server starts port:" + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
