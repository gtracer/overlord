package minion

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"

	v1 "github.com/gtracer/overlord/api/v1"
	"github.com/gtracer/overlord/pkg/client"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type minion struct {
	Name    string `json:"name,omitempty"`
	Role    string `json:"role,omitempty"`
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

// Status ...
func Status(id, minionID string, r *http.Request) (*v1.NodeStatus, error) {
	//Read all the data in r.Body from a byte[], convert it to a string, and assign store it in 's'.
	s, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Errorf("unable to read http body for %s/%s, %v", id, minionID, err)
	}
	jsonData := &v1.NodeStatus{}
	// use the built in Unmarshal function
	err = json.Unmarshal(s, jsonData)
	if err != nil {
		return nil, errors.Errorf("unable to unmarshal json body %s/%s, %v", id, minionID, err)
	}

	return jsonData, nil
}

// Report ...
func Report(userID, id, minionID string, minionStatus *v1.NodeStatus) ([]byte, error) {
	minionName := fmt.Sprintf("%s-%s", id, minionID)
	client, err := client.New()
	if err != nil {
		return nil, errors.Errorf("failed to get client %v", err)
	}

	minion := &v1.Minion{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: minionName, Namespace: userID}, minion)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, errors.Errorf("failed to get minion %s, %v", minionName, err)
		}
		minion = &v1.Minion{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: userID,
				Name:      minionName,
				Labels: map[string]string{
					"kubernetes.ov3rlord.me/cluster": id,
				},
			},
			Spec: v1.MinionSpec{
				Name: minionID,
			},
			Status: v1.MinionStatus{
				NodeStatus: *minionStatus,
			},
		}
		err = client.Create(context.TODO(), minion)
		if err != nil {
			return nil, errors.Errorf("failed to create minion %s, %v", minionName, err)
		}
		return nil, nil
	}

	minion.Status.NodeStatus = *minionStatus
	minion.Status.LastTimestamp.Time = time.Now()
	err = client.Status().Update(context.TODO(), minion)
	if err != nil {
		return nil, errors.Errorf("failed to update minion status %s, %v", minionName, err)
	}

	return json.Marshal(minion.Spec)
}

// List ...
func List(userID, id string) ([]byte, error) {
	client, err := client.New()
	if err != nil {
		return nil, errors.Errorf("failed to get client %v", err)
	}

	matchingLabels := ctrlclient.MatchingLabels(
		map[string]string{
			"kubernetes.ov3rlord.me/cluster": id,
		},
	)

	minionList := &v1.MinionList{}
	err = client.List(context.TODO(), minionList, ctrlclient.InNamespace(userID), matchingLabels)
	if err != nil {
		return nil, errors.Errorf("failed to list minions %v", err)
	}
	var list []minion

	for _, minionListItem := range minionList.Items {
		role := "Agent"
		if minionListItem.Spec.Master == minionListItem.Spec.Name {
			role = "Master"
		}
		minionItem := minion{
			Name:    minionListItem.Spec.Name,
			Role:    role,
			Status:  string(minionListItem.Status.State),
			Message: minionListItem.Status.Message,
		}
		list = append(list, minionItem)
	}

	return json.Marshal(list)
}
