package main
import "C"
import (
	"context"
	"flag"
	"fmt"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"github.com/tidwall/gjson"
	"strconv"
)
type Template struct {
	Min int32
	Max int32
	AppName string
	NameSp string
	K8sConfig K8sconfig
	PodNmuber int32
}
type K8sconfig struct {
	ClientSet *kubernetes.Clientset
}
func init(){
	fmt.Println("初始化config文件")
}
func (t *Template)GetAppName(){
	//fmt.Println("请输入要扩展的APP Name:")
	//fmt.Scanln(&t.AppName)
	if t.AppName != "" {
	 t.GetNameSpace()
	}else {
		fmt.Println("请输入正确的APP Name")
		os.Exit(3)
	}
}
func (t *Template)GetValue(){
	//fmt.Println("请输入要扩展的pod个数:")
	//fmt.Scanln(t.PodNmuber)
	if t.PodNmuber <= 10 {
		t.Min=t.PodNmuber
		t.Max=t.PodNmuber
		t.AutoScale()
	}else {
		fmt.Println("请输入11以下数字，如不满足需求请联系 SE Team, TKS")
		os.Exit(3)
	}
}
func (t *Template)GetNameSpace(){
	namespaceList, err := t.K8sConfig.ClientSet.CoreV1().Namespaces().List(context.TODO(),metav1.ListOptions{})
	if err != nil {
		fmt.Println(err.Error())
	}else{
		for _,nsList := range namespaceList.Items {
			nsList1,_:=json.Marshal(nsList)
			nsList2:=gjson.Get(string(nsList1),"metadata.name").String()
			_, err := t.K8sConfig.ClientSet.AppsV1().Deployments(nsList2).Get(context.TODO(),t.AppName, metav1.GetOptions{})
			if err == nil {
				t.NameSp=nsList2
				//fmt.Println("已匹配到。。。Namespace =",nsList2)
			}else{
				continue
			}
		}
	}
}
func (t *Template)AutoScale(){
	hpa, _ := t.K8sConfig.ClientSet.AutoscalingV1().HorizontalPodAutoscalers(t.NameSp).Get(context.TODO(), t.AppName, metav1.GetOptions{})
	//fmt.Println(hpa.Spec.MaxReplicas)
	fmt.Println("未更新前的pod个数为：",*hpa.Spec.MinReplicas)
	hpa.Spec.MaxReplicas = t.Max
	hpa.Spec.MinReplicas = &t.Min
	_, _ = t.K8sConfig.ClientSet.AutoscalingV1().HorizontalPodAutoscalers(t.NameSp).Update(context.TODO(), hpa, metav1.UpdateOptions{})
	hpa, _ = t.K8sConfig.ClientSet.AutoscalingV1().HorizontalPodAutoscalers(t.NameSp).Get(context.TODO(), t.AppName, metav1.GetOptions{})
	//fmt.Println(hpa.Spec.MaxReplicas)
	fmt.Println("更新后的pod个数为：",*hpa.Spec.MinReplicas)
}
func main() {
	t:=new(Template)
	kubeconfig := flag.String("kubeconfig", filepath.Join("/home/", "latest.yaml"), "(optional) absolute path to the kubeconfig file")
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("https://K8SApiserverip:443", *kubeconfig)
	if err != nil {
		panic(err)
	}
	t.K8sConfig.ClientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	args := os.Args
	t.AppName=args[1]
	podnummber,_:=strconv.ParseInt(args[2],10,32)
	t.PodNmuber=int32(podnummber)
	t.GetAppName()
	t.GetValue()
}
