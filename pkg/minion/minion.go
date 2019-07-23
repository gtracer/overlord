package minion

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	v1 "github.com/gtracer/overlord/api/v1"
	"github.com/gtracer/overlord/pkg/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Status ...
func Status(minionName string, r *http.Request) (*v1.NodeStatus, error) {
	//Read all the data in r.Body from a byte[], convert it to a string, and assign store it in 's'.
	s, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Errorf("unable to read http body for %s, %v", minionName, err)
	}
	jsonData := &v1.NodeStatus{}
	// use the built in Unmarshal function
	err = json.Unmarshal(s, jsonData)
	if err != nil {
		return nil, errors.Errorf("unable to unmarshal json body %s, %v", minionName, err)
	}

	return jsonData, nil
}

// Report ...
func Report(userID, id, minionName string, minionStatus *v1.NodeStatus) (string, error) {
	client, err := client.New()
	if err != nil {
		return "", errors.Errorf("failed to get client %v", err)
	}

	minion := &v1.Minion{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: minionName, Namespace: userID}, minion)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return "", errors.Errorf("failed to get minion %s, %v", minionName, err)
		}
		minion = &v1.Minion{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: userID,
				Name:      minionName,
				Labels: map[string]string{
					"kubernetes.ov3rlord.me/cluster": id,
				},
			},
			Spec: v1.MinionSpec{},
			Status: v1.MinionStatus{
				NodeStatus: *minionStatus,
			},
		}
		err = client.Create(context.TODO(), minion)
		if err != nil {
			return "", errors.Errorf("failed to create minion %s, %v", minionName, err)
		}
		return "", nil
	}

	minion.Status.NodeStatus = *minionStatus
	err = client.Status().Update(context.TODO(), minion)
	if err != nil {
		return "", errors.Errorf("failed to update minion status %s, %v", minionName, err)
	}

	return minion.Spec.Master, nil
}
