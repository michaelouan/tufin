package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	// Program Name is always the first (implicit) argument
	if len(os.Args) == 1 {
		fmt.Println("tufin cli use :")
		fmt.Println("     cluster        to deploy k3s cluster")
		fmt.Println("     deploy         to deploy Wordpress App")
		fmt.Println("     status         get status of pods in default namespace")
		os.Exit(1)
	} else {
		//cmd := os.Args[1]
		switch os.Args[1] {
		case "cluster":
			fmt.Printf("Deploy k3s cluster \n")
			app := "bash"
			args := []string{"-c", "curl -sfL https://get.k3s.io | sh -"}
			execCmd(app, args)
		case "deploy":
			deploy(getClientSet())
		case "status":
			getDefaultNamespacePodsStatus(getClientSet())
		case "--help":
			fmt.Println("tufin cli use :")
			fmt.Println("     cluster        to deploy k3s cluster")
			fmt.Println("     deploy         to deploy Wordpress App")
			fmt.Println("     status         get status of pods in default namespace")
		default:
			fmt.Println("Command not found use tufin --help to see disponible options")
			os.Exit(1)
		}
	}

}

func execCmd(app string, args []string) {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	cmd := exec.Command(app, args...)
	cmd.Dir = exPath

	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)

	cmd.Stdout = mw
	cmd.Stderr = mw

	// Execute the command
	if err := cmd.Run(); err != nil {
		log.Panic(err)
	}
}

func getClientSet() *kubernetes.Clientset {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join("/etc", "rancher", "k3s", "k3s.yaml"), "")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientset
}

func getDefaultNamespacePodsStatus(clientset *kubernetes.Clientset) {

	podClient, err := clientset.CoreV1().Pods(core.NamespaceDefault).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Println(err)
	}
	for _, podClient := range podClient.Items {
		fmt.Println(podClient.Name, podClient.Status.Phase)
	}
}
func deploy(clientset *kubernetes.Clientset) {

	wordpressPod := podWorpressStructure()
	sqlpod := podSqlStructure()
	pvc := getPersistentVolumeObj()
	svcSql := getServiceObject("sql", 3306, "ClusterIP")
	svcWP := getServiceObject("wordpress", 80, "NodePort")

	createPVC(clientset, pvc)
	createPod(clientset, wordpressPod)
	createPod(clientset, sqlpod)
	createService(clientset, svcSql)
	createService(clientset, svcWP)

}

func createPod(clientset *kubernetes.Clientset, pod *core.Pod) {

	fmt.Println("Creating pod %q.\n", pod.GetObjectMeta().GetName())
	pod, err := clientset.CoreV1().Pods(core.NamespaceDefault).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Created pod %q.\n", pod.GetObjectMeta().GetName())
	}
}

func createService(clientset *kubernetes.Clientset, svc *core.Service) {

	fmt.Println("Creating service...")
	svc, err := clientset.CoreV1().Services(core.NamespaceDefault).Create(context.TODO(), svc, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Created pod %q.\n", svc.GetObjectMeta().GetName())
	}
}

func createPVC(clientset *kubernetes.Clientset, pvc *core.PersistentVolumeClaim) {

	fmt.Println("Creating pvc %q.\n", pvc.GetObjectMeta().GetName())
	pvc, err := clientset.CoreV1().PersistentVolumeClaims(core.NamespaceDefault).Create(context.TODO(), pvc, metav1.CreateOptions{})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Created pvc %q.\n", pvc.GetObjectMeta().GetName())
	}
}

func podWorpressStructure() *core.Pod {
	return &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wordpress",
			Namespace: "default",
			Labels: map[string]string{
				"app": "wordpress",
			},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "wordpress",
					Image: "wordpress",
					Ports: []core.ContainerPort{
						{
							ContainerPort: 8080,
						},
					},
					Env: []core.EnvVar{
						{
							Name:  "WORDPRESS_DB_HOST",
							Value: "sql",
						},
						{
							Name:  "WORDPRESS_DB_USER",
							Value: "wordpress",
						},
						{
							Name:  "WORDPRESS_DB_PASSWORD",
							Value: "wordpress",
						},
						{
							Name:  "WORDPRESS_DB_NAME",
							Value: "wordpress",
						},
					},
				},
			},
		},
	}
}

func podSqlStructure() *core.Pod {
	return &core.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sql",
			Namespace: "default",
			Labels: map[string]string{
				"app": "sql",
			},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "sql",
					Image: "mysql:5.7",
					Ports: []core.ContainerPort{
						{
							ContainerPort: 3306,
						},
					},
					Env: []core.EnvVar{
						{
							Name:  "MYSQL_ROOT_PASSWORD",
							Value: "somewordpress",
						},
						{
							Name:  "MYSQL_DATABASE",
							Value: "wordpress",
						},
						{
							Name:  "MYSQL_USER",
							Value: "wordpress",
						},
						{
							Name:  "MYSQL_PASSWORD",
							Value: "wordpress",
						},
					},
					VolumeMounts: []core.VolumeMount{
						{
							Name:      "sqlvol",
							MountPath: "/var/lib/mysql",
						},
					},
				},
			},
			Volumes: []core.Volume{
				{
					Name: "sqlvol",
					VolumeSource: core.VolumeSource{
						PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
							ClaimName: "sql-pv-claim",
						},
					},
				},
			},
		},
	}
}

func getServiceObject(name string, port int32, serviceTypee core.ServiceType) *core.Service {
	return &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				{
					Port: port,
				},
			},
			Selector: map[string]string{
				"app": name,
			},
			Type: serviceTypee,
		},
	}
}

func getPersistentVolumeObj() *core.PersistentVolumeClaim {
	return &core.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sql-pv-claim",
			Namespace: "default",
			Labels: map[string]string{
				"app": "sql",
			},
		},
		Spec: core.PersistentVolumeClaimSpec{
			AccessModes: []core.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: core.ResourceRequirements{
				Requests: core.ResourceList{
					core.ResourceName(core.ResourceStorage): resource.MustParse("20Gi"),
				},
			},
		},
	}
}
