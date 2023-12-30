package k8s

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func NewClient(namespace string) (*kubernetes.Clientset, error) {

	var kubeConfig *genericclioptions.ConfigFlags
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// Set properties manually from official rest config
	kubeConfig = genericclioptions.NewConfigFlags(false)
	kubeConfig.APIServer = &config.Host
	kubeConfig.BearerToken = &config.BearerToken
	kubeConfig.CAFile = &config.CAFile
	kubeConfig.Namespace = &namespace

	//var kubeconfig *string
	//if home := homedir.HomeDir(); home != "" {
	//	kubeconfig = flag.String("backchannel-kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	//} else {
	//	kubeconfig = flag.String("backchannel-kubeconfig", "", "absolute path to the kubeconfig file")
	//}
	//flag.Parse()

	////config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	//config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	//if err != nil {
	//	// If kubeconfig is not found, try to use InClusterConfig
	//	config, err = rest.InClusterConfig()
	//	if err != nil {
	//		return nil, err
	//	}
	//}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
