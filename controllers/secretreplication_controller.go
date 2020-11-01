/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	replicationv1 "github.com/thomas-maurice/secret-replicator/api/v1"
)

const (
	// ControllerName is the name of the controller
	ControllerName = "secret-replicator"
)

var (
	sourceObjectAnnotation  = "secretreplications.replication.apis.maurice.fr/source-object"
	sourceVersionAnnotation = "secretreplications.replication.apis.maurice.fr/source-version"
	ownerKey                = ".metadata.controller"
	apiGVStr                = replicationv1.GroupVersion.String()
)

// SecretReplicationReconciler reconciles a SecretReplication object
type SecretReplicationReconciler struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	RequeueDuration time.Duration
	EventRecorder   record.EventRecorder
}

// +kubebuilder:rbac:groups=replication.apis.maurice.fr,resources=secretreplications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=replication.apis.maurice.fr,resources=secretreplications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=core,resources=secrets/status,verbs=get

// Reconcile is the main reconciliation loop
func (r *SecretReplicationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("secretreplication", req.NamespacedName)
	result := ctrl.Result{RequeueAfter: r.RequeueDuration}

	var secretReplication replicationv1.SecretReplication
	if err := r.Get(ctx, req.NamespacedName, &secretReplication); err != nil {
		if !apierrs.IsNotFound(err) {
			log.Error(err, "Unable to fetch SecretReplication")
		}
		return result, ignoreNotFound(err)
	}

	sourceSecretName := types.NamespacedName{
		Namespace: secretReplication.Spec.SrcNamespace,
		Name:      secretReplication.Spec.SrcName,
	}

	var srcSecret corev1.Secret
	if err := r.Get(ctx, sourceSecretName, &srcSecret); err != nil {
		log.Error(err, "Unable to fetch source secret")
		r.EventRecorder.Event(&secretReplication, corev1.EventTypeWarning, "Unable to fetch the source secret", fmt.Sprintf("%s", err))
		return result, err
	}

	var dstSecret corev1.Secret
	dstSecretName := types.NamespacedName{
		Namespace: secretReplication.Spec.DstNamespace,
		Name:      secretReplication.Spec.DstName,
	}

	if err := r.Get(ctx, dstSecretName, &dstSecret); err != nil {
		if apierrs.IsNotFound(err) {
			log.Info("Creating a new replicated secret")
			dstSecret.Type = srcSecret.Type
			dstSecret.Data = srcSecret.Data
			dstSecret.Name = secretReplication.Spec.DstName
			dstSecret.Namespace = secretReplication.Spec.DstNamespace
			dstSecret.Annotations = map[string]string{
				sourceObjectAnnotation:  sourceSecretName.String(),
				sourceVersionAnnotation: srcSecret.ResourceVersion,
			}
			if err := ctrl.SetControllerReference(&secretReplication, &dstSecret, r.Scheme); err != nil {
				return result, err
			}

			if err := r.Create(ctx, &dstSecret); err != nil {
				log.Error(err, "unable to create replicated secret")
				r.EventRecorder.Event(&secretReplication, corev1.EventTypeWarning, "Unable to create the replicated secret", fmt.Sprintf("%s", err))
				return result, err
			}
			r.EventRecorder.Event(&secretReplication, corev1.EventTypeNormal, "Created replicated secret", "")
		}
	} else if err == nil {
		if srcSecret.ResourceVersion == dstSecret.Annotations[sourceVersionAnnotation] {
			return result, nil
		}
		log.Info("Updating the secret, since it appears to have changed")
		dstSecret.Type = srcSecret.Type
		dstSecret.Data = srcSecret.Data
		dstSecret.Annotations = map[string]string{
			sourceObjectAnnotation:  sourceSecretName.String(),
			sourceVersionAnnotation: srcSecret.ResourceVersion,
		}

		if err := r.Update(ctx, &dstSecret); err != nil {
			log.Error(err, "Unable to update replicated secret")
			r.EventRecorder.Event(&secretReplication, corev1.EventTypeWarning, "Failed to update replicated secret", fmt.Sprintf("%s", err))
			return result, err
		}
		r.EventRecorder.Event(&secretReplication, corev1.EventTypeNormal, "Updated/synced replicated secret", "")
	} else {
		r.EventRecorder.Event(&secretReplication, corev1.EventTypeWarning, "Failed to update or create the replicated secret", fmt.Sprintf("%s", err))
		return result, err
	}

	return result, nil
}

// SetupWithManager sets up the controller
func (r *SecretReplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Secret{}, ownerKey, func(rawObj runtime.Object) []string {
		// grab the secret object, extract the owner
		secret := rawObj.(*corev1.Secret)
		owner := metav1.GetControllerOf(secret)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != "SecretReplication" {
			return nil
		}

		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&replicationv1.SecretReplication{}).
		Owns(&corev1.Secret{}).
		Named(ControllerName).
		Complete(r)
}

func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}
