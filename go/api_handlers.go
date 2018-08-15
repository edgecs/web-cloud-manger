/*
 * ECS API
 *
 * ECS API Server Implementation
 *
 * API version: 0.0.1
 * Contact: eiffel.zhou@gmail.com
 */

package swagger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"
	// "k8s.io/client-go/util/retry"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	//config clientcmd.ClientConfig
	config *restclient.Config
	tasks  = make(map[string]*Task)
)

// InitClientConfig ...
func InitClientConfig(cfg *restclient.Config) {
	config = cfg
}

// AddTask Create a new task
func AddTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var task Task
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		Logger2(400, "invalid body")
		return
	}
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), 400)
		Logger2(400, "invalid body")
		return
	}

	if _, ok := tasks[task.Name]; ok {
		http.Error(w, "Duplicated task name", 400)
		Logger2(400, "Duplicated task name")
		return
	}

	tasks[task.Name] = &task

	createDeployment(task.Name, task.Image, task.Cpu, task.Memory)
	if task.Duration > 0 {
		go timeoutStopDeployment(task.Duration, task.Name)
	}
	w.WriteHeader(http.StatusOK)
}

// ListTasks list all tasks
func ListTasks(w http.ResponseWriter, r *http.Request) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	list, err := deploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Printf("%v\n", d)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}

func parseResource(cpu, memory string) apiv1.ResourceRequirements {
	resList := apiv1.ResourceList{}
	if cpu != "" {
		resList[apiv1.ResourceCPU] = resource.MustParse(cpu)
	}
	if memory != "" {
		resList[apiv1.ResourceMemory] = resource.MustParse(memory)
	}
	res := apiv1.ResourceRequirements{}
	res.Requests = resList
	res.Limits = resList
	return res
}

func createDeployment(name, imageName, cpu, memory string) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "ecs-demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "ecs-demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:      name,
							Image:     imageName,
							Resources: parseResource(cpu, memory),
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
}

func int32Ptr(i int32) *int32 { return &i }

func timeoutStopDeployment(duration int32, taskName string) {
	timer := time.NewTimer(time.Duration(duration) * time.Second)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	<-timer.C
	fmt.Println("Timer expired, stopping task ", taskName)
	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete(taskName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
	fmt.Println("Deleted deployment.")

}
