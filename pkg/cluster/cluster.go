package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	v1 "github.com/gtracer/overlord/api/v1"
	"github.com/gtracer/overlord/pkg/client"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Bootstrap ...
func Bootstrap(userID, clusterName string) error {
	client, err := client.New()
	if err != nil {
		return errors.Errorf("failed to get client %v", err)
	}

	nsName := types.NamespacedName{
		Namespace: userID,
		Name:      clusterName,
	}

	cluster := &v1.Cluster{}
	err = client.Get(context.TODO(), nsName, cluster)
	if err != nil {
		return errors.Errorf("failed to get cluster %s, %v", nsName.Name, err)
	}

	cluster.Spec.Bootstrap = true
	err = client.Update(context.TODO(), cluster)
	if err != nil {
		return err
	}

	return nil
}

// Report ...
func Report(userID, clusterName string) error {
	client, err := client.New()
	if err != nil {
		return errors.Errorf("failed to get client %v", err)
	}

	ns := &corev1.Namespace{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: userID}, ns)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Errorf("failed to get namespace %s, %v", userID, err)
		}
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: userID,
			},
		}
		err = client.Create(context.TODO(), ns)
		if err != nil {
			return err
		}
	}

	nsName := types.NamespacedName{
		Namespace: userID,
		Name:      clusterName,
	}
	cluster := &v1.Cluster{}
	err = client.Get(context.TODO(), nsName, cluster)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Errorf("failed to get cluster %s, %v", nsName.Name, err)
		}
		cluster = &v1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: nsName.Namespace,
				Name:      nsName.Name,
			},
			Spec: v1.ClusterSpec{
				Bootstrap: true,
			},
		}
		err = client.Create(context.TODO(), cluster)
		if err != nil {
			return err
		}
	}

	return nil
}

// List ...
func List() ([]byte, error) {
	client, err := client.New()
	if err != nil {
		return nil, errors.Errorf("failed to get client %v", err)
	}

	clusterList := &v1.ClusterList{}
	err = client.List(context.TODO(), clusterList)
	if err != nil {
		return nil, errors.Errorf("failed to list clusters %v", err)
	}
	var list []string

	for _, cluster := range clusterList.Items {
		list = append(list, fmt.Sprintf("%s/%s", cluster.Namespace, cluster.Name))
	}

	return json.Marshal(list)
}

// Kubeconfig ...
func Kubeconfig(userID, clusterName string) ([]byte, error) {
	client, err := client.New()
	if err != nil {
		return nil, errors.Errorf("failed to get client %v", err)
	}

	nsName := types.NamespacedName{
		Namespace: userID,
		Name:      clusterName,
	}
	cluster := &v1.Cluster{}
	err = client.Get(context.TODO(), nsName, cluster)
	if err != nil {
		return nil, errors.Errorf("failed to get kubeconfig %v", err)
	}

	kubeconfig := strings.Replace(cluster.Status.Kubeconfig, "localhost", cluster.Status.Master, -1)
	return []byte(kubeconfig), nil
}
