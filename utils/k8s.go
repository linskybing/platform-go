package utils

import (
	"context"
	applyJson "encoding/json"
	"fmt"
	"log"

	"github.com/linskybing/platform-go/config"
	"github.com/linskybing/platform-go/k8sclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func ValidateK8sJSON(jsonStr string) (*schema.GroupVersionKind, string, error) {
	decoder := json.NewSerializerWithOptions(
		json.DefaultMetaFactory, config.Scheme, config.Scheme,
		json.SerializerOptions{Yaml: false, Pretty: false, Strict: true},
	)
	obj, gvk, err := decoder.Decode([]byte(jsonStr), nil, nil)
	if err != nil {
		return nil, "", err
	}
	if obj == nil {
		return nil, "", fmt.Errorf("decoded object is nil")
	}

	metaObj, ok := obj.(metav1.Object)
	if !ok {
		return nil, "", fmt.Errorf("object does not implement metav1.Object interface")
	}

	return gvk, metaObj.GetName(), nil
}

func CreateByJson(jsonStr []byte, ns string) error {
	// decode
	var obj unstructured.Unstructured
	if err := applyJson.Unmarshal(jsonStr, &obj.Object); err != nil {
		return err
	}

	gvk := obj.GroupVersionKind()
	mapping, err := k8sclient.Mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	if ns == "" {
		ns = "default"
	}
	resourceClient := k8sclient.DynamicClient.Resource(mapping.Resource).Namespace(ns)
	result, err := resourceClient.Create(context.TODO(), &obj, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("✅ Created %s/%s\n", result.GetKind(), result.GetName())
	return nil
}

func DeleteByJson(jsonStr []byte, ns string) error {
	// decode
	var obj unstructured.Unstructured
	if err := applyJson.Unmarshal(jsonStr, &obj.Object); err != nil {
		return err
	}

	gvk := obj.GroupVersionKind()
	mapping, err := k8sclient.Mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	if ns == "" {
		ns = "default"
	}
	resourceClient := k8sclient.DynamicClient.Resource(mapping.Resource).Namespace(ns)
	err = resourceClient.Delete(context.TODO(), obj.GetName(), metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func UpdateByJson(jsonStr []byte, ns string) error {
	// decode
	var obj unstructured.Unstructured
	if err := applyJson.Unmarshal(jsonStr, &obj.Object); err != nil {
		return err
	}

	gvk := obj.GroupVersionKind()
	mapping, err := k8sclient.Mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	if ns == "" {
		ns = "default"
	}
	resourceClient := k8sclient.DynamicClient.Resource(mapping.Resource).Namespace(ns)
	result, err := resourceClient.Update(context.TODO(), &obj, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("✅ Updated %s/%s\n", result.GetKind(), result.GetName())
	return nil
}

func CreateNamespace(name string) {
	_, err := k8sclient.Clientset.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
	if err == nil {
		fmt.Printf("namespace %s already exist \n", name)
		return
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	_, err = k8sclient.Clientset.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("failed create namespace: %v", err)
	}

	fmt.Printf("create Namespace: %s successfully\n", name)
}

func resourceMustParse(size string) resource.Quantity {
	q, err := resource.ParseQuantity(size)
	if err != nil {
		panic(fmt.Errorf("invalid PVC size: %v", err))
	}
	return q
}

func ExpandPVC(ns, pvcName, newSize string) error {
	if ns == "" {
		ns = "default"
	}

	client := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(ns)

	// 先取得現有 PVC
	pvc, err := client.Get(context.TODO(), pvcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get PVC: %w", err)
	}

	// 更新 storage requests
	pvc.Spec.Resources.Requests[corev1.ResourceStorage] = resource.MustParse(newSize)

	// 傳回 API server 更新
	_, err = client.Update(context.TODO(), pvc, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to expand PVC: %w", err)
	}

	fmt.Printf("✅ PVC %s in namespace %s expanded to %s\n", pvcName, ns, newSize)
	return nil
}

func CreatePVC(ns string, name string, storageClassName string, size string) error {
	if ns == "" {
		ns = "default"
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resourceMustParse(size),
				},
			},
			StorageClassName: &storageClassName,
		},
	}

	client := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(ns)

	_, err := client.Get(context.TODO(), name, metav1.GetOptions{})
	if err == nil {
		fmt.Printf("PVC %s already exists in namespace %s\n", name, ns)
		return nil
	}

	result, err := client.Create(context.TODO(), pvc, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create PVC: %w", err)
	}

	fmt.Printf("✅ PVC %s created in namespace %s\n", result.Name, ns)
	return nil
}

func DeletePVC(ns string, pvcName string) error {
	if ns == "" {
		ns = "default"
	}

	client := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(ns)

	policy := metav1.DeletePropagationForeground
	err := client.Delete(context.TODO(), pvcName, metav1.DeleteOptions{
		PropagationPolicy: &policy,
	})
	if err != nil {
		return fmt.Errorf("failed to delete PVC: %w", err)
	}

	fmt.Printf("🗑️ PVC %s deleted from namespace %s\n", pvcName, ns)
	return nil
}

func GetPVC(ns string, pvcName string) (*corev1.PersistentVolumeClaim, error) {
	if ns == "" {
		ns = "default"
	}
	client := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(ns)
	pvc, err := client.Get(context.TODO(), pvcName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get PVC: %w", err)
	}
	fmt.Printf("🔍 PVC %s in namespace %s: %s, size: %v\n",
		pvc.Name, ns, pvc.Status.Phase, pvc.Spec.Resources.Requests[corev1.ResourceStorage])

	return pvc, nil
}

func ListPVCs(ns string) ([]corev1.PersistentVolumeClaim, error) {
	if ns == "" {
		ns = "default"
	}
	client := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(ns)
	pvcList, err := client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list PVCs: %w", err)
	}

	for _, pvc := range pvcList.Items {
		q := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
		fmt.Println(q.String())
	}

	return pvcList.Items, nil
}
