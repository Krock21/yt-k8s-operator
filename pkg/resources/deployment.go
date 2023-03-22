package resources

import (
	"context"

	"github.com/YTsaurus/yt-k8s-operator/pkg/apiproxy"
	labeller2 "github.com/YTsaurus/yt-k8s-operator/pkg/labeller"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Deployment struct {
	name     string
	labeller *labeller2.Labeller
	apiProxy *apiproxy.APIProxy

	oldObject appsv1.Deployment
	newObject appsv1.Deployment
	built     bool
}

func NewDeployment(
	name string,
	labeller *labeller2.Labeller,
	apiProxy *apiproxy.APIProxy) *Deployment {
	return &Deployment{
		name:     name,
		labeller: labeller,
		apiProxy: apiProxy,
	}
}

func (d *Deployment) OldObject() client.Object {
	return &d.oldObject
}

func (d *Deployment) Name() string {
	return d.name
}

func (d *Deployment) Sync(ctx context.Context) error {
	return d.apiProxy.SyncObject(ctx, &d.oldObject, &d.newObject)
}

func (d *Deployment) Build() *appsv1.Deployment {
	if !d.built {
		d.newObject.ObjectMeta = d.labeller.GetObjectMeta(d.name)
		d.newObject.Spec = appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: d.labeller.GetSelectorLabelMap(),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      d.labeller.GetMetaLabelMap(),
					Annotations: d.apiProxy.Ytsaurus().Spec.ExtraPodAnnotations,
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: d.apiProxy.Ytsaurus().Spec.ImagePullSecrets,
				},
			},
		}
	}

	d.built = true
	return &d.newObject
}

func (d *Deployment) NeedSync(replicas int32, image string) bool {
	if d.oldObject.Spec.Replicas == nil {
		return true
	}

	if *d.oldObject.Spec.Replicas != replicas {
		return true
	}

	if len(d.oldObject.Spec.Template.Spec.Containers) != 1 {
		return true
	}

	if d.oldObject.Spec.Template.Spec.Containers[0].Image != image {
		return true
	}

	return false
}

func (d *Deployment) Fetch(ctx context.Context) error {
	return d.apiProxy.FetchObject(ctx, d.name, &d.oldObject)
}
